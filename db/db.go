// Package db provides a DB which manage collections
package db

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sync"
	"time"

	"github.com/alecthomas/jsonschema"
	threadcbor "github.com/dcnetio/gothreads-lib/cbor"
	"github.com/dcnetio/gothreads-lib/core/app"
	core "github.com/dcnetio/gothreads-lib/core/db"
	lstore "github.com/dcnetio/gothreads-lib/core/logstore"
	"github.com/dcnetio/gothreads-lib/core/net"
	"github.com/dcnetio/gothreads-lib/core/thread"
	kt "github.com/dcnetio/gothreads-lib/db/keytransform"
	"github.com/dcnetio/gothreads-lib/util"
	"github.com/dop251/goja"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	format "github.com/ipfs/go-ipld-format"
	logging "github.com/ipfs/go-log/v2"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	idFieldName                 = "_id"
	modFieldName                = "_mod"
	getBlockRetries             = 3
	getBlockInitialTimeout      = time.Millisecond * 500
	pullThreadBackgroundTimeout = time.Hour
	createNetRecordTimeout      = time.Second * 15
)

var (
	log = logging.Logger("db")

	// ErrThreadReadKeyRequired indicates the provided thread key does not contain a read key.
	ErrThreadReadKeyRequired = errors.New("thread read key is required")
	// ErrInvalidName indicates the provided name isn't valid for a Collection.
	ErrInvalidName = errors.New("name may only contain alphanumeric characters or non-consecutive hyphens, " +
		"and cannot begin or end with a hyphen")
	// ErrInvalidCollectionSchema indicates the provided schema isn't valid for a Collection.
	ErrInvalidCollectionSchema = errors.New("the collection schema _id property must be a string")
	// ErrCannotIndexIDField indicates a custom index was specified on the ID field.
	ErrCannotIndexIDField = errors.New("cannot create custom index on " + idFieldName)

	nameRx *regexp.Regexp

	dsPrefix     = ds.NewKey("/db")
	dsName       = dsPrefix.ChildString("name")
	dsSchemas    = dsPrefix.ChildString("schema")
	dsIndexes    = dsPrefix.ChildString("index")
	dsValidators = dsPrefix.ChildString("validator")
	dsFilters    = dsPrefix.ChildString("filter")
)

func init() {
	nameRx = regexp.MustCompile(`^[A-Za-z0-9]+(?:[-][A-Za-z0-9]+)*$`)
	// register empty map in order to gob-encode old-format events with non-encodable time.Time,
	// which cbor decodes as map[string]interface{}
	gob.Register(map[string]interface{}{})
}

// DB is the aggregate-root of events and state. External/remote events
// are dispatched to the DB, and are internally processed to impact collection
// states. Likewise, local changes in collections registered produce events dispatched
// externally.
type DB struct {
	io.Closer

	name      string
	connector *app.Connector

	datastore  kt.TxnDatastoreExtended
	dispatcher *dispatcher
	eventcodec core.EventCodec

	lock        sync.RWMutex
	txnlock     sync.RWMutex
	collections map[string]*Collection
	closed      bool

	localEventsBus      *app.LocalEventsBus
	stateChangedNotifee *stateChangedNotifee
}

// NewDB creates a new DB, which will *own* ds and dispatcher for internal use.
// Saying it differently, ds and dispatcher shouldn't be used externally.
func NewDB(
	ctx context.Context,
	store kt.TxnDatastoreExtended,
	network app.Net,
	id thread.ID,
	opts ...NewOption,
) (*DB, error) {
	log.Debugf("creating db %s", id)
	args := &NewOptions{}
	for _, opt := range opts {
		opt(args)
	}
	if err := util.SetLogLevels(map[string]logging.LogLevel{
		"db": util.LevelFromDebugFlag(args.Debug),
	}); err != nil {
		return nil, err
	}

	if args.Key.Defined() && !args.Key.CanRead() {
		return nil, ErrThreadReadKeyRequired
	}
	if _, err := network.CreateThread(
		ctx,
		id,
		net.WithThreadKey(args.Key),
		net.WithLogKey(args.LogKey),
		net.WithNewThreadToken(args.Token),
	); err != nil && !errors.Is(err, lstore.ErrThreadExists) && !errors.Is(err, lstore.ErrLogExists) {
		return nil, err
	}
	return newDB(store, network, id, args)
}

