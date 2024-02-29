package jsonpatcher

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	core "github.com/dcnetio/gothreads-lib/core/db"
	jsonpatch "github.com/evanphx/json-patch"
	ds "github.com/ipfs/go-datastore"
	cbornode "github.com/ipfs/go-ipld-cbor"
	format "github.com/ipfs/go-ipld-format"
	logging "github.com/ipfs/go-log/v2"
	"github.com/multiformats/go-multihash"
)

type operationType int

const (
	create operationType = iota
	save
	del
)

func (ot operationType) String() (s string) {
	switch ot {
	case create:
		s = "create"
	case save:
		s = "save"
	case del:
		s = "delete"
	}
	return s
}

var (
	log                           = logging.Logger("jsonpatcher")
	errCantCreateExistingInstance = errors.New("cant't create already existent instance")
	errUnknownOperation           = errors.New("unknown operation type")
)

type operation struct {
	Type       operationType
	InstanceID core.InstanceID
	JSONPatch  []byte
}

type jsonPatcher struct{}

var _ core.EventCodec = (*jsonPatcher)(nil)

func init() {
	cbornode.RegisterCborType(patchEvent{})
	cbornode.RegisterCborType(recordEvents{})
	cbornode.RegisterCborType(operation{})
}

// New returns a JSON-Patcher EventCodec
func New() core.EventCodec {
	return &jsonPatcher{}
}

// Create returns a new JSON-Patcher event from a set of actions
func (jp *jsonPatcher) Create(actions []core.Action) ([]core.Event, format.Node, error) {
	if len(actions) == 0 {
		return nil, nil, nil
	}
	revents := recordEvents{Patches: make([]patchEvent, len(actions))}
	events := make([]core.Event, len(actions))
	for i := range actions {
		var op *operation
		var err error
		switch actions[i].Type {
		case core.Create:
			op, err = createEvent(actions[i].InstanceID, actions[i].Current)
		case core.Save:
			op, err = saveEvent(actions[i].InstanceID, actions[i].Previous, actions[i].Current)
		case core.Delete:
			op, err = deleteEvent(actions[i].InstanceID)
		default:
			panic("unkown action type")
		}
		if err != nil {
			return nil, nil, err
		}
		revents.Patches[i] = patchEvent{
			Timestamp:      time.Now().UnixNano(),
			ID:             actions[i].InstanceID,
			CollectionName: actions[i].CollectionName,
			Patch:          *op,
		}
		events[i] = revents.Patches[i]
	}

	n, err := cbornode.WrapObject(revents, multihash.SHA2_256, -1)
	if err != nil {
		return nil, nil, err
	}
	return events, n, nil
}

func (jp *jsonPatcher) Reduce(
	events []core.Event,
	store ds.TxnDatastore,
	baseKey ds.Key,
	indexFunc core.IndexFunc,
) ([]core.ReduceAction, error) {
	txn, err := store.NewTransaction(context.Background(), false)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(context.Background())

	sort.Slice(events, func(i, j int) bool {
		ei, oki := events[i].(patchEvent)
		ej, okj := events[j].(patchEvent)

		if !(oki && okj) {
			return false
		}

		return ei.time().Before(ej.time())
	})

	actions := make([]core.ReduceAction, len(events))
	for i, e := range events {
		je, ok := e.(patchEvent)
		if !ok {
			return nil, fmt.Errorf("event unrecognized for jsonpatcher eventcodec")
		}
		key := baseKey.ChildString(e.Collection()).ChildString(e.InstanceID().String())
		switch je.Patch.Type {
		case create:
			exist, err := txn.Has(context.Background(), key)
			if err != nil {
				return nil, err
			}
			if exist {
				return nil, fmt.Errorf("%s,key:%s,patch:%s", errCantCreateExistingInstance.Error(), key, je.Patch.JSONPatch)
			}
			if err := txn.Put(context.Background(), key, je.Patch.JSONPatch); err != nil {
				return nil, fmt.Errorf("error when reducing create event: %w", err)
			}
			if err := indexFunc(e.Collection(), key, nil, je.Patch.JSONPatch, txn); err != nil {
				return nil, fmt.Errorf("error when indexing created data: %w", err)
			}
			actions[i] = core.ReduceAction{Type: core.Create, Collection: e.Collection(), InstanceID: e.InstanceID()}

			log.Debug("\tcreate operation applied")
		case save:
			value, err := txn.Get(context.Background(), key)
			if errors.Is(err, ds.ErrNotFound) {
				value = []byte("{}")
			} else if err != nil {
				return nil, err
			}
			patchedValue, err := jsonpatch.MergePatch(value, je.Patch.JSONPatch)
			if err != nil {
				return nil, fmt.Errorf("error when reducing save event: %w", err)
			}
			if err = txn.Put(context.Background(), key, patchedValue); err != nil {
				return nil, err
			}
			if err := indexFunc(e.Collection(), key, value, patchedValue, txn); err != nil {
				return nil, fmt.Errorf("error when indexing created data: %w", err)
			}
			actions[i] = core.ReduceAction{Type: core.Save, Collection: e.Collection(), InstanceID: e.InstanceID()}
			log.Debug("\tsave operation applied")
			//to do need delete for release
		case del:
			value, err := txn.Get(context.Background(), key)
			if err != nil {
				return nil, err
			}
			if err := txn.Delete(context.Background(), key); err != nil {
				return nil, err
			}
			if err := indexFunc(e.Collection(), key, value, nil, txn); err != nil {
				return nil, fmt.Errorf("error when removing index: %w", err)
			}
			actions[i] = core.ReduceAction{Type: core.Delete, Collection: e.Collection(), InstanceID: e.InstanceID()}
			log.Debug("\tdelete operation applied")
		default:
			return nil, errUnknownOperation
		}
	}
	if err := txn.Commit(context.Background()); err != nil {
		return nil, err
	}

	return actions, nil
}

