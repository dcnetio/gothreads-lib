// Copyright 2019 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package db

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"sort"

	"github.com/alecthomas/jsonschema"
	dse "github.com/dcnetio/gothreads-lib/go-datastore-extensions"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const (
	// iteratorKeyMinCacheSize is the size of iterator keys stored in memory before more are fetched.
	iteratorKeyMinCacheSize = 100
)

var (
	// ErrUniqueExists indicates an insert resulted in a unique constraint violation.
	ErrUniqueExists = errors.New("unique constraint violation")
	// ErrNotIndexable indicates an index path does not resolve to a value.
	ErrNotIndexable = errors.New("value not indexable")
	// ErrCantCreateUniqueIndex indicates a unique index can't be created because multiple instances share a value at path.
	ErrCantCreateUniqueIndex = errors.New("can't create unique index (duplicate instances exist)")
	// ErrIndexNotFound indicates a requested index was not found.
	ErrIndexNotFound = errors.New("index not found")

	indexPrefix = ds.NewKey("_index")
	indexTypes  = []string{"string", "number", "integer", "boolean"}
)

// Index defines an index.
type Index struct {
	// Path to the field to index in dot syntax, e.g., "name.last" or "age".
	Path string `json:"path"`
	// Unique indicates that only one instance should exist per field value.
	Unique bool `json:"unique,omitempty"`
}

// GetIndexes returns the current indexes.
func (c *Collection) GetIndexes() []Index {
	if len(c.indexes) == 0 {
		return nil
	}
	indexes := make([]Index, len(c.indexes)-1)
	var i int
	for _, index := range c.indexes {
		if index.Path != idFieldName {
			indexes[i] = index
			i++
		}
	}
	return indexes
}

// addIndex creates a new index based on path.
// Use dot syntax to reach nested fields, e.g., "name.last".
// The field at path must be one of the supported JSON Schema types: string, number, integer, or boolean
// Set unique to true if you want a unique constraint on path.
// Adding an index will override any overlapping index values if they already exist.
// @note: This does NOT currently build the index. If items have been added prior to adding
// a new index, they will NOT be indexed a posteriori.
func (c *Collection) addIndex(schema *jsonschema.Schema, index Index, opts ...Option) error {
	args := &Options{}
	for _, opt := range opts {
		opt(args)
	}

	// Don't allow the default index to be overwritten
	if index.Path == idFieldName {
		if _, ok := c.indexes[idFieldName]; ok {
			return nil
		}
	}

	// Validate path and type.
	jt, err := getSchemaTypeAtPath(schema, index.Path)
	if err != nil {
		return err
	}
	var valid bool
	for _, t := range indexTypes {
		if jt.Type == t {
			valid = true
			break
		}
	}
	if !valid {
		return ErrNotIndexable
	}

	// Skip if nothing to do
	if x, ok := c.indexes[index.Path]; ok && index.Unique == x.Unique {
		return nil
	}

	// Ensure collection does not contain multiple instances with the same value at path
	if index.Unique && index.Path != idFieldName {
		vals := make(map[interface{}]struct{})
		all, err := c.Find(&Query{}, WithTxnToken(args.Token))
		if err != nil {
			return err
		}
		for _, i := range all {
			res := gjson.GetBytes(i, index.Path)
			if !res.Exists() {
				continue
			}
			if _, ok := vals[res.Value()]; ok {
				return ErrCantCreateUniqueIndex
			} else {
				vals[res.Value()] = struct{}{}
			}
		}
	}

	c.indexes[index.Path] = index
	return c.saveIndexes()
}

// dropIndex drops the index at path.
func (c *Collection) dropIndex(pth string) error {
	// Don't allow the default index to be dropped
	if pth == idFieldName {
		return errors.New(idFieldName + " index cannot be dropped")
	}
	delete(c.indexes, pth)
	return c.saveIndexes()
}

// saveIndexes persists the current indexes.
func (c *Collection) saveIndexes() error {
	ib, err := json.Marshal(c.indexes)
	if err != nil {
		return err
	}
	return c.db.datastore.Put(context.Background(), dsIndexes.ChildString(c.name), ib)
}