// NewDBFromAddr creates a new DB from a thread hosted by another peer at address,
// which will *own* ds and dispatcher for internal use.
// Saying it differently, ds and dispatcher shouldn't be used externally.
func NewDBFromAddr(
	ctx context.Context,
	store kt.TxnDatastoreExtended,
	network app.Net,
	addr ma.Multiaddr,
	key thread.Key,
	opts ...NewOption,
) (*DB, error) {
	log.Debugf("creating db from address %s", addr)
	args := &NewOptions{}
	for _, opt := range opts {
		opt(args)
	}
	if err := util.SetLogLevels(map[string]logging.LogLevel{
		"db": util.LevelFromDebugFlag(args.Debug),
	}); err != nil {
		return nil, err
	}

	if key.Defined() && !key.CanRead() {
		return nil, ErrThreadReadKeyRequired
	}
	info, err := network.AddThread(
		ctx,
		addr,
		net.WithThreadKey(key),
		net.WithLogKey(args.LogKey),
		net.WithNewThreadToken(args.Token),
	)
	if err != nil {
		return nil, err
	}
	d, err := newDB(store, network, info.ID, args)
	if err != nil {
		return nil, err
	}

	if args.Block {
		if err = network.PullThread(ctx, info.ID, net.WithThreadToken(args.Token)); err != nil {
			return nil, err
		}
	} else {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), pullThreadBackgroundTimeout)
			defer cancel()
			if err := network.PullThread(ctx, info.ID, net.WithThreadToken(args.Token)); err != nil {
				log.Errorf("error pulling thread %s", info.ID)
			}
		}()
	}
	return d, nil
}

// newDB is used directly by a db manager to create new dbs with the same config.
func newDB(s kt.TxnDatastoreExtended, n app.Net, id thread.ID, opts *NewOptions) (*DB, error) {
	if opts.EventCodec == nil {
		opts.EventCodec = newDefaultEventCodec()
	}

	d := &DB{
		datastore:           s,
		dispatcher:          newDispatcher(s),
		eventcodec:          opts.EventCodec,
		collections:         make(map[string]*Collection),
		localEventsBus:      app.NewLocalEventsBus(),
		stateChangedNotifee: &stateChangedNotifee{},
	}
	if err := d.loadName(); err != nil {
		return nil, err
	}
	prevName := d.name
	if opts.Name != "" {
		d.name = opts.Name
	} else if prevName == "" {
		d.name = "unnamed"
	}
	if err := d.saveName(prevName); err != nil {
		return nil, err
	}
	if err := d.reCreateCollections(); err != nil {
		return nil, err
	}
	d.dispatcher.Register(d)

	connector, err := n.ConnectApp(d, id)
	if err != nil {
		return nil, err
	}
	d.connector = connector

	for _, cc := range opts.Collections {
		if _, err := d.NewCollection(cc); err != nil {
			return nil, err
		}
	}
	return d, nil
}

// saveName saves the db name.
func (d *DB) saveName(prevName string) error {
	if d.name == prevName {
		return nil
	}
	if !nameRx.MatchString(d.name) {
		return ErrInvalidName
	}
	return d.datastore.Put(context.Background(), dsName, []byte(d.name))
}

// loadName loads db name if present.
func (d *DB) loadName() error {
	bytes, err := d.datastore.Get(context.Background(), dsName)
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}
	if bytes != nil {
		d.name = string(bytes)
	}
	return nil
}

