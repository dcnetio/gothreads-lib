package lstoremem

import (
	core "github.com/dcnetio/gothreads-lib/core/logstore"
	lstore "github.com/dcnetio/gothreads-lib/logstore"
)

// Define if storage will accept empty dumps.
var AllowEmptyRestore = true

// NewLogstore creates an in-memory threadsafe collection of thread logs.
func NewLogstore() core.Logstore {
	return lstore.NewLogstore(
		NewKeyBook(),
		NewAddrBook(),
		NewHeadBook(),
		NewThreadMetadata())
}
