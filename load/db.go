package load

import (
	"fmt"

	"gopkg.in/mgo.v2/bson"

	"gopkg.in/mgo.v2"
)

var mdb *mgo.Session

const DB_NAME = "torrent"
const DB_COLLECTION = "torrent_hash"

func init() {
	//var err error
	//mdb, err = mgo.Dial(config.Conf.MongoUri)
	//if err != nil {
	//	fmt.Printf("dail mgo err : %s\n", err.Error())
	//	return
	//}
	//hashIndex := mgo.Index{
	//	Key:        []string{"hash"},
	//	Name:       "h_index",
	//	Unique:     true,
	//	Background: true,
	//}
	//mdb.DB(DB_NAME).C(DB_COLLECTION).EnsureIndex(hashIndex)
}

func InsertHash(t *Torrent) {
	selector := bson.M{"hash": t.HashHex}
	_, err := mdb.DB(DB_NAME).C(DB_COLLECTION).Upsert(selector, t)
	if err != nil {
		fmt.Printf("mg insert error:%s\n", err.Error())
	}
}