// reCreateCollections loads and registers schemas from the datastore.
func (d *DB) reCreateCollections() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	log.Debugf("recreating collections in %s", d.name)

	results, err := d.datastore.Query(context.Background(), query.Query{
		Prefix: dsSchemas.String(),
	})
	if err != nil {
		return err
	}
	defer results.Close()
	for res := range results.Next() {
		name := ds.RawKey(res.Key).Name()
		schema := &jsonschema.Schema{}
		if err := json.Unmarshal(res.Value, schema); err != nil {
			return err
		}
		wv, err := d.datastore.Get(context.Background(), dsValidators.ChildString(name))
		if err != nil && !errors.Is(err, ds.ErrNotFound) {
			return err
		}
		rf, err := d.datastore.Get(context.Background(), dsFilters.ChildString(name))
		if err != nil && !errors.Is(err, ds.ErrNotFound) {
			return err
		}
		c, err := newCollection(d, CollectionConfig{
			Name:           name,
			Schema:         schema,
			WriteValidator: string(wv),
			ReadFilter:     string(rf),
		})
		if err != nil {
			return err
		}
		var indexes map[string]Index
		index, err := d.datastore.Get(context.Background(), dsIndexes.ChildString(name))
		if err == nil && index != nil {
			if err := json.Unmarshal(index, &indexes); err != nil {
				return err
			}
		}
		for _, index := range indexes {
			if index.Path != "" { // Catch bad indexes from an old bug
				c.indexes[index.Path] = index
			}
		}
		d.collections[c.name] = c
	}
	return nil
}

// Info wraps info about a db.
type Info struct {
	Name  string
	Addrs []ma.Multiaddr
	Key   thread.Key
}

// GetDBInfo returns the addresses and key that can be used to join the DB thread.
func (d *DB) GetDBInfo(opts ...Option) (info Info, err error) {
	log.Debugf("getting db info in %s", d.name)
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	thrd, err := d.connector.Net.GetThread(
		context.Background(),
		d.connector.ThreadID(),
		net.WithThreadToken(options.Token),
	)
	if err != nil {
		return info, err
	}
	return Info{
		Name:  d.name,
		Addrs: thrd.Addrs,
		Key:   thrd.Key,
	}, nil
}

// CollectionConfig describes a new Collection.
type CollectionConfig struct {
	// Name is the name of the collection.
	// Must only contain alphanumeric characters or non-consecutive hyphens, and cannot begin or end with a hyphen.
	Name string
	// Schema is JSON Schema used for instance validation.
	Schema *jsonschema.Schema
	// Indexes is a list of index configurations, which define how instances are indexed.
	Indexes []Index
	// An optional JavaScript (ECMAScript 5.1) function that is used to validate instances on write.
	// The function receives three arguments:
	//   - writer: The multibase-encoded public key identity of the writer.
	//   - event: An object describing the update event (see core.Event).
	//   - instance: The current instance as a JavaScript object before the update event is applied.
	// A "falsy" return value indicates a failed validation (https://developer.mozilla.org/en-US/docs/Glossary/Falsy).
	// Note: Only the function body should be defined here.
	WriteValidator string
	// An optional JavaScript (ECMAScript 5.1) function that is used to filter instances on read.
	// The function receives two arguments:
	//   - reader: The multibase-encoded public key identity of the reader.
	//   - instance: The current instance as a JavaScript object.
	// The function must return a JavaScript object.
	// Most implementation will modify and return the current instance.
	// Note: Only the function body should be defined here.
	ReadFilter string
}

// NewCollection creates a new db collection with config.
func (d *DB) NewCollection(config CollectionConfig, opts ...Option) (*Collection, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	log.Debugf("creating collection %s in %s", config.Name, d.name)
	args := &Options{}
	for _, opt := range opts {
		opt(args)
	}
	if err := d.connector.Validate(args.Token, false); err != nil {
		return nil, err
	}
	if _, ok := d.collections[config.Name]; ok {
		return nil, ErrCollectionAlreadyRegistered
	}
	c, err := newCollection(d, config)
	if err != nil {
		return nil, err
	}
	if err := d.addIndexes(c, config.Schema, config.Indexes, opts...); err != nil {
		return nil, err
	}
	if err := d.saveCollection(c); err != nil {
		return nil, err
	}
	return c, nil
}

