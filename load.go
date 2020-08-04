package main

import (
	"DHTsimple/dht"
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
			fmt.Println("err : ", err.Error())
			break
		}

		if len(s) == 0 {
			break
		}

		for _, info := range s {
			fmt.Println("do addr: ", info.Addr)
			d := dht.NewMeta(info.Addr, []byte(info.Hash))
			d.Start()
		}
		beginId = s[len(s)-1].Id
		if len(s) < 100 {
			break
		}
		break
	}
}
