package gmodel

import (
	"bytes"
	"errors"
	"strconv"
	"sync"
)

// 封装 article 和 tag 两个模块，方便外部使用
type GModel struct {
	articleMgr *ArticleMgr
	tagMgr     *TagMgr

	// 为了方便按分类查找文章，增加一个数据库用来作为索引
	// 索引格式：tagId_articleId -> articleId
	indexDB *KVStore

	// 全局一把锁
	mutex sync.RWMutex
}

// 初始化，使用三个数据库文件，分别存放文章、分类、索引
func (this *GModel) Open(articleDBPath, tagDBPath, indexDBPath string) error {
	this.articleMgr = &ArticleMgr{}
	this.tagMgr = &TagMgr{}
	this.indexDB = &KVStore{}

	if err := this.articleMgr.Open(articleDBPath); err != nil {
		return err
	}
	if err := this.tagMgr.Open(tagDBPath); err != nil {
		return err
	}
	if err := this.indexDB.Open(indexDBPath); err != nil {
		return err
	}

	return nil
}

func (this *GModel) Close() error {
	this.articleMgr.Close()
	this.tagMgr.Close()
	this.indexDB.Close()
	return nil
}

// 返回文章总数
func (this *GModel) GetArticleCount() uint64 {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return this.articleMgr.Count()
}

// 返回分类总数
func (this *GModel) GetTagCount() uint64 {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return this.tagMgr.Count()
}

// 返回当前最大（最新）的文章ID
func (this *GModel) GetMaxArticleId() uint64 {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return this.articleMgr.GetMaxId()
}

// 增加文章，返回文章ID
// tags：文章分类名称，可以为空，tags为空表示该文章属于未分类
// data: 文章数据内容，不能为空
func (this *GModel) AddArticle(tags []string, data string) (uint64, error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if len(data) == 0 {
		return 0, errors.New("GModel AddArticle data must not empty!")
	}

	// tags 为空，则补一个空字符串，方便索引，也就是每篇文章至少存在一个分类，该分类可以为空
	// tagMgr.Add 支持插入空字符串
	if len(tags) == 0 {
		tags = []string{""}
	}

	// 增加分类
	tagIds := make([]uint64, 0)
	for _, tag := range tags {
		id, err := this.tagMgr.Add(tag)
		if err == nil {
			tagIds = append(tagIds, id)
		}
	}

	article := &Article{
		TagIds: tagIds,
		Data:   data,
	}

	// 增加文章
	articleId, err := this.articleMgr.Add(article)
	if err != nil {
		return 0, err
	}

	// 增加索引
	this.addIndex(tagIds, articleId)

	return articleId, nil
}

// 删除文章
func (this *GModel) DeleteArticle(articleId uint64) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	article, err := this.articleMgr.GetById(articleId)
	if err != nil {
		return err
	}

	// TODO: 先不删除分类
	// 删除文章
	if err = this.articleMgr.Delete(article.Id); err != nil {
		return err
	}

	// 删除索引
	this.deleteIndex(article.TagIds, article.Id)

	return nil
}

// 获取文章
func (this *GModel) GetArticle(articleId uint64) (*Article, error) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return this.articleMgr.GetById(articleId)
}

// 获取指定文章的后N篇（不包括当前这篇）
// 如果 articleId 等于 0，则表示获取最旧的N篇文章（id最小的N篇）
func (this *GModel) GetNextArticles(articleId uint64, n int) []*Article {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return this.articleMgr.Next(articleId, n)
}

// 获取指定文章的前N篇（不包括当前这篇）
// 如果 articleId 大于 最大的文章ID，则表示获取最新的N篇文章（id最大的N篇）
func (this *GModel) GetPrevArticles(articleId uint64, n int) []*Article {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return this.articleMgr.Prev(articleId, n)
}

// 获取指定文章的后N篇（不包括当前这篇），保证这N篇文章的分类为tagName
// tagName为文章分类，如果tag不存在，则返回空数组，如果tag为空，则表示未分类，会返回未分类的文章
// articleId为文章ID，如果articleId为0，则返回该分类最旧的N篇文章（id最小的N篇）
func (this *GModel) GetNextArticlesByTag(tagName string, articleId uint64, n int) []*Article {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	articles := make([]*Article, 0)
	if n == 0 {
		return articles
	}

	tag, err := this.tagMgr.GetByName(tagName)
	if err != nil {
		return articles
	}

	// 按索引前缀查找
	// 比如 tagName 为 java， 对应的tagid 为 23， articleId 为 99,
	// 那么 prefix 为 23_
	// 那么 searchKey 为 23_000000000000099

	prefix := []byte(this.getIndexKeyPrefix(tag.Id))
	searchKey := this.getIndexKey(tag.Id, articleId)
	keys := this.indexDB.Next(searchKey, n)

	for _, key := range keys {
		if bytes.HasPrefix(key, prefix) {
			value, err := this.indexDB.Get(key)
			if err != nil {
				continue
			}

			if articleId, err := strconv.ParseUint(string(value), 10, 64); err == nil {
				if article, err := this.articleMgr.GetById(articleId); err == nil {
					articles = append(articles, article)
				}
			}
		}
	}

	return articles
}

