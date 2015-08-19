package service

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"models"
)

func QueryPage(dbUri string, pageNo int, pageSize int) []models.Model {
	session, err := mgo.Dial(dbUri)

	defer session.Close()

	m := session.DB("meikong").C("model")

	var results []models.Model

	start := (pageNo - 1) * pageSize
	end := pageNo * pageSize

	err = m.Find(bson.M{"number": bson.M{"$gt": start, "$lte": end}}).Sort("number").All(&results)

	if err != nil {
		panic(err)
	}

	return results
}
