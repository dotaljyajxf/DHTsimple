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

func InsertHash(hash string, from string, peerId string) {
	selector := bson.M{"hash": hash}
	updator := bson.M{"hash": hash, "addr": from, "peer_id": peerId}
	date := time.Now().Format("20060102")
	mdb.DB("info_hash").C(date).EnsureIndexKey("hash")
	_, err := mdb.DB("info_hash").C("hash_"+date).Upsert(selector, updator)
	if err != nil {
		fmt.Printf("mg insert error:%s\n", err.Error())
	}
}

type HashInfo struct {
	Id     bson.ObjectId `bson:"_id"`
	Hash   string        `bson:"hash"`
	Addr   string        `bson:"addr"`
	PeerId string        `bson:"peer_id"`
}

func GetHash(date string, beginId bson.ObjectId, limit int) ([]*HashInfo, error) {
	ret := make([]*HashInfo, 0)
	err := mdb.DB("info_hash").C("hash_" + date).Find(bson.M{"_id": bson.M{"$gt": beginId}}).Limit(limit).All(&ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