// 获取指定文章的前N篇（不包括当前这篇），保证这N篇文章的分类为tagName
// tagName为文章分类，如果tag不存在，则返回空数组，如果tag为空，则表示未分类，会返回未分类的文章
// articleId为文章ID，如果 articleId 大于 最大的文章ID，则返回该分类最新的N篇文章（id最大的N篇）
func (this *GModel) GetPrevArticlesByTag(tagName string, articleId uint64, n int) []*Article {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	articles := make([]*Article, 0)
	if n == 0 {
		return articles
	}

	tag, err := this.tagMgr.GetByName(tagName)
	if err != nil {
		return articles
	}

	// 按索引前缀查找
	// 比如 tagName 为 java， 对应的tagid 为 23， articleId 为 99,
	// 那么 prefix 为 23_
	// 那么 searchKey 为 23_000000000000099

	prefix := []byte(this.getIndexKeyPrefix(tag.Id))
	searchKey := this.getIndexKey(tag.Id, articleId)
	// 注意，这里n+1，因为查找某个tag的最新N篇时，填的articleId一般是最大的id+1，
	// 那么指向的位置可能是下一个tag的第一篇，而这一篇会被过滤掉
	keys := this.indexDB.Prev(searchKey, n+1)

	for _, key := range keys {
		if bytes.HasPrefix(key, prefix) {
			value, err := this.indexDB.Get(key)
			if err != nil {
				continue
			}

			if articleId, err := strconv.ParseUint(string(value), 10, 64); err == nil {
				if article, err := this.articleMgr.GetById(articleId); err == nil {
					articles = append(articles, article)

					if len(articles) == n {
						break
					}
				}
			}
		}
	}

	return articles
}

// 修改文章
// articleId：待修改的文章ID
// newTags：新的分类名称，可以为空，tags为空表示该文章属于未分类
// newData: 文章数据内容，不能为空
func (this *GModel) UpdateArticle(articleId uint64, newTags []string, newData string) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	article, err := this.articleMgr.GetById(articleId)
	if err != nil {
		return err
	}

	// tags 为空，则补一个空字符串，方便索引
	// tagMgr.Add 支持插入空字符串
	if len(newTags) == 0 {
		newTags = []string{""}
	}

	// 增加分类
	tagIds := make([]uint64, 0)
	for _, tag := range newTags {
		id, err := this.tagMgr.Add(tag)
		if err == nil {
			tagIds = append(tagIds, id)
		}
	}

	// 删掉旧索引
	this.deleteIndex(article.TagIds, articleId)

	// 更新文章
	article.TagIds = tagIds
	article.Data = newData
	err = this.articleMgr.Update(article)
	if err != nil {
		return err
	}

	// 增加新索引
	this.addIndex(tagIds, articleId)

	return nil
}

// 根据分类ID获取分类
func (this *GModel) GetTagById(id uint64) (*Tag, error) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return this.tagMgr.GetById(id)
}

// 根据分类名称获取分类
func (this *GModel) GetTagByName(name string) (*Tag, error) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return this.tagMgr.GetByName(name)
}

// 修改分类名称
func (this *GModel) RenameTag(oldName, newName string) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	return this.tagMgr.Rename(oldName, newName)
}

// 增加索引
func (this *GModel) addIndex(tagIds []uint64, articleId uint64) {
	value := []byte(strconv.FormatUint(articleId, 10))
	for _, tagId := range tagIds {
		key := this.getIndexKey(tagId, articleId)
		this.indexDB.Put(key, value)
	}
}

// 删除索引
func (this *GModel) deleteIndex(tagIds []uint64, articleId uint64) {
	for _, tagId := range tagIds {
		key := this.getIndexKey(tagId, articleId)
		this.indexDB.Delete(key)
	}
}

// 返回索引key
func (this *GModel) getIndexKey(tagId uint64, articleId uint64) []byte {
	// 比如tagId为101，articleId为99
	// 那么存储的索引key为: 101_000000000000099
	return []byte(this.getIndexKeyPrefix(tagId) + string(this.articleMgr.getKeyFromId(articleId)))
}

// 返回索引key的前缀
func (this *GModel) getIndexKeyPrefix(tagId uint64) string {
	return strconv.FormatUint(tagId, 10) + "_"
}
