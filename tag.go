package gmodel

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

type Tag struct {
	Id           uint64 `json:"id"`            // 分类ID，从1开始自增，唯一标识，不允许修改
	Name         string `json:"name"`          // 分类名称，唯一标识，允许修改
	ArticleCount uint64 `json:"article_count"` // 该分类下的文章数量
}

var (
	// 同时将ID和名称作为key保存，ID加上 id_ 前缀，名称加上 name_ 前缀
	tagKeyPrefixId   = "id_"
	tagKeyPrefixName = "name_"
)

// 分类管理器
type TagMgr struct {
	db    *KVStore
	mutex sync.RWMutex
}

// 打开数据库文件
func (this *TagMgr) Open(path string) error {
	this.db = &KVStore{}
	return this.db.Open(path)
}

func (this *TagMgr) Close() error {
	return this.db.Close()
}

// 返回指定分类的后N个分类（不包括当前分类）
// 如果分类不存在，则表示获取最旧的N个分类
func (this *TagMgr) NextByName(name string, n int) []*Tag {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	var id uint64 = 0
	if tag, err := this.getByName(name); err == nil {
		id = tag.Id
	}

	return this.next(id, n)
}

// 返回指定分类ID的后N个分类（不包括当前id）
// 如果id等于0，则表示获取最旧的N个分类
func (this *TagMgr) Next(id uint64, n int) []*Tag {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	return this.next(id, n)
}

func (this *TagMgr) next(id uint64, n int) []*Tag {
	tags := make([]*Tag, 0)
	if n == 0 || id >= this.db.CurrentSequence() {
		return tags
	}

	// 存储的key是按id 15位补全的，这样id递增的时候也是按字典序递增的，
	// 所以即使中间删除了大量的tag，也不会影响查询效率
	searchKey := []byte(this.getKeyFromId(id))
	keys := this.db.Next(searchKey, n)

	for _, key := range keys {
		if bytes.HasPrefix(key, []byte(tagKeyPrefixId)) {
			// 读取数据
			if value, err := this.db.Get(key); err == nil {
				tag := &Tag{}
				if err = json.Unmarshal(value, tag); err == nil {
					tags = append(tags, tag)
				}
			}
		}
	}

	return tags
}

// 返回指定分类的前N个分类（不包括当前分类）
// 如果分类不存在，则表示获取最新的N个分类
func (this *TagMgr) PrevByName(name string, n int) []*Tag {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	var id uint64 = this.db.CurrentSequence() + 1
	if tag, err := this.getByName(name); err == nil {
		id = tag.Id
	}

	return this.prev(id, n)
}

// 返回指定分类ID的前N个分类（不包括当前id）
// 如果id大于CurrentSequence，则表示获取最新的N个分类
func (this *TagMgr) Prev(id uint64, n int) []*Tag {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	return this.prev(id, n)
}

func (this *TagMgr) prev(id uint64, n int) []*Tag {
	tags := make([]*Tag, 0)
	if n == 0 || id <= 1 {
		return tags
	}

	// id超过范围，则强制设置为当前最大，避免搜索时出现意外
	if id > this.db.CurrentSequence() {
		id = this.db.CurrentSequence()

		// 先判断当前id是否存在
		if value, err := this.db.Get(this.getKeyFromId(id)); err == nil {
			tag := &Tag{}
			if err = json.Unmarshal(value, tag); err == nil {
				tags = append(tags, tag)
				n = n - 1
			}
		}
	}

	searchKey := []byte(this.getKeyFromId(id))
	keys := this.db.Prev(searchKey, n)

	for _, key := range keys {
		if bytes.HasPrefix(key, []byte(tagKeyPrefixId)) {
			// 读取数据
			if value, err := this.db.Get(key); err == nil {
				tag := &Tag{}
				if err = json.Unmarshal(value, tag); err == nil {
					tags = append(tags, tag)
				}
			}
		}
	}

	return tags
}

// 获取分类数量
func (this *TagMgr) Count() uint64 {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	// 注意要除以2
	return this.db.Count() / 2
}

// 增加分类，返回分类ID
// 如果分类已经存在，则返回分类ID
func (this *TagMgr) Add(name string) (uint64, error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	tag, err := this.getByName(name)
	if err == nil {
		return tag.Id, nil
	}

	newTag := &Tag{}
	if newTag.Id, err = this.db.NextSequence(); err != nil {
		return 0, err
	}
	newTag.Name = name

	return newTag.Id, this.putTag(newTag)
}

