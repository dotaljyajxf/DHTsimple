package main

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2"
)

func main() {
	var err error
	url := "mongodb://hash:ljy1314@62.234.136.238/torrent"
	mdb, err := mgo.DialWithTimeout(url, 3*time.Second)
	if err != nil {
		fmt.Printf("dail mgo err : %s\n", err.Error())
		return
	}
	fmt.Println(mdb)
	//date := time.Now().Add(-24 * time.Hour).Format("20060102")
	//date := time.Now().Format("20060102")
	//
	//var beginId bson.ObjectId
	//
	//for {
	//	s, err := dht.GetHash(date, beginId, 100)
	//	if err != nil {
	//		fmt.Println("err : ", err.Error())
	//		break
	//	}
	//
	//	if len(s) == 0 {
	//		break
	//	}
	//
	//	for _, info := range s {
	//		fmt.Println("do addr: ", info.Addr)
	//		d := dht.NewMeta(info.Addr, []byte(info.Hash))
	//		d.Start()
	//	}
	//	beginId = s[len(s)-1].Id
	//	if len(s) < 100 {
	//		break
	//	}
	//	break
	//}
}
