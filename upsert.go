package main

import (
	"fmt"
	"time"

	"gopkg.in/mgo.v2/bson"

	"gopkg.in/mgo.v2"
)

const MONGODB_NAME = "test_device"
const MONGODB_COLLECTION = "device"

type VisitedDevice struct {
	ID        bson.ObjectId `bson:"_id"`
	Date      string        `bson:"date"`
	Uid       int64         `bson:"uid"`
	ProjectId int64         `bson:"project_id"`
	DeviceID  string        `bson:"device_id"`
	CreateAt  time.Time     `bson:"create_at"`
}

func main() {
	mdb, err := mgo.DialWithTimeout("62.234.136.238:27017", 3*time.Second)
	if err != nil {
		fmt.Printf("dail mgo err : %s\n", err.Error())
		return
	}
	v := VisitedDevice{Uid: 100, ProjectId: 121212121, DeviceID: "232323232"}
	v.Date = time.Now().Format("20060102")
	v.ID = bson.NewObjectId()
	v.CreateAt = time.Now().Local()

	selector := bson.M{"date": v.Date, "uid": v.Uid, "project_id": v.ProjectId, "device_id": v.DeviceID}
	_, err = mdb.DB(MONGODB_NAME).C(MONGODB_COLLECTION).Upsert(selector, v)
	if err != nil {
		fmt.Println("upsert error")
		return
	}

	dIndex := mgo.Index{Key: []string{"date", "project_id", "user_id"}, Background: true}
	expireIndex := mgo.Index{Key: []string{"create_at"}, Background: true, ExpireAfter: time.Hour}
	err = mdb.DB(MONGODB_NAME).C(MONGODB_COLLECTION).EnsureIndex(dIndex)
	if err != nil {
		fmt.Println("dindex error")
		return
	}
	err = mdb.DB(MONGODB_NAME).C(MONGODB_COLLECTION).EnsureIndex(expireIndex)
	if err != nil {
		fmt.Println("eindex error")
		return
	}

}
