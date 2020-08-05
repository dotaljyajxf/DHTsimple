package dht

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"

	"gopkg.in/mgo.v2"
)

var mdb *mgo.Session

const DB_NAME = "torrent"
const DB_COLLECTION = "torrent_hash"

func init() {
	var err error
	url := "mongodb://hash:ljy1314@62.234.136.238/torrent"
	mdb, err = mgo.DialWithTimeout(url, 3*time.Second)
	if err != nil {
		fmt.Printf("dail mgo err : %s\n", err.Error())
		return
	}
	hashIndex := mgo.Index{
		Key:        []string{"hash"},
		Name:       "h_index",
		Unique:     true,
		Background: true,
	}
	mdb.DB(DB_NAME).C(DB_COLLECTION).EnsureIndex(hashIndex)
}

func InsertHash(t *Torrent) {
	selector := bson.M{"hash": t.infohashHex}
	_, err := mdb.DB(DB_NAME).C(DB_COLLECTION).Upsert(selector, t)
	if err != nil {
		fmt.Printf("mg insert error:%s\n", err.Error())
	}
}

//func GetHash(date string, beginId bson.ObjectId, limit int) ([]*HashInfo, error) {
//
//	fmt.Println("222")
//	ret := make([]*HashInfo, 0)
//	var err error
//	if len(beginId) == 0 {
//		err = mdb.DB("info_hash").C("hash_" + date).Find(bson.M{}).Sort("_id").Limit(limit).All(&ret)
//	} else {
//		err = mdb.DB("info_hash").C("hash_" + date).Find(bson.M{"_id": bson.M{"$gt": beginId}}).Limit(limit).All(&ret)
//	}
//	fmt.Println("111")
//	if err != nil {
//		return nil, err
//	}
//	return ret, nil
//}
