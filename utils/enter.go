package utils

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/zeromicro/go-zero/core/logx"
	"reflect"
	"regexp"
)

// InList 检查给定的key是否存在于list中。
// list: 要检查的字符串列表。
// key: 要查找的字符串。
// 返回值: 如果key存在于list中，则返回true；否则返回false。
func InList(list []string, key string) (ok bool) {
	for _, s := range list {
		if s == key {
			return true
		}
	}
	return false
}

// InListByRegex 检查给定的key是否与list中的任何正则表达式匹配。
// list: 包含正则表达式的字符串列表。
// key: 要匹配的字符串。
// 返回值: 如果key与任何正则表达式匹配，则返回true；否则返回false。
func InListByRegex(list []string, key string) (ok bool) {
	for _, s := range list {
		regex, err := regexp.Compile(s)
		if err != nil {
			logx.Error(err)
			return
		}
		if regex.MatchString(key) {
			return true
		}
	}
	return false
}

// MD5 计算给定数据的MD5哈希值。
// data: 要计算哈希值的數據。
// 返回值: 数据的MD5哈希值的字符串表示。
func MD5(data []byte) string {
	h := md5.New()
	h.Write(data)
	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

// DeduplicationList 去重列表
func DeduplicationList[T string | int | uint | uint32](req []T) (response []T) {
	i32Map := make(map[T]bool)
	for _, i32 := range req {
		if !i32Map[i32] {
			i32Map[i32] = true
		}
	}
	for key, _ := range i32Map {
		response = append(response, key)
	}
	return
}

// ReverseAny 通用反转函数，适用于任何实现了排序接口的类型。
// 参数 s 接收一个实现了排序接口的切片或数组的接口类型。
// 该函数通过反射来获取切片或数组的长度，并使用反射提供的Swapper函数进行元素交换，实现反转效果。
func ReverseAny(s interface{}) {
	// 通过反射获取传入切片或数组的长度
	n := reflect.ValueOf(s).Len()
	// 使用反射的Swapper函数获取可用于元素交换的函数
	swap := reflect.Swapper(s)
	// 使用双指针法从两端开始交换元素，直到中间位置，实现反转
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		// 调用swap函数进行元素交换
		swap(i, j)
	}
}
