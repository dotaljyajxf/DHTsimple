package dht

import "fmt"

func GetHash(hash string) {
	ret := ""
	str := "0123456789abcdef"
	for i := 0; i < 20; i++ {
		tmp := hash[i]
		ret += string(str[tmp>>4])
		ret += string(str[tmp&0xf])
	}
	ret += `\0`

	fmt.Println(ret)
}