// UpdateCollection updates an existing db collection with a new config.
// Indexes to new paths will be created.
// Indexes to removed paths will be dropped.
func (d *DB) UpdateCollection(config CollectionConfig, opts ...Option) (*Collection, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	log.Debugf("updating collection %s in %s", config.Name, d.name)
	args := &Options{}
	for _, opt := range opts {
		opt(args)
	}
	if err := d.connector.Validate(args.Token, false); err != nil {
		return nil, err
	}
	xc, ok := d.collections[config.Name]
	if !ok {
		return nil, ErrCollectionNotFound
	}
	c, err := newCollection(d, config)
	if err != nil {
		return nil, err
	}
	if err := d.addIndexes(c, config.Schema, config.Indexes, opts...); err != nil {
		return nil, err
	}

	// Drop indexes that are no longer requested
	for _, index := range xc.indexes {
		if _, ok := c.indexes[index.Path]; !ok {
			if err := c.dropIndex(index.Path); err != nil {
				return nil, err
			}
		}
	}

	if err := d.saveCollection(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (d *DB) addIndexes(c *Collection, schema *jsonschema.Schema, indexes []Index, opts ...Option) error {
	log.Debugf("adding indexes to collection %s in %s", c.name, d.name)
	for _, index := range indexes {
		if index.Path == idFieldName {
			return ErrCannotIndexIDField
		}
		if err := c.addIndex(schema, index, opts...); err != nil {
			return err
		}
	}
	return c.addIndex(schema, Index{Path: idFieldName, Unique: true}, opts...)
}

func (d *DB) saveCollection(c *Collection) error {
	log.Debugf("saving collection %s in %s", c.name, d.name)
	if err := d.datastore.Put(context.Background(), dsSchemas.ChildString(c.name), c.GetSchema()); err != nil {
		return err
	}
	if c.rawWriteValidator != nil {
		if err := d.datastore.Put(context.Background(), dsValidators.ChildString(c.name), c.rawWriteValidator); err != nil {
			return err
		}
	}
	if c.rawReadFilter != nil {
		if err := d.datastore.Put(context.Background(), dsFilters.ChildString(c.name), c.rawReadFilter); err != nil {
			return err
		}
	}
	d.collections[c.name] = c
	return nil
}

// GetCollection returns a collection by name.
func (d *DB) GetCollection(name string, opts ...Option) *Collection {
	d.lock.Lock()
	defer d.lock.Unlock()
	log.Debugf("getting collection %s in %s", name, d.name)
	args := &Options{}
	for _, opt := range opts {
		opt(args)
	}
	if err := d.connector.Validate(args.Token, true); err != nil {
		return nil
	}
	return d.collections[name]
}

// ListCollections returns all collections.
func (d *DB) ListCollections(opts ...Option) []*Collection {
	d.lock.Lock()
	defer d.lock.Unlock()
	log.Debugf("listing collections in %s", d.name)
	args := &Options{}
	for _, opt := range opts {
		opt(args)
	}
	if err := d.connector.Validate(args.Token, true); err != nil {
		return nil
	}
	list := make([]*Collection, len(d.collections))
	var i int
	for _, c := range d.collections {
		list[i] = c
		i++
	}
	return list
}

// DeleteCollection deletes collection by name and drops all indexes.
func (d *DB) DeleteCollection(name string, opts ...Option) error {
	ctx := context.Background()
	d.lock.Lock()
	defer d.lock.Unlock()
	log.Debugf("deleting collection %s in %s", name, d.name)
	args := &Options{}
	for _, opt := range opts {
		opt(args)
	}
	if err := d.connector.Validate(args.Token, false); err != nil {
		return err
	}
	c, ok := d.collections[name]
	if !ok {
		return ErrCollectionNotFound
	}
	txn, err := d.datastore.NewTransaction(ctx, false)
	if err != nil {
		return err
	}
	if err := txn.Delete(ctx, dsIndexes.ChildString(c.name)); err != nil {
		return err
	}
	if err := txn.Delete(ctx, dsSchemas.ChildString(c.name)); err != nil {
		return err
	}
	if err := txn.Delete(ctx, dsValidators.ChildString(c.name)); err != nil {
		return err
	}
	if err := txn.Delete(ctx, dsFilters.ChildString(c.name)); err != nil {
		return err
	}
	if err := txn.Commit(ctx); err != nil {
		return err
	}
	delete(d.collections, c.name)
	return nil
}

func (d *DB) Close() error {
	log.Debugf("closing %s", d.name)
	d.lock.Lock()
	defer d.lock.Unlock()
	d.txnlock.Lock()
	defer d.txnlock.Unlock()

	if d.closed {
		return nil
	}
	d.closed = true
	d.localEventsBus.Discard()
	d.stateChangedNotifee.close()
	return nil
}

func (d *DB) Reduce(events []core.Event) error {
	log.Debugf("reducing events in %s", d.name)
	codecActions, err := d.eventcodec.Reduce(events, d.datastore, baseKey, defaultIndexFunc(d))
	if err != nil {
		return err
	}
	actions := make([]Action, len(codecActions))
	for i, ca := range codecActions {
		var actionType ActionType
		switch codecActions[i].Type {
		case core.Create:
			actionType = ActionCreate
		case core.Save:
			actionType = ActionSave
		case core.Delete:
			actionType = ActionDelete
		default:
			panic("eventcodec action not recognized")
		}
		actions[i] = Action{Collection: ca.Collection, Type: actionType, ID: ca.InstanceID}
	}
	d.notifyStateChanged(actions)
	return nil
}

func defaultIndexFunc(d *DB) func(collection string, key ds.Key, oldData, newData []byte, txn ds.Txn) error {
	return func(collection string, key ds.Key, oldData, newData []byte, txn ds.Txn) error {
		c := d.GetCollection(collection)
		if c == nil {
			return fmt.Errorf("collection (%s) not found", collection)
		}
		if err := c.indexDelete(txn, key, oldData); err != nil {
			return err
		}
		if newData == nil {
			return nil
		}
		return c.indexAdd(txn, key, newData)
	}
}
func (d *DB) RebuildIndex(collection string, field string, uniqueFlag bool) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	log.Debugf("rebuilding index %s in collection %s in %s", field, collection, d.name)
	//遍历collection所有记录
	c := d.collections[collection]
	if c == nil {
		return ErrCollectionNotFound
	}
	txn, err := d.datastore.NewTransactionExtended(context.Background(), false)
	if err != nil {
		return fmt.Errorf("error building internal query: %v", err)
	}
	defer txn.Discard(context.Background())
	q := &Query{}
	iter, err := newIterator(txn, c.baseKey(), q)
	if err != nil {
		return err
	}
	defer iter.Close()
	index := Index{Path: field, Unique: uniqueFlag}
	for {
		res, ok := iter.NextSync()
		if !ok {
			break
		}
		key := ds.NewKey(res.Key)
		value := res.Value
		if err = c.indexUpdate(index.Path, index, txn, key, value, true); err != nil {
			break
		}
		if err = c.indexUpdate(index.Path, index, txn, key, value, false); err != nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("error rebuilding index %s in collection %s in %s: %v", field, collection, d.name, err)
	}
	err = txn.Commit(context.Background())
	if err != nil {
		return fmt.Errorf("error committing index rebuild transaction: %v", err)
	}
	return nil
}

func (d *DB) ValidateNetRecordBody(_ context.Context, body format.Node, identity thread.PubKey) error {
	events, err := d.eventcodec.EventsFromBytes(body.RawData())
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}
	for _, e := range events {
		c, ok := d.collections[e.Collection()]
		if !ok {
			return ErrCollectionNotFound
		}
		if err := c.validWrite(identity, e); err != nil {
			return err
		}
	}
	return nil
}

