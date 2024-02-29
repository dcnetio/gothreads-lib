package db

import (
	"context"
	"errors"
	"sync"
	"time"

	core "github.com/dcnetio/gothreads-lib/core/db"
	dse "github.com/dcnetio/gothreads-lib/go-datastore-extensions"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
)

type TxnMapDatastore struct {
	*ds.MapDatastore
	lock sync.RWMutex
}

var _ dse.DatastoreExtensions = (*TxnMapDatastore)(nil)

func NewTxMapDatastore() *TxnMapDatastore {
	return &TxnMapDatastore{
		MapDatastore: ds.NewMapDatastore(),
	}
}

func (d *TxnMapDatastore) NewTransaction(ctx context.Context, _ bool) (ds.Txn, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return NewSimpleTx(d), nil
}

func (d *TxnMapDatastore) NewTransactionExtended(ctx context.Context, _ bool) (dse.TxnExt, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return NewSimpleTx(d), nil
}

func (d *TxnMapDatastore) QueryExtended(ctx context.Context, q dse.QueryExt) (query.Results, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.Query(ctx, q.Query)
}

type op struct {
	delete bool
	value  []byte
}

// SimpleTx implements the transaction interface for datastores who do
// not have any sort of underlying transactional support.
type SimpleTx struct {
	ops    map[ds.Key]op
	lock   sync.RWMutex
	target ds.Datastore
}

func NewSimpleTx(store ds.Datastore) dse.TxnExt {
	return &SimpleTx{
		ops:    make(map[ds.Key]op),
		target: store,
	}
}

func (bt *SimpleTx) Query(ctx context.Context, q query.Query) (query.Results, error) {
	bt.lock.RLock()
	defer bt.lock.RUnlock()
	return bt.target.Query(ctx, q)
}

func (bt *SimpleTx) QueryExtended(ctx context.Context, q dse.QueryExt) (query.Results, error) {
	bt.lock.RLock()
	defer bt.lock.RUnlock()
	return bt.target.Query(ctx, q.Query)
}

func (bt *SimpleTx) Get(ctx context.Context, k ds.Key) ([]byte, error) {
	bt.lock.RLock()
	defer bt.lock.RUnlock()
	return bt.target.Get(ctx, k)
}

func (bt *SimpleTx) Has(ctx context.Context, k ds.Key) (bool, error) {
	bt.lock.RLock()
	defer bt.lock.RUnlock()
	return bt.target.Has(ctx, k)
}

func (bt *SimpleTx) GetSize(ctx context.Context, k ds.Key) (int, error) {
	bt.lock.RLock()
	defer bt.lock.RUnlock()
	return bt.target.GetSize(ctx, k)
}

func (bt *SimpleTx) Put(ctx context.Context, key ds.Key, val []byte) error {
	bt.lock.RLock()
	defer bt.lock.RUnlock()
	bt.ops[key] = op{value: val}
	return nil
}

func (bt *SimpleTx) Delete(ctx context.Context, key ds.Key) error {
	bt.lock.RLock()
	defer bt.lock.RUnlock()
	bt.ops[key] = op{delete: true}
	return nil
}

func (bt *SimpleTx) Discard(ctx context.Context) {
	bt.lock.RLock()
	defer bt.lock.RUnlock()
}

func (bt *SimpleTx) Commit(ctx context.Context) error {
	bt.lock.RLock()
	defer bt.lock.RUnlock()
	var err error
	for k, op := range bt.ops {
		if op.delete {
			err = bt.target.Delete(ctx, k)
		} else {
			err = bt.target.Put(ctx, k, op.value)
		}
		if err != nil {
			break
		}
	}

	return err
}

type nullReducer struct{}

var _ Reducer = (*nullReducer)(nil)

func (n *nullReducer) Reduce(_ []core.Event) error {
	return nil
}

type errorReducer struct{}

var _ Reducer = (*errorReducer)(nil)

func (n *errorReducer) Reduce(_ []core.Event) error {
	return errors.New("error")
}

type slowReducer struct{}

var _ Reducer = (*slowReducer)(nil)

func (n *slowReducer) Reduce(_ []core.Event) error {
	time.Sleep(2 * time.Second)
	return nil
}
