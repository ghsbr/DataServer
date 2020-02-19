package main

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

var db Database

func TestNewDatabase(dc *testing.T) {
	db, _, err := NewDatabase(":memory:")
	if err != nil {
		dc.Errorf("Error while creating Database: %v", err)
	}
}

var out Data

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
	el, err := db.PreciseQuery(out.Longitude, out.Latitude, time.Now().UTC().Truncate(time.Hour*24).Unix())
	if err != nil {
		dc.Errorf("Error found %v", err)
	} else {
		dc.Logf("PreciseQuery returned: %+v", el)
	}
}

func TestClose(dc *testing.T) {
	if err := db.Close(); err != nil {
		dc.Errorf("Error found %v", err)
	}
}
