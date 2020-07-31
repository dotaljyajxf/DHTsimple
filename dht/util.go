package dht

import (
	"crypto/rand"
	"fmt"
	"sync"
)

type ByteBuf []byte

var BytePool = &sync.Pool{
	New: func() interface{} {
		return make(ByteBuf, 256)
	},
}

func NewBufferByte() ByteBuf {
	return BytePool.Get().(ByteBuf)
}

func (b ByteBuf) Release() {
	BytePool.Put(b)
}

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

func neighborId(nodeId string, target string) string {
	head := target[0:15]
	tail := nodeId[15:]
	return head + tail
}

func MakeRequest(method string, nodeId string, target string) map[string]interface{} {
	ret := make(map[string]interface{})
	ret["t"] = RandString(2)
	ret["y"] = 'q'
	ret["q"] = method
	ret["a"] = map[string]interface{}{"id": neighborId(nodeId, target), "target": RandString(20)}

	return ret
}

func MakeResponse(r map[string]interface{}) map[string]interface{} {
	ret := make(map[string]interface{})
	ret["t"] = RandString(2)
	ret["y"] = 'r'
	ret["r"] = r

	return ret
}