func parseJSON(vm *goja.Runtime, val []byte) (goja.Value, error) {
	return vm.RunString(fmt.Sprintf("JSON.parse(`%s`);", string(val)))
}

func (d *DB) HandleNetRecord(ctx context.Context, rec net.ThreadRecord, key thread.Key) error {
	event, err := threadcbor.EventFromRecord(ctx, d.connector.Net, rec.Value())
	if err != nil {
		block, err := d.getBlockWithRetry(ctx, rec.Value())
		if err != nil {
			return fmt.Errorf("error when getting block from record: %v", err)
		}
		event, err = threadcbor.EventFromNode(block)
		if err != nil {
			return fmt.Errorf("error when decoding block to event: %v", err)
		}
	}
	body, err := event.GetBody(ctx, d.connector.Net, key.Read())
	if err != nil {
		return fmt.Errorf("error when getting body of event on thread %s/%s: %v", d.connector.ThreadID(), rec.LogID(), err)
	}
	events, err := d.eventcodec.EventsFromBytes(body.RawData())
	if err != nil {
		return fmt.Errorf("error when unmarshaling event from bytes: %v", err)
	}
	log.Debugf("dispatching new record: %s/%s", rec.ThreadID(), rec.LogID())
	return d.dispatch(events)
}

func (d *DB) GetNetRecordCreateTime(ctx context.Context, rec net.ThreadRecord, key thread.Key) (int64, error) {
	event, err := threadcbor.EventFromRecord(ctx, d.connector.Net, rec.Value())
	if err != nil {
		block, err := d.getBlockWithRetry(ctx, rec.Value())
		if err != nil {
			return 0, fmt.Errorf("error when getting block from record: %v", err)
		}
		event, err = threadcbor.EventFromNode(block)
		if err != nil {
			return 0, fmt.Errorf("error when decoding block to event: %v", err)
		}
	}
	body, err := event.GetBody(ctx, d.connector.Net, key.Read())
	if err != nil {
		return 0, fmt.Errorf("error when getting body of event on thread %s/%s: %v", d.connector.ThreadID(), rec.LogID(), err)
	}
	events, err := d.eventcodec.EventsFromBytes(body.RawData())
	if err != nil {
		return 0, fmt.Errorf("error when unmarshaling event from bytes: %v", err)
	}
	if len(events) == 0 {
		return 0, fmt.Errorf("no events found in record")
	}
	core_event := events[0]
	buf := bytes.NewBuffer(core_event.Time())
	var unix int64
	if err = binary.Read(buf, binary.BigEndian, &unix); err != nil {
		return 0, err
	}
	return unix, nil
}

