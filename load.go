package main

import (
	"dhtTest/dht"
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"
)

func main() {
	//date := time.Now().Add(-24 * time.Hour).Format("20060102")
	date := time.Now().Format("20060102")

	var beginId bson.ObjectId

	for {
		s, err := dht.GetHash(date, beginId, 100)
		if err != nil {
			continue
		}

		for _, info := range s {
			fmt.Println("do addr: ", info.Addr)
			d := dht.NewMeta(info.PeerId, info.Addr, []byte(info.Hash))
			d.Start()
		}
		beginId = s[len(s)-1].Id
		if len(s) < 100 {
			break
		}
		break
	}
}
