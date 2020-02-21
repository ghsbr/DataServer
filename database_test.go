package main

import (
	"encoding/json"
	"github.com/ghsbr/DataServer/data"
	"github.com/ghsbr/DataServer/database"
	"os"
	"testing"
)

var db database.Database

func TestNewDatabase(dc *testing.T) {
	var err error
	db, _, err = database.NewDatabase(":memory:")
	if err != nil {
		dc.Errorf("Error while creating Database: %v", err)
	}
}

var out data.Data

func TestInsert(dc *testing.T) {
	var jsonExample []byte
	{
		file, err := os.Open("data_example.json")
		if err != nil {
			dc.Errorf("Error while opening data_example: %v", err)
		}
		if stat, err := file.Stat(); err == nil {
			jsonExample = make([]byte, stat.Size())
			file.Read(jsonExample)
		} else {
			dc.Errorf("Error while reading File stats: %v", err)
		}
		file.Close()
	}
	if err := json.Unmarshal(jsonExample, &out); err != nil {
		dc.Errorf("Error while parsing json: %v", err)
	}

	err := db.Insert(out)
	if err != nil {
		dc.Errorf("Error found %v", err)
	}
}

func TestPreciseQuery(dc *testing.T) {
	el, err := db.PreciseQuery(out.Longitude, out.Latitude, out.Ts)
	if err != nil {
		dc.Errorf("Error found %v", err)
	} else {
		dc.Logf("PreciseQuery returned: %+v", el)
	}
}

func TestApproximateQueryInRange(dc *testing.T) {
	els, err := db.ApproximateQuery(out.Longitude, out.Latitude, out.Ts, 3)
	if err != nil {
		dc.Errorf("Error found %v", err)
	} else if len(els) > 0 {
		dc.Logf("ApproximateQuery returned: %+v", els)
	} else {
		dc.Error("ApproximateQuery didn't return anything")
	}
}

func TestApproximateQueryOutOfRange(dc *testing.T) {
	els, err := db.ApproximateQuery(out.Longitude-50, out.Latitude+50, out.Ts, 100)
	if err != nil {
		dc.Errorf("Error found %v", err)
	} else if len(els) > 0 {
		dc.Logf("ApproximateQuery returned: %+v", els)
	} else {
		dc.Error("ApproximateQuery didn't return anything")
	}
}

func TestClose(dc *testing.T) {
	if err := db.Close(); err != nil {
		dc.Errorf("Error found %v", err)
	}
}
