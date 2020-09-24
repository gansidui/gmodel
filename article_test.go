package gmodel

import (
	"fmt"
	"os"
	"strconv"
	"testing"
)

func TestArticle(t *testing.T) {
	dbPath := "test.db"
	defer os.RemoveAll(dbPath)

	mgr := &ArticleMgr{}
	err := mgr.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	article := &Article{
		TagIds: []uint64{1, 2, 3},
		Data:   "data1",
	}

	id, err := mgr.Add(article)
	if err != nil {
		t.Fatal()
	}
	if id != 1 {
		t.Fatal()
	}
	if mgr.Count() != 1 {
		t.Fatal()
	}

	article, err = mgr.GetById(id)
	if err != nil || article.Id != id {
		t.Fatal()
	}

	for i := 0; i < 100; i++ {
		article := &Article{}
		article.TagIds = []uint64{100 + uint64(i), uint64(i)}
		article.Data = "data " + strconv.Itoa(i)

		id, err := mgr.Add(article)
		if err != nil {
			t.Fatal()
		}

		art, err := mgr.GetById(id)
		if err != nil {
			t.Fatal()
		}
		if art.Data != article.Data {
			t.Fatal()
		}
	}
	if mgr.Count() != uint64(101) {
		t.Fatal()
	}

	articles := mgr.Next(0, 5)
	if len(articles) != 5 {
		t.Fatal()
	}
	for i := 1; i <= 5; i++ {
		if articles[i-1].Id != uint64(i) {
			t.Fatal()
		}
	}

	articles = mgr.Next(100, 2)
	if len(articles) != 1 || articles[0].Id != 101 {
		t.Fatal()
	}

	articles = mgr.Next(101, 1)
	if len(articles) != 0 {
		t.Fatal()
	}

	articles = mgr.Next(102, 1)
	if len(articles) != 0 {
		t.Fatal()
	}

	articles = mgr.Prev(0, 5)
	if len(articles) != 0 {
		t.Fatal()
	}

	articles = mgr.Prev(1, 1)
	if len(articles) != 0 {
		t.Fatal()
	}

	articles = mgr.Prev(101, 2)
	if len(articles) != 2 || articles[0].Id != 100 || articles[1].Id != 99 {
		t.Fatal()
	}
}

func TestArticleUpdate(t *testing.T) {
	dbPath := "test.db"
	defer os.RemoveAll(dbPath)

	mgr := &ArticleMgr{}
	err := mgr.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	article := &Article{
		TagIds: []uint64{1, 2, 3},
		Data:   "data1",
	}

	id, err := mgr.Add(article)
	if err != nil {
		t.Fatal()
	}
	if mgr.Count() != 1 {
		t.Fatal()
	}

	article = &Article{
		Id:     id,
		TagIds: []uint64{99, 98, 97},
		Data:   "newdata",
	}
	if err = mgr.Update(article); err != nil {
		t.Fatal()
	}

	articles := mgr.Next(0, 10)
	if len(articles) != 1 {
		t.Fatal()
	}
	if articles[0].Id != article.Id || articles[0].Data != article.Data {
		t.Fatal()
	}
	fmt.Println(articles[0].TagIds)

	err = mgr.Delete(article.Id)
	if err != nil {
		t.Fatal()
	}
	err = mgr.Delete(article.Id)
	if err == nil {
		t.Fatal()
	}

	_, err = mgr.GetById(article.Id)
	if err == nil {
		t.Fatal()
	}
	if mgr.Count() != 0 {
		t.Fatal()
	}
}
