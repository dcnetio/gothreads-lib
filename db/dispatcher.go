package db

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/gob"
	"strconv"
	"sync"

	core "github.com/dcnetio/gothreads-lib/core/db"
	kt "github.com/dcnetio/gothreads-lib/db/keytransform"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"golang.org/x/sync/errgroup"
)

var (
	dsDispatcherPrefix = dsPrefix.ChildString("dispatcher")
)

// Reducer applies an event to an existing state.
type Reducer interface {
	Reduce(events []core.Event) error
}

// dispatcher is used to dispatch events to registered reducers.
//
// This is different from generic pub-sub systems because reducers are not subscribed to particular events.
// Every event is dispatched to every registered reducer. When a given reducer is registered, it returns a `token`,
// which can be used to deregister the reducer later.
type dispatcher struct {
	store    kt.TxnDatastoreExtended
	reducers []Reducer
	lock     sync.RWMutex
	lastID   int
}

// NewDispatcher creates a new EventDispatcher.
func newDispatcher(store kt.TxnDatastoreExtended) *dispatcher {
	return &dispatcher{
		store: store,
	}
}

// Store returns the internal event store.
func (d *dispatcher) Store() kt.TxnDatastoreExtended {
	return d.store
}

// Register takes a reducer to be invoked with each dispatched event.
func (d *dispatcher) Register(reducer Reducer) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.lastID++
	d.reducers = append(d.reducers, reducer)
}

// Dispatch dispatches a payload to all registered reducers.
// The logic is separated in two parts:
// 1. Save all txn events with transaction guarantees.
// 2. Notify all reducers about the known events.
func (d *dispatcher) Dispatch(events []core.Event) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	txn, err := d.store.NewTransaction(context.Background(), false)
	if err != nil {
		return err
	}
	defer txn.Discard(context.Background())
	for _, event := range events {
		key, err := getKey(event)
		if err != nil {
			return err
		}
		// Encode and add an Event to event store
		b := bytes.Buffer{}
		e := gob.NewEncoder(&b)
		if err := e.Encode(event); err != nil {
			return err
		}
		if err := txn.Put(context.Background(), key, b.Bytes()); err != nil {
			return err
		}
	}
	if err := txn.Commit(context.Background()); err != nil {
		return err
	}
	// Safe to fire off reducers now that event is persisted
	g, _ := errgroup.WithContext(context.Background())
	for _, reducer := range d.reducers {
		reducer := reducer
		// Launch each reducer in a separate goroutine
		g.Go(func() error {
			return reducer.Reduce(events)
		})
	}
	// Wait for all reducers to complete or error out
	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}

// Query searches the internal event store and returns a query result.
// This is a synchronous version of github.com/ipfs/go-datastore's Query method.
func (d *dispatcher) Query(query query.Query) ([]query.Entry, error) {
	result, err := d.store.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	return result.Rest()
}

// Key format: <timestamp>/<instance-id>/<type>
// @todo: This is up for debate, its a 'fake' Event struct right now anyway
func getKey(event core.Event) (key ds.Key, err error) {
	buf := bytes.NewBuffer(event.Time())
	var unix int64
	if err = binary.Read(buf, binary.BigEndian, &unix); err != nil {
		return
	}
	time := strconv.FormatInt(unix, 10)
	key = dsDispatcherPrefix.ChildString(time).
		ChildString(event.Collection()).
		Instance(event.InstanceID().String())

	return key, nil
}
