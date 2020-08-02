package dht

import (
	"encoding/hex"
	"fmt"
)

func GetHash(hash string, from string, id string) {
	ret := ""
	str := "0123456789abcdef"
	for i := 0; i < 20; i++ {
		tmp := hash[i]
		ret += string(str[tmp>>4])
		ret += string(str[tmp&0xf])
	}
	ret += "\\0"

	fmt.Println(from, " ", ret)
	fmt.Println(from, "_HEX ", hex.EncodeToString([]byte(hash)))
}