// getBlockWithRetry gets a record block with exponential backoff.
func (d *DB) getBlockWithRetry(ctx context.Context, rec net.Record) (format.Node, error) {
	backoff := getBlockInitialTimeout
	var err error
	for i := 1; i <= getBlockRetries; i++ {
		n, err := rec.GetBlock(ctx, d.connector.Net)
		if err == nil {
			return n, nil
		}
		log.Warnf("error when fetching block %s in retry %d", rec.Cid(), i)
		time.Sleep(backoff)
		backoff *= 2
	}
	return nil, err
}

// dispatch applies external events to the db. This function guarantee
// no interference with registered collection states, and viceversa.
func (d *DB) dispatch(events []core.Event) error {
	log.Debugf("dispatching events in %s", d.name)
	d.txnlock.Lock()
	defer d.txnlock.Unlock()
	if err := d.dispatcher.Dispatch(events); err != nil {
		return err
	}
	log.Debugf("dispatched events in %s", d.name)
	return nil
}

func (d *DB) readTxn(c *Collection, f func(txn *Txn) error, opts ...TxnOption) error {
	log.Debugf("starting read txn in %s", d.name)
	d.txnlock.RLock()
	defer d.txnlock.RUnlock()

	args := &TxnOptions{}
	for _, opt := range opts {
		opt(args)
	}
	txn := &Txn{collection: c, token: args.Token, readonly: true}
	defer txn.Discard()
	if err := f(txn); err != nil {
		return err
	}
	log.Debugf("ending read txn in %s", d.name)
	return nil
}

func (d *DB) writeTxn(c *Collection, f func(txn *Txn) error, opts ...TxnOption) error {
	log.Debugf("starting write txn in %s", d.name)
	d.txnlock.Lock()
	defer d.txnlock.Unlock()

	args := &TxnOptions{}
	for _, opt := range opts {
		opt(args)
	}
	txn := &Txn{collection: c, token: args.Token}
	defer txn.Discard()
	if err := f(txn); err != nil {
		return err
	}
	if err := txn.Commit(); err != nil {
		return err
	}
	log.Debugf("ending write txn in %s", d.name)
	return nil
}
