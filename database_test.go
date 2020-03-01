package main

import (
	"encoding/json"
	"log"
	"os"
	"testing"

	"github.com/ghsbr/DataServer/data"
	"github.com/ghsbr/DataServer/database"
)

var db *database.Database

func TestNewDatabase(dc *testing.T) {
	var err error
	var perf bool
	db, perf, err = database.NewDatabase("dataserver", "dataserver", "dataserver", log.New(os.Stdout, "[DataServer] ", 0))
	if err != nil {
		dc.Fatalf("Error while creating Database: %v", err)
	}
	dc.Logf("Setup performed? %v", perf)
}

var out data.Data
var jsonExample []byte

func TestInsert(dc *testing.T) {
	{
		file, err := os.Open("data_example.json")
		if err != nil {
			dc.Errorf("Error while opening data_example: %v", err)
		}
		if stat, err := file.Stat(); err == nil {
			jsonExample = make([]byte, stat.Size())
			file.Read(jsonExample)
		} else {
			file.Close()
			dc.Errorf("Error while reading File stats: %v", err)
		}
		file.Close()
	}
	if err := json.Unmarshal(jsonExample, &out); err != nil {
		dc.Errorf("Error while parsing json: %v", err)
	}

	err := db.Insert(out)
	if err != nil {
		dc.Fatalf("Error found %v", err)
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
	els, err := db.ApproximateQuery(out.Longitude, out.Latitude, out.Ts, .4174)
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
