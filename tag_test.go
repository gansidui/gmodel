package gmodel

import (
	"os"
	"strconv"
	"testing"
)

func TestTag(t *testing.T) {
	dbPath := "test.db"
	defer os.RemoveAll(dbPath)

	mgr := &TagMgr{}
	err := mgr.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	if id, err := mgr.Add("tag1"); err != nil || id != uint64(1) {
		t.Fatal(err)
	}

	mgr.Add("tag2")

	if id, _ := mgr.Add("tag3"); id != uint64(3) {
		t.Fatal()
	}

	if id, err := mgr.Add("tag3"); err != nil || id != uint64(3) {
		t.Fatal()
	}
	if id, err := mgr.Add("tag2"); err != nil || id != uint64(2) {
		t.Fatal()
	}

	// 这里插入一个空的tag
	id, err := mgr.Add("")
	if err != nil || id != uint64(4) {
		t.Fatal()
	}

	// 第二次重复插入不影响
	id, err = mgr.Add("")
	if err != nil || id != uint64(4) {
		t.Fatal()
	}

	if mgr.Count() != 4 {
		t.Fatal()
	}

	if err = mgr.Delete(""); err != nil {
		t.Fatal()
	}

	tag, err := mgr.getByName("tag2")
	if err != nil {
		t.Fatal(err)
	}
	if tag.Id != 2 || tag.Name != "tag2" {
		t.Fatal()
	}

	tag, err = mgr.getById(1)
	if err != nil || tag.Id != 1 || tag.Name != "tag1" {
		t.Fatal()
	}

	if err = mgr.Delete("tag1"); err != nil {
		t.Fatal(err)
	}
	if mgr.Count() != 2 {
		t.Fatal()
	}

	_, err = mgr.getByName("tag1")
	if err == nil {
		t.Fatal(err)
	}

	if err = mgr.Rename("tag2", "tag100"); err != nil {
		t.Fatal(err)
	}
	if mgr.Count() != 2 {
		t.Fatal()
	}
	tag, _ = mgr.getById(2)
	if tag.Name != "tag100" {
		t.Fatal()
	}

}

func TestTagId(t *testing.T) {
	dbPath := "test.db"
	defer os.RemoveAll(dbPath)

	mgr := &TagMgr{}
	err := mgr.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	limit := 10000
	for i := 1; i <= limit; i++ {
		name := strconv.Itoa(i)
		mgr.Add(name)

		if mgr.Count() != uint64(i) {
			t.Fatal()
		}

		tag, err := mgr.getById(uint64(i))
		if err != nil {
			t.Fatal()
		}
		if tag.Id != uint64(i) || tag.Name != name {
			t.Fatal()
		}
	}

	if mgr.Count() != 10000 {
		t.Fatal()
	}

	tags := mgr.Next(0, 10)
	if len(tags) != 10 {
		t.Fatal()
	}
	for i := 1; i <= len(tags); i++ {
		if tags[i-1].Id != uint64(i) {
			t.Fatal()
		}
	}

	tags = mgr.Next(1, 2)
	if len(tags) != 2 {
		t.Fatal()
	}
	if tags[0].Id != uint64(2) || tags[1].Id != uint64(3) {
		t.Fatal()
	}

	tags = mgr.Next(10000, 2)
	if len(tags) != 0 {
		t.Fatal()
	}

	tags = mgr.Next(9999, 2)
	if len(tags) != 1 {
		t.Fatal()
	}
	if tags[0].Id != 10000 {
		t.Fatal()
	}

	tags = mgr.Prev(11, 10)
	if len(tags) != 10 {
		t.Fatal()
	}
	for i := 1; i <= len(tags); i++ {
		if tags[i-1].Id != 11-uint64(i) {
			t.Fatal()
		}
	}

	count := mgr.Count()
	if count != uint64(limit) {
		t.Fatal()
	}

	tags = mgr.Prev(count, 2)
	if len(tags) != 2 {
		t.Fatal()
	}
	if tags[0].Id != uint64(limit-1) || tags[1].Id != uint64(limit-2) {
		t.Fatal()
	}

	tags = mgr.Prev(count+1, 2)
	if len(tags) != 2 {
		t.Fatal()
	}
	if tags[0].Id != uint64(limit) || tags[1].Id != uint64(limit-1) {
		t.Fatal()
	}

	tags = mgr.Prev(0, 2)
	if len(tags) != 0 {
		t.Fatal()
	}

	tags = mgr.Prev(1, 1)
	if len(tags) != 0 {
		t.Fatal()
	}

}