// indexAdd adds an item to the index.
func (c *Collection) indexAdd(tx ds.Txn, key ds.Key, data []byte) error {
	for path, index := range c.indexes {
		err := c.indexUpdate(path, index, tx, key, data, false)
		if err != nil {
			return err
		}
	}
	return nil
}

// indexDelete removes an item from the index.
// Be sure to pass the data from the old record, not the new one.
func (c *Collection) indexDelete(tx ds.Txn, key ds.Key, originalData []byte) error {
	for path, index := range c.indexes {
		err := c.indexUpdate(path, index, tx, key, originalData, true)
		if err != nil {
			return err
		}
	}
	return nil
}

// indexUpdate adds or removes a specific index on an item.
func (c *Collection) indexUpdate(field string, index Index, tx ds.Txn, key ds.Key, input []byte, delete bool) error {
	valueKey, err := getIndexValue(field, input)
	if err != nil {
		if errors.Is(err, ErrNotIndexable) {
			return nil
		}
		return err
	}

	indexKey := indexPrefix.Child(c.baseKey()).ChildString(field).ChildString(valueKey.String()[1:])
	data, err := tx.Get(context.Background(), indexKey)
	if err != nil && err != ds.ErrNotFound {
		return err
	}
	indexValue := make(keyList, 0)
	if data != nil {
		err = DefaultDecode(data, &indexValue)
		if err != nil {
			return err
		}
	}
	if delete {
		indexValue.remove(key)
	} else {
		if index.Unique {
			indexValue.remove(key)
		}
		indexValue.add(key)
	}
	if len(indexValue) == 0 {
		return tx.Delete(context.Background(), indexKey)
	}
	val, err := DefaultEncode(indexValue)
	if err != nil {
		return err
	}
	return tx.Put(context.Background(), indexKey, val)
}

// getIndexValue returns the result of a field search on input.
func getIndexValue(field string, input []byte) (ds.Key, error) {
	result := gjson.GetBytes(input, field)
	if !result.Exists() {
		return ds.Key{}, ErrNotIndexable
	}
	return ds.NewKey(result.String()), nil
}

// keyList is a slice of unique, sorted keys([]byte) such as what an index points to
type keyList [][]byte

func (v *keyList) add(key ds.Key) {
	b := key.Bytes()
	i := sort.Search(len(*v), func(i int) bool {
		return bytes.Compare((*v)[i], b) >= 0
	})

	if i < len(*v) && bytes.Equal((*v)[i], b) {
		// already added
		return
	}

	*v = append(*v, nil)
	copy((*v)[i+1:], (*v)[i:])
	(*v)[i] = b
}

func (v *keyList) remove(key ds.Key) {
	b := key.Bytes()
	i := sort.Search(len(*v), func(i int) bool {
		return bytes.Compare((*v)[i], b) >= 0
	})

	if i < len(*v) {
		copy((*v)[i:], (*v)[i+1:])
		(*v)[len(*v)-1] = nil
		*v = (*v)[:len(*v)-1]
	}
}

func (v *keyList) In(key ds.Key) bool {
	b := key.Bytes()
	i := sort.Search(len(*v), func(i int) bool {
		return bytes.Compare((*v)[i], b) >= 0
	})
	return i < len(*v) && bytes.Equal((*v)[i], b)
}

type MarshaledResult struct {
	query.Result
	MarshaledValue map[string]interface{}
}

type iterator struct {
	nextKeys func() ([]ds.Key, error)
	txn      ds.Txn
	query    *Query
	keyCache []ds.Key
	iter     query.Results
}

