package gmodel

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
)

// ID是递增的，所以不需要考虑删除，只管做映射即可
// 这个类的作用：将整型ID和字符串ID互相映射
// 只能转换整型ID到字符串ID
type IdMgr struct {
	db    *KVStore
	mutex sync.RWMutex
}

var (
	idPrefixInt    = "int_"
	idPrefixString = "str_"
)

// 打开数据库文件
func (this *IdMgr) Open(path string) error {
	this.db = &KVStore{}
	return this.db.Open(path)
}

func (this *IdMgr) Close() error {
	return this.db.Close()
}

// 获取记录的数量
func (this *IdMgr) Count() uint64 {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	// 注意要除以2
	return this.db.Count() / 2
}

// 增加一个整型ID
// 返回对应的字符串ID
func (this *IdMgr) AddIntId(intId uint64) (string, error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	// 如果 intId 已经存在，则返回对应的 stringId
	if this.hasIntId(intId) {
		if stringId, ok := this.getStringId(intId); ok {
			return stringId, nil
		} else {
			return "", errors.New(fmt.Sprintf("intId[%v] not found", intId))
		}
	}

	// 生成字符串ID
	stringId := this.generateStringId()

	return stringId, this.setIdMap(intId, stringId)
}

// 获取整型ID对应的字符串ID
func (this *IdMgr) GetStringId(intId uint64) (string, bool) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return this.getStringId(intId)
}

func (this *IdMgr) getStringId(intId uint64) (string, bool) {
	if value, err := this.db.Get(this.getKeyFromInt(intId)); err == nil {
		return string(value), true
	}
	return "", false
}

// 获取字符串ID对应的整型ID
func (this *IdMgr) GetIntId(stringId string) (uint64, bool) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	if value, err := this.db.Get(this.getKeyFromString(stringId)); err == nil {
		if intId, err := strconv.ParseUint(string(value), 10, 64); err == nil {
			return intId, true
		}
	}
	return 0, false
}

// 保存intId和stringId的映射
func (this *IdMgr) SetIdMap(intId uint64, stringId string) error {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return this.setIdMap(intId, stringId)
}

func (this *IdMgr) setIdMap(intId uint64, stringId string) error {
	// str_stringId --> intId
	err1 := this.db.Put(this.getKeyFromString(stringId), []byte(strconv.FormatUint(intId, 10)))

	// int_intId --> stringId
	err2 := this.db.Put(this.getKeyFromInt(intId), []byte(stringId))

	if err1 != nil || err2 != nil {
		// 如果有一个失败就全部删掉
		this.db.Delete(this.getKeyFromString(stringId))
		this.db.Delete(this.getKeyFromInt(intId))

		return errors.New(fmt.Sprintf("AddIntId[%v] failed, err1[%v] err2[%v]", intId, err1, err2))
	}

	return nil
}

func (this *IdMgr) getKeyFromInt(intId uint64) []byte {
	return []byte(idPrefixInt + strconv.FormatUint(intId, 10))
}

func (this *IdMgr) getKeyFromString(stringId string) []byte {
	return []byte(idPrefixString + stringId)
}

func (this *IdMgr) hasIntId(intId uint64) bool {
	return this.db.Has(this.getKeyFromInt(intId))
}

func (this *IdMgr) hasStringId(stringId string) bool {
	return this.db.Has(this.getKeyFromString(stringId))
}

// 产生一个字符串ID
func (this *IdMgr) generateStringId() string {
	for {
		stringId := string(RandomBytes(8, 8))
		if !this.hasStringId(stringId) {
			return stringId
		}
	}
	return ""
}