// 保存分类
func (this *TagMgr) putTag(tag *Tag) error {
	value, err := json.Marshal(tag)
	if err != nil {
		return err
	}

	err1 := this.db.Put(this.getKeyFromId(tag.Id), value)
	err2 := this.db.Put(this.getKeyFromName(tag.Name), value)
	if err1 != nil || err2 != nil {
		// 如果有一个失败，则全部清空
		this.db.Delete(this.getKeyFromId(tag.Id))
		this.db.Delete(this.getKeyFromName(tag.Name))
		return errors.New(fmt.Sprintf("Tag put [%v %v] failed", tag.Id, tag.Name))
	}

	return nil
}

// 删除分类
func (this *TagMgr) DeleteByName(name string) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	tag, err := this.getByName(name)
	if err != nil {
		return err
	}
	return this.deleteTag(tag)
}

// 删除分类
func (this *TagMgr) DeleteById(id uint64) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	tag, err := this.getById(id)
	if err != nil {
		return err
	}
	return this.deleteTag(tag)
}

func (this *TagMgr) deleteTag(tag *Tag) error {
	this.db.Delete(this.getKeyFromId(tag.Id))
	this.db.Delete(this.getKeyFromName(tag.Name))
	return nil
}

// 修改分类名字
func (this *TagMgr) Rename(oldName string, newName string) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	oldTag, err := this.getByName(oldName)
	if err != nil {
		return err
	}

	_, err = this.getByName(newName)
	if err == nil {
		return errors.New(fmt.Sprintf("Tag newName[%v] exist", newName))
	}

	// 删除旧的
	this.deleteTag(oldTag)

	// 增加新的
	oldTag.Name = newName
	return this.putTag(oldTag)
}

// 增加（或减少）指定分类下的文章数量
// 返回更新后该分类下的文章数量
func (this *TagMgr) AddArticleCountForName(name string, count int64) (uint64, error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	tag, err := this.getByName(name)
	if err != nil {
		return 0, err
	}

	return this.addArticleCount(tag, count)
}

// 增加（或减少）指定分类下的文章数量
// 返回更新后该分类下的文章数量
func (this *TagMgr) AddArticleCountForId(id uint64, count int64) (uint64, error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	tag, err := this.getById(id)
	if err != nil {
		return 0, err
	}

	return this.addArticleCount(tag, count)
}

func (this *TagMgr) addArticleCount(tag *Tag, count int64) (uint64, error) {
	// TODO 不需要考虑溢出情况
	num := int64(tag.ArticleCount)
	num += count
	if num < 0 {
		num = 0
	}
	tag.ArticleCount = uint64(num)

	return uint64(num), this.putTag(tag)
}

// 获取指定分类下的文章数量
func (this *TagMgr) GetArticleCountByName(name string) uint64 {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	if tag, err := this.getByName(name); err == nil {
		return tag.ArticleCount
	}

	return 0
}

// 获取指定分类下的文章数量
func (this *TagMgr) GetArticleCountById(id uint64) uint64 {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	if tag, err := this.getById(id); err == nil {
		return tag.ArticleCount
	}

	return 0
}

// 根据ID获取分类
func (this *TagMgr) GetById(id uint64) (*Tag, error) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return this.getById(id)
}

func (this *TagMgr) getById(id uint64) (*Tag, error) {
	value, err := this.db.Get(this.getKeyFromId(id))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Tag ID[%v] not found", id))
	}

	tag := &Tag{}
	err = json.Unmarshal(value, tag)
	return tag, err
}

// 根据名字获取分类
func (this *TagMgr) GetByName(name string) (*Tag, error) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return this.getByName(name)
}

func (this *TagMgr) getByName(name string) (*Tag, error) {
	value, err := this.db.Get(this.getKeyFromName(name))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Tag Name[%v] not found", name))
	}

	tag := &Tag{}
	err = json.Unmarshal(value, tag)
	return tag, err
}

// 根据分类ID获取存储key
func (this *TagMgr) getKeyFromId(id uint64) []byte {
	return []byte(tagKeyPrefixId + GetStringKey(id))
}

// 根据分类名称获取存储key
func (this *TagMgr) getKeyFromName(name string) []byte {
	return []byte(tagKeyPrefixName + name)
}
