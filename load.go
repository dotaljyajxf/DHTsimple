package main

import (
	"dhtTest/dht"
	"time"

	"gopkg.in/mgo.v2/bson"
)

func main() {
	//date := time.Now().Add(-24 * time.Hour).Format("20060102")
	date := time.Now().Format("20060102")

	beginId := bson.ObjectIdHex("0")

	for {
		s, err := dht.GetHash(date, beginId, 100)
		if err != nil {
			continue
		}

		for _, info := range s {
			d := dht.NewMeta(info.PeerId, info.Addr, []byte(info.Hash))
			d.Start()
		}

		if len(s) < 100 {
			break
		}
	}
}
