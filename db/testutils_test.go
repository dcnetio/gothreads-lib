package db

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/dcnetio/gothreads-lib/common"
	"github.com/dcnetio/gothreads-lib/core/thread"
	"github.com/dcnetio/gothreads-lib/util"
)

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func createTestDB(t *testing.T, opts ...NewOption) (*DB, func()) {
	dir, err := os.MkdirTemp("", "")
	checkErr(t, err)
	n, err := common.DefaultNetwork(
		common.WithNetBadgerPersistence(dir),
		common.WithNetHostAddr(util.FreeLocalAddr()),
		common.WithNetDebug(true),
	)
	checkErr(t, err)
	store, err := util.NewBadgerDatastore(dir, "eventstore", false)
	checkErr(t, err)
	d, err := NewDB(context.Background(), store, n, thread.NewIDV1(thread.Raw, 32), opts...)
	checkErr(t, err)
	return d, func() {
		time.Sleep(time.Second) // Give threads a chance to finish work
		if err := n.Close(); err != nil {
			panic(err)
		}
		if err := store.Close(); err != nil {
			panic(err)
		}
		_ = os.RemoveAll(dir)
	}
}
