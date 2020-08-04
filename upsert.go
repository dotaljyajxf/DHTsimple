package main

import (
	"fmt"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"

	"gopkg.in/mgo.v2"
)

const MONGODB_NAME = "relative_reward"
const MONGODB_COLLECTION = "visited_device"

type VisitedDevice struct {
	ID        bson.ObjectId `bson:"_id,omitempty"`
	Date      string        `bson:"date"`
	Uid       int64         `bson:"uid"`
	ProjectId int64         `bson:"project_id"`
	DeviceID  string        `bson:"device_id"`
	CreateAt  time.Time     `bson:"create_at"`
}

func Insert(mdb *mgo.Session) {
	uidBeign := 100
	pidBeign := 200
	deviceBegin := "device_"
	for i := 0; i < 1000000; i++ {
		v := VisitedDevice{Uid: int64(uidBeign + i), ProjectId: int64(pidBeign + i), DeviceID: deviceBegin + strconv.Itoa(i)}
		v.Date = time.Now().Format("20060102")
		v.CreateAt = time.Now().Local()
		mdb.DB(MONGODB_NAME).C(MONGODB_COLLECTION).Insert(v)
	}
}

func main() {
	url := "mongodb://relative_reward:JwQuxn26uE$5VvM68ub&vy@dds-bp1d4e097a8e5b041594-pub.mongodb.rds.aliyuncs.com:3717,dds-bp1d4e097a8e5b042792-pub.mongodb.rds.aliyuncs.com:3717/relative_reward?replicaSet=mgset-5409953"
	mdb, err := mgo.DialWithTimeout(url, 3*time.Second)
	if err != nil {
		fmt.Printf("dail mgo err : %s\n", err.Error())
		return
	}

	Insert(mdb)

	//v := VisitedDevice{}
	//v.Uid = 100
	//v.ProjectId = 200
	//v.DeviceID = "test_device"
	//v.Date = time.Now().Format("20060102")
	//v.CreateAt = time.Now().Local()
	//mdb.DB(MONGODB_NAME).C(MONGODB_COLLECTION).Insert(v)
	//rr := make([]*VisitedDevice, 0)
	//err = mdb.DB(MONGODB_NAME).C(MONGODB_COLLECTION).Find(bson.M{}).All(&rr)
	//for i := 0; i < len(rr); i++ {
	//	fmt.Println(*rr[i])
	//}
	fmt.Println(mdb.DB(MONGODB_NAME).C(MONGODB_COLLECTION).Count())

	//v := VisitedDevice{Uid: 100, ProjectId: 121212121, DeviceID: "232323232"}
	//v.Date = time.Now().Format("20060102")
	//v.ID = bson.NewObjectId()
	//v.CreateAt = time.Now().Local()
	//
	//fmt.Println(len(v.ID))
	//fmt.Println(len(v.Date))
	//fmt.Println(len(v.CreateAt.String()))
	//
	//selector := bson.M{"date": v.Date, "uid": v.Uid, "project_id": v.ProjectId, "device_id": v.DeviceID}
	//_, err = mdb.DB(MONGODB_NAME).C(MONGODB_COLLECTION).Upsert(selector, v)
	//if err != nil {
	//	fmt.Println("upsert error")
	//	return
	//}
	//
	//dIndex := mgo.Index{Key: []string{"date", "project_id", "user_id"}, Background: true}
	//expireIndex := mgo.Index{Key: []string{"create_at"}, Background: true, ExpireAfter: time.Hour}
	//err = mdb.DB(MONGODB_NAME).C(MONGODB_COLLECTION).EnsureIndex(dIndex)
	//if err != nil {
	//	fmt.Println("dindex error")
	//	return
	//}
	//err = mdb.DB(MONGODB_NAME).C(MONGODB_COLLECTION).EnsureIndex(expireIndex)
	//if err != nil {
	//	fmt.Println("eindex error")
	//	return
	//}

}
