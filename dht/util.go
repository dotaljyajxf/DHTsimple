package dht

import (
	"crypto/rand"
	"fmt"
)

func RandString(num int) string {
	b := make([]byte, num)
	n, err := rand.Read(b)
	if err != nil {
		fmt.Printf("rand id err:%s\n", err.Error())
		panic(err)
	}
	if n != num {
		fmt.Printf("rand id len error :%d\n", n)
		panic(err)
	}
	return string(b)
}

func MakeRequest() {

}

func MakeResponse() {

}
