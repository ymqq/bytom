package main

import (
	"encoding/hex"
	"golang.org/x/crypto/sha3"
	"os"
	"fmt"
)

func main() {
	if (len(os.Args) != 2) {
		fmt.Println("command args: [data]")
		os.Exit(0)
	}

	data, err := hex.DecodeString(os.Args[1])
	if err!= nil{
		fmt.Println("DecodeString err:", err)
		os.Exit(0)
	}
	fmt.Println("data string:", hex.EncodeToString(data))
	hash := sha3.Sum256(data)

	var hashvalue []byte
	for _, s := range hash{
		hashvalue = append(hashvalue, s)
	}
	fmt.Println("hash result:", hex.EncodeToString(hashvalue))

}