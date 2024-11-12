package pwd

import (
	"fmt"
	"testing"
)

func TestHashPwd(t *testing.T) {
	hash := HashPwd("123456")
	fmt.Println(hash)
}
func TestCheckPwd(t *testing.T) {
	ok := CheckPwd("$2a$04$Wy49QmOJaRUFT6L0hIAFIeVArGbzu6aqgO1RxDCd5yW11B/RwbPF6", "123456")
	fmt.Println(ok)
	ok = CheckPwd("$2a$04$Wy49QmOJaRUFT6L0hIAFIeVArGbzu6aqgO1RxDCd5yW11B/RwbPF6", "1234567")
	fmt.Println(ok)
}