type recordEvents struct {
	Patches []patchEvent
}

// EventsFromBytes returns a unmarshaled event from its bytes representation
func (jp *jsonPatcher) EventsFromBytes(data []byte) ([]core.Event, error) {
	revents := recordEvents{}
	if err := cbornode.DecodeInto(data, &revents); err != nil {
		return nil, err
	}

	res := make([]core.Event, len(revents.Patches))
	for i := range revents.Patches {
		res[i] = revents.Patches[i]
	}

	return res, nil
}

func createEvent(id core.InstanceID, v []byte) (*operation, error) {
	return &operation{
		Type:       create,
		InstanceID: id,
		JSONPatch:  v,
	}, nil
}

func saveEvent(id core.InstanceID, prev []byte, curr []byte) (*operation, error) {
	jsonPatch, err := jsonpatch.CreateMergePatch(prev, curr)
	if err != nil {
		return nil, err
	}
	return &operation{
		Type:       save,
		InstanceID: id,
		JSONPatch:  jsonPatch,
	}, nil
}

func deleteEvent(id core.InstanceID) (*operation, error) {
	return &operation{
		Type:       del,
		InstanceID: id,
		JSONPatch:  nil,
	}, nil
}

type patchEvent struct {
	Timestamp      interface{}
	ID             core.InstanceID
	CollectionName string
	Patch          operation
}

func (je patchEvent) Time() []byte {
	var nanos int64
	switch ts := je.Timestamp.(type) {
	case time.Time:
		nanos = ts.UnixNano()
	case int64:
		nanos = ts
	case int:
		nanos = int64(ts)
	}
	buf := new(bytes.Buffer)
	// Use big endian to preserve lexicographic sorting
	_ = binary.Write(buf, binary.BigEndian, nanos)
	return buf.Bytes()
}

func (je patchEvent) time() (t time.Time) {
	switch ts := je.Timestamp.(type) {
	case time.Time:
		t = ts
	case int:
		t = time.Unix(0, int64(ts))
	case uint:
		t = time.Unix(0, int64(ts))
	case int64:
		t = time.Unix(0, ts)
	case uint64:
		t = time.Unix(0, int64(ts))
	}
	return t
}

func (je patchEvent) InstanceID() core.InstanceID {
	return je.ID
}

func (je patchEvent) Collection() string {
	return je.CollectionName
}

type patchEventJson struct {
	Timestamp      interface{}   `json:"timestamp"`
	ID             string        `json:"_id"`
	CollectionName string        `json:"collection_name"`
	Patch          operationJson `json:"patch"`
}

type operationJson struct {
	Type       string      `json:"type"`
	InstanceID string      `json:"instance_id"`
	JSONPatch  interface{} `json:"json_patch,omitempty"`
}

func (je patchEvent) Marshal() ([]byte, error) {
	var patch interface{}
	if je.Patch.JSONPatch != nil {
		if err := json.Unmarshal(je.Patch.JSONPatch, &patch); err != nil {
			return nil, err
		}
	}
	return json.Marshal(patchEventJson{
		Timestamp:      je.Timestamp,
		ID:             string(je.ID),
		CollectionName: je.CollectionName,
		Patch: operationJson{
			Type:       je.Patch.Type.String(),
			InstanceID: string(je.Patch.InstanceID),
			JSONPatch:  patch,
		},
	})
}

var _ core.Event = (*patchEvent)(nil)
