package net

import (
	"context"

	"github.com/dcnetio/gothreads-lib/crypto"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
)

// Event is the Block format used by threads
type Event interface {
	format.Node

	// HeaderID returns the cid of the event header.
	HeaderID() cid.Cid

	// GetHeader loads and optionally decrypts the event header.
	// If no key is given, the header time and key methods will return an error.
	GetHeader(context.Context, format.DAGService, crypto.DecryptionKey) (EventHeader, error)

	// BodyID returns the cid of the event body.
	BodyID() cid.Cid

	// GetBody loads and optionally decrypts the event body.
	GetBody(context.Context, format.DAGService, crypto.DecryptionKey) (format.Node, error)
}

// EventHeader is the format of the event's header object
type EventHeader interface {
	format.Node

	// Key returns a single-use decryption key for the event body.
	Key() (crypto.DecryptionKey, error)
}
