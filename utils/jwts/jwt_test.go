package jwts

import (
	"fmt"
	"testing"
)

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken(JwtPayLoad{
		UserID:   1,
		Role:     1,
		NickName: "Barton"}, "123456", 8)
	fmt.Println(token, err)
}
func TestParseToken(t *testing.T) {
	payload, err := ParseToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyaWQiOjEsIm5pY2tuYW1lIjoiQmFydG9uIiwicm9sZSI6MSwiZXhwIjoxNzE2OTI5MTEwfQ.DR3gbDACFD6eNNBdIKtKUJcAEypSVVjZO9Ffh9KAZhg", "123456")
	fmt.Println(payload, err)

}
