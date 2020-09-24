package gmodel

import (
	"fmt"
	"math/rand"
	"time"
)

// 得到一个长度在区间[minLen, maxLen]内的随机字符数组
// 字符内容为：a-zA-Z0-9，去掉 l L 1  0 o O 这6个不易识别的字符，每个字符有62-6=56种可能
// 一般6位长度就够了：56^6 = 30840979456  =  308亿
func RandomBytes(minLen, maxLen int) []byte {
	num := 0
	if minLen < maxLen {
		rand.Seed(time.Now().UnixNano())
		num = rand.Intn(maxLen-minLen+1) + minLen
	} else {
		num = minLen
	}

	byteArray := make([]byte, num)
	const alphabet = "abcdefghijkmnpqrstuvwxyzABCDEFGHIJKMNPQRSTUVWXYZ23456789"
	alphabetLen := len(alphabet)

	for i := range byteArray {
		rand.Seed(time.Now().UnixNano())
		byteArray[i] = alphabet[rand.Intn(alphabetLen)]
	}

	return byteArray
}

// 将uint64转成string，不足15位时左边补0，这样id递增的时候，对应的key也是字典序递增的
// leveldb是按key的字典序存储的，方便查找
func GetStringKey(id uint64) string {
	return fmt.Sprintf("%015v", id)
}
