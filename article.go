package gmodel

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

type Article struct {
	Id     uint64   `json:"id"`      // 文章ID，从1开始自增，唯一标识，不允许修改
	TagIds []uint64 `json:"tag_ids"` // 文章分类，多个分类ID
	Data   string   `json:"data"`    // 文章数据，由上层解析
}

type ArticleMgr struct {
	db    *KVStore
	mutex sync.RWMutex
}

// 打开数据库文件
func (this *ArticleMgr) Open(path string) error {
	this.db = &KVStore{}
	return this.db.Open(path)
}

func (this *ArticleMgr) Close() error {
	return this.db.Close()
}

// 返回指定文章ID的后N篇文章（不包括当前id）
// 如果id等于0，则表示获取最旧的N篇文章
func (this *ArticleMgr) Next(id uint64, n int) []*Article {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	articles := make([]*Article, 0)
	if n == 0 || id >= this.db.CurrentSequence() {
		return articles
	}

	// 存储的key是按id 15位补全的，这样id递增的时候也是按字典序递增的，
	// 所以即使中间删除了大量的article，也不会影响查询效率
	searchKey := []byte(this.getKeyFromId(id))
	keys := this.db.Next(searchKey, n)

	for _, key := range keys {
		if value, err := this.db.Get(key); err == nil {
			article := &Article{}
			if err = json.Unmarshal(value, article); err == nil {
				articles = append(articles, article)
			}
		}
	}

	return articles
}

// 返回指定文章ID的前N篇文章（不包括当前id）
// 如果id大于CurrentSequence，则表示获取最新的N篇文章
func (this *ArticleMgr) Prev(id uint64, n int) []*Article {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	articles := make([]*Article, 0)
	if n == 0 || id <= 1 {
		return articles
	}

	// id超过范围，则强制设置为当前最大，避免搜索时出现意外
	if id > this.db.CurrentSequence() {
		id = this.db.CurrentSequence()

		// 先判断当前id是否存在
		if value, err := this.db.Get(this.getKeyFromId(id)); err == nil {
			article := &Article{}
			if err = json.Unmarshal(value, article); err == nil {
				articles = append(articles, article)
				n = n - 1
			}
		}
	}

	searchKey := []byte(this.getKeyFromId(id))
	keys := this.db.Prev(searchKey, n)

	for _, key := range keys {
		if value, err := this.db.Get(key); err == nil {
			article := &Article{}
			if err = json.Unmarshal(value, article); err == nil {
				articles = append(articles, article)
			}
		}
	}

	return articles
}

// 获取文章数量
func (this *ArticleMgr) Count() uint64 {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return this.db.Count()
}

// 返回当前最大（最新）的文章ID
func (this *ArticleMgr) GetMaxId() uint64 {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return this.db.CurrentSequence()
}

// 增加文章，返回文章ID
func (this *ArticleMgr) Add(article *Article) (uint64, error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	var err error
	if article.Id, err = this.db.NextSequence(); err != nil {
		return 0, err
	}

	return article.Id, this.putArticle(article)
}

// 保存文章
func (this *ArticleMgr) putArticle(article *Article) error {
	key := this.getKeyFromId(article.Id)
	value, err := json.Marshal(article)
	if err != nil {
		return err
	}
	return this.db.Put(key, value)
}

// 删除文章
func (this *ArticleMgr) Delete(id uint64) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	return this.db.Delete([]byte(this.getKeyFromId(id)))
}

// 修改文章
func (this *ArticleMgr) Update(article *Article) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	exist := this.has(article.Id)
	if !exist {
		return errors.New(fmt.Sprintf("Article ID[%v] not found", article.Id))
	}

	return this.putArticle(article)
}

// 获取文章
func (this *ArticleMgr) GetById(id uint64) (*Article, error) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return this.getById(id)
}

func (this *ArticleMgr) getById(id uint64) (*Article, error) {
	value, err := this.db.Get(this.getKeyFromId(id))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Article ID[%v] not found", id))
	}

	article := &Article{}
	err = json.Unmarshal(value, article)
	return article, err
}

// 判断文章ID是否存在
func (this *ArticleMgr) has(id uint64) bool {
	_, err := this.db.Get(this.getKeyFromId(id))
	if err == nil {
		return true
	}
	return false
}

// 根据文章ID获取存储key
func (this *ArticleMgr) getKeyFromId(id uint64) []byte {
	return []byte(GetStringKey(id))
}
