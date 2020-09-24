package gmodel

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
)

// 注意：
// leveldb.OpenFile 返回的对象是线程安全的，详见：https://github.com/syndtr/goleveldb
// 需要加锁是因为要维护 keyForCount、 keyForSequence 等成员变量，只需要在读写成员变量的地方加读写锁即可。
// 另外，leveldb.Get 等接口返回的字节数组是不允许修改的，为了安全，最好是先拷贝一份再返回

var (
	// 由于leveldb没有接口获取key的数量，所以需要自己维护一个key来存储key的总数
	keyForCount = []byte("__key_for_count__")

	// 用于序号递增
	keyForSequence = []byte("__key_for_sequence__")

	// 内部保留key不允许被外界直接读取
	reservedlKeys = make([][]byte, 0)
)

func init() {
	reservedlKeys = append(reservedlKeys, keyForCount)
	reservedlKeys = append(reservedlKeys, keyForSequence)
}

func isReservedlKey(key []byte) bool {
	for _, reservedlKey := range reservedlKeys {
		if bytes.Equal(key, reservedlKey) {
			return true
		}
	}
	return false
}

type KVStore struct {
	db     *leveldb.DB
	dbPath string

	// 保护 keyForCount、keyForSequence 等成员变量的读写
	mutex sync.RWMutex
}

func (this *KVStore) Open(path string) error {
	var err error
	if this.db, err = leveldb.OpenFile(path, nil); err != nil {
		return errors.New(fmt.Sprintf("KVStore open [%v] failed: %v", path, err))
	}
	this.dbPath = path
	log.Printf("KVStore open [%v] success\n", path)
	return nil
}

func (this *KVStore) Close() error {
	log.Printf("KVStore close [%v]\n", this.dbPath)
	return this.db.Close()
}

func (this *KVStore) Put(key, value []byte) error {
	if isReservedlKey(key) {
		return errors.New("Not allow put reserved key")
	}

	if this.Has(key) {
		return this.db.Put(key, value, nil)
	}

	this.mutex.Lock()
	defer this.mutex.Unlock()

	// key总数+1
	count := this.count() + 1

	// 需要使用批处理，同时更新两个key
	batch := new(leveldb.Batch)
	batch.Put(key, value)
	batch.Put(keyForCount, []byte(strconv.FormatUint(count, 10)))

	return this.db.Write(batch, nil)
}

func (this *KVStore) Get(key []byte) ([]byte, error) {
	if isReservedlKey(key) {
		return nil, errors.New("Not allow get reserved key")
	}

	value, err := this.db.Get(key, nil)
	if err == nil {
		copyValue := make([]byte, len(value))
		copy(copyValue, value)
		return copyValue, nil
	}
	return nil, err
}

func (this *KVStore) Delete(key []byte) error {
	if isReservedlKey(key) {
		return errors.New("Not allow delete reserved key")
	}

	if this.Has(key) {
		this.mutex.Lock()
		defer this.mutex.Unlock()

		// key总数-1
		count := this.count() - 1

		// 需要使用批处理，同时更新两个key
		batch := new(leveldb.Batch)
		batch.Delete(key)
		batch.Put(keyForCount, []byte(strconv.FormatUint(count, 10)))

		return this.db.Write(batch, nil)
	}

	return errors.New("Key not exist")
}

func (this *KVStore) Has(key []byte) bool {
	exist, err := this.db.Has(key, nil)
	if err == nil && exist {
		return true
	}
	return false
}

// 返回指定key后面的n个key（不包括当前key，当前key也可以不存在）
// 如果当前key为空数组或者nil，表示从头开始遍历
// 如果当前key为字典序最大，则返回的结果为空；如果当前key为字典序最小，则返回最前的n个key
// 注意：leveldb 是根据 key 的字典序排序的
func (this *KVStore) Next(key []byte, n int) [][]byte {
	keys := make([][]byte, 0)
	iter := this.db.NewIterator(nil, nil)

	ok := false
	if key == nil || bytes.Equal(key, []byte("")) {
		ok = iter.First()
	} else {
		ok = iter.Seek(key)
	}

	for ; ok && len(keys) < n; ok = iter.Next() {
		// 过滤当前key 和 保留key
		if bytes.Equal(iter.Key(), key) || isReservedlKey(iter.Key()) {
			continue
		}

		copyKey := make([]byte, len(iter.Key()))
		copy(copyKey, iter.Key())
		keys = append(keys, copyKey)
	}
	iter.Release()

	return keys
}

// 返回指定key前面的n个key（不包括当前key，当前key也可以不存在）
// 如果当前key为空数组或者nil，表示从末尾开始遍历
// 如果当前key为字典序最小，则返回的结果为空；如果当前key为字典序最大，则返回最后的n个key
// 注意：leveldb 是根据 key 的字典序排序的
func (this *KVStore) Prev(key []byte, n int) [][]byte {
	keys := make([][]byte, 0)
	iter := this.db.NewIterator(nil, nil)

	ok := false
	if key == nil || bytes.Equal(key, []byte("")) {
		ok = iter.Last()
	} else {
		ok = iter.Seek(key)
	}

	for ; ok && len(keys) < n; ok = iter.Prev() {
		// 过滤当前key 和 保留key
		if bytes.Equal(iter.Key(), key) || isReservedlKey(iter.Key()) {
			continue
		}

		copyKey := make([]byte, len(iter.Key()))
		copy(copyKey, iter.Key())
		keys = append(keys, copyKey)
	}
	iter.Release()

	return keys
}

// 返回 key 的数量
func (this *KVStore) Count() uint64 {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return this.count()
}

func (this *KVStore) count() uint64 {
	if !this.Has(keyForCount) {
		return 0
	}

	value, err := this.db.Get(keyForCount, nil)
	if err != nil {
		return 0
	}

	count, err := strconv.ParseUint(string(value), 10, 64)
	if err != nil {
		return 0
	}
	return count
}

// 返回当前Sequence
func (this *KVStore) CurrentSequence() uint64 {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return this.currentSequence()
}

func (this *KVStore) currentSequence() uint64 {
	if !this.Has(keyForSequence) {
		return 0
	}

	value, err := this.db.Get(keyForSequence, nil)
	if err != nil {
		return 0
	}

	sequence, err := strconv.ParseUint(string(value), 10, 64)
	if err != nil {
		return 0
	}
	return sequence
}

// 生成并返回下一个Sequence
// 注意这是一个读写操作
func (this *KVStore) NextSequence() (uint64, error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	sequence := this.currentSequence() + 1
	err := this.db.Put(keyForSequence, []byte(strconv.FormatUint(sequence, 10)), nil)

	return sequence, err
}
