package pwd

import (
	"golang.org/x/crypto/bcrypt"
	"log"
)

// HashPwd 对给定的密码进行bcrypt哈希处理并返回哈希后的字符串
// 参数:
//	pwd string - 需要哈希处理的原始密码
// 返回值:
//	string - 哈希后的密码字符串

func HashPwd(pwd string) string {
	// 使用bcrypt生成密码的哈希值，设置最小成本
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.MinCost)
	if err != nil {
		log.Println(err) // 记录无法生成哈希值的错误
	}
	return string(hash)
}

// CheckPwd 根据提供的哈希密码和原始密码验证密码是否匹配
// 参数:
//
//	hashPwd string - 存储的哈希密码字符串
//	pwd string - 需要验证的原始密码
//
// 返回值:
//
//	bool - 如果原始密码匹配哈希密码，则返回true，否则返回false
func CheckPwd(hashPwd, pwd string) bool {
	// 将哈希密码转换为字节数组，以供bcrypt比较使用
	byteHash := []byte(hashPwd)
	// 比较提供的密码和哈希密码是否匹配
	err := bcrypt.CompareHashAndPassword(byteHash, []byte(pwd))
	if err != nil {
		log.Println(err) // 记录密码比较失败的错误
		return false
	}
	return true
}
