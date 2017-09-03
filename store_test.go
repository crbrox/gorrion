package gorrion

import (
	"testing"

	"gopkg.in/mgo.v2"
)

// Common functions for testing storage in mongoDB

var population []*Entity

func setupTestDB(t *testing.T) {
	var err error

	url = "localhost"
	db = "TEST_gorrion"
	entitiesColl = "TEST_ent"

	err = StartStore()
	if err != nil {
		t.Fatal(err)
	}

	err = initialSession.DB(db).C(entitiesColl).DropCollection()
	if err != nil {
		if mgoErr, ok := err.(*mgo.QueryError); ok {
			// 26 <-> ns not found => collection does not exist
			if mgoErr.Code != 26 {
				t.Fatal(err)
			}
		} else { // not a mgo.QueryError
			t.Fatal(err)
		}
	}
}

func teardownTestDB(t *testing.T) {
	StopStore()
}

func populateDB(t *testing.T) {

	var data = []struct {
		id          string
		t           string
		temperature float64
		status      string
	}{
		{"I1", "T1", 12.3, "ON"},
		{"I2", "T1", 22.3, "ON"},
		{"I3", "T1", 32.3, "ON"},
		{"E4", "T2", 42.3, "OFF"},
		{"E5", "T2", 52.3, "OFF"},
		{"E6", "T2", 62.3, "OFF"},
	}

	population = population[0:0]

	var emptyMap = map[string]interface{}{}
	for _, d := range data {
		e := &Entity{ID: EntityID{ID: d.id, Type: d.t, Service: "S", ServicePath: "SP"},
			Attrs: map[string]Attribute{
				"temperature": {Value: d.temperature, Type: "celsius", Md: emptyMap},
				"status":      {Value: d.status, Md: emptyMap},
			},
		}
		err := CreateEntity(e)
		if err != nil {
			t.Fatal(err)
		}
		population = append(population, e)
	}
}
