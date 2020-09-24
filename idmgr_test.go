package gmodel

import (
	"os"
	"testing"
)

func TestIdMgr(t *testing.T) {
	dbPath := "test.db"
	defer os.RemoveAll(dbPath)

	mgr := &IdMgr{}
	err := mgr.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	if mgr.Count() != 0 {
		t.Fatal()
	}

	for i := 1; i < 1000; i++ {
		stringId, err := mgr.AddIntId(uint64(i))
		if err != nil {
			t.Fatal()
		}

		intId, ok := mgr.GetIntId(stringId)
		if !ok || intId != uint64(i) {
			t.Fatal()
		}

		strId, ok := mgr.GetStringId(intId)
		if !ok || strId != stringId {
			t.Fatal()
		}

		if mgr.Count() != uint64(i) {
			t.Fatal()
		}
	}

}
