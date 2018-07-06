package api

import (
	"crypto/sha256"

	"golang.org/x/crypto/sha3"

	//chainjson "github.com/bytom/encoding/json"
	"fmt"
	"encoding/hex"
	"testing"
)

func TestHex(t *testing.T) {
	hash := sha3.Sum256([]byte("9963265eb601df48501cc240e1480780e9ed6e0c8f18fd7dd57954068c5dfd02"+"4c97d7412b04d49acc33762fc748cd0780d8b44086c229c1a6d0f2adfaaac2db"))
	fmt.Println("-------origin:", hex.EncodeToString([]byte("hello")))
	fmt.Println("-------sha3 hash:", hex.EncodeToString(hash[:]))

	hash = sha256.Sum256([]byte("hello"))
	fmt.Println("-------sha256 hash:", hex.EncodeToString(hash[:]))
}