func newIterator(txn dse.TxnExt, baseKey ds.Key, q *Query) (*iterator, error) {
	i := &iterator{
		txn:   txn,
		query: q,
	}
	var prefix ds.Key
	if q.Index == "" {
		prefix = baseKey
	} else {
		prefix = indexPrefix.Child(baseKey).ChildString(q.Index)
	}

	dsq := dse.QueryExt{
		Query: query.Query{
			Prefix: prefix.String(),
			// Pull out Skip and Limit here because we won't know ahead of time
			// (due to readFilters) how many to skip/limit
			// Limit:  q.Limit,
			// Offset: q.Skip,
		},
	}
	if q.Sort.FieldPath == idFieldName {
		if q.Sort.Desc {
			dsq.Orders = []query.Order{query.OrderByKeyDescending{}}
		} else {
			dsq.Orders = []query.Order{query.OrderByKey{}}
		}
	}
	if q.Seek != "" {
		dsq.SeekPrefix = prefix.Child(ds.NewKey(string(q.Seek))).String()
	}
	iter, err := txn.QueryExtended(context.Background(), dsq)
	if err != nil {
		return nil, err
	}
	i.iter = iter

	// Key field or index not specified, pass thru to base 'iterator'
	if q.Index == "" {
		i.nextKeys = func() ([]ds.Key, error) {
			return nil, nil
		}
		return i, nil
	}

	// indexed field, get keys from index
	first := true
	i.nextKeys = func() ([]ds.Key, error) {
		var nKeys []ds.Key
		for len(nKeys) < iteratorKeyMinCacheSize {
			result, ok := i.iter.NextSync()
			if !ok {
				if first {
					return nil, ErrIndexNotFound
				}
				return nKeys, result.Error
			}
			first = false
			// result.Key contains the indexed value, extract here first
			key := ds.RawKey(result.Key)
			base := prefix.Name()
			name := key.Name()
			var val interface{}
			if gjson.Valid(name) {
				val = gjson.Parse(name).Value()
				if val == nil {
					val = name
				}
			} else {
				val = name
			}

			doc, err := sjson.Set("", base, val)
			if err != nil {
				return nil, err
			}
			value := make(map[string]interface{})
			if err := json.Unmarshal([]byte(doc), &value); err != nil { //In order for the business logic to proceed normally, an error will not be reported here, but will be skipped.
				continue
				//return nil, fmt.Errorf("error when unmarshaling query result: %v", err)
			}
			ok, err = q.match(value)
			if err != nil { //The matching condition is an error, which may be caused by data errors. In order for the business logic to proceed normally, an error is not reported here, but skipped.
				continue
				//return nil, fmt.Errorf("error when matching entry with query: %v", err)
			}
			if ok {
				indexValue := make(keyList, 0)
				if err := DefaultDecode(result.Value, &indexValue); err != nil {
					return nil, err
				}
				for _, v := range indexValue {
					nKeys = append(nKeys, ds.RawKey(string(v)))
				}
			}
		}
		return nKeys, nil
	}
	return i, nil
}

// NextSync returns the next key value that matches the iterators criteria
// If there is an error, ok is false and result.Error() will return the error
func (i *iterator) NextSync() (MarshaledResult, bool) {
	if i.query.Index == "" {
		value := MarshaledResult{}
		var ok bool
		for res := range i.iter.Next() {
			val := make(map[string]interface{})
			if value.Error = json.Unmarshal(res.Value, &val); value.Error != nil {
				break
			}
			ok, value.Error = i.query.match(val)
			if value.Error != nil {
				break
			}
			if ok {
				return MarshaledResult{
					Result:         res,
					MarshaledValue: val,
				}, true
			}
		}
		return value, ok
	}
	if len(i.keyCache) == 0 {
		newKeys, err := i.nextKeys()
		if err != nil {
			return MarshaledResult{
				Result: query.Result{
					Entry: query.Entry{},
					Error: err,
				},
			}, false
		}

		if len(newKeys) == 0 {
			return MarshaledResult{
				Result: query.Result{
					Entry: query.Entry{},
					Error: nil,
				},
			}, false
		}
		i.keyCache = append(i.keyCache, newKeys...)
	}

	key := i.keyCache[0]
	i.keyCache = i.keyCache[1:]

	value, err := i.txn.Get(context.Background(), key)
	if err != nil {
		return MarshaledResult{
			Result: query.Result{
				Entry: query.Entry{},
				Error: err,
			}}, false
	}
	return MarshaledResult{
		Result: query.Result{
			Entry: query.Entry{
				Key:   key.String(),
				Value: value,
			},
			Error: nil,
		}}, true
}

func (i *iterator) Close() {
	i.iter.Close()
}
