package dht

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"

	"gopkg.in/mgo.v2"
)

var mdb *mgo.Session

func init() {
	var err error
	mdb, err = mgo.DialWithTimeout("62.234.136.238:27017", 3*time.Second)
	if err != nil {
		fmt.Printf("dail mgo err : %s\n", err.Error())
		return
	}

}

func InsertHash(hash string, from string) {
	selector := bson.M{"hash": hash}
	updator := bson.M{"hash": hash, "addr": from}
	date := time.Now().Format("20060102")
	mdb.DB("info_hash").C(date).EnsureIndexKey("hash")
	_, err := mdb.DB("info_hash").C("hash_"+date).Upsert(selector, updator)
	if err != nil {
		fmt.Printf("mg insert error:%s\n", err.Error())
	}
}

type HashInfo struct {
	Id   bson.ObjectId `bson:"_id"`
	Hash string        `bson:"hash"`
	Addr string        `bson:"addr"`
}

func GetHash(date string, beginId bson.ObjectId, limit int) ([]*HashInfo, error) {

	fmt.Println("222")
	ret := make([]*HashInfo, 0)
	var err error
	if len(beginId) == 0 {
		err = mdb.DB("info_hash").C("hash_" + date).Find(bson.M{}).Sort("_id").Limit(limit).All(&ret)
	} else {
		err = mdb.DB("info_hash").C("hash_" + date).Find(bson.M{"_id": bson.M{"$gt": beginId}}).Limit(limit).All(&ret)
	}
	fmt.Println("111")
	if err != nil {
		return nil, err
	}
	return ret, nil
}
