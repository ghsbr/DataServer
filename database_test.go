package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/ghsbr/DataServer/data"
	"github.com/ghsbr/DataServer/database"
)

var db *database.Database

func TestNewDatabase(dc *testing.T) {
	var err error
	db, _, err = database.NewDatabase(":memory:", log.New(os.Stdout, "[DataServer] ", 0))
	if err != nil {
		dc.Errorf("Error while creating Database: %v", err)
	}
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
	els, err := db.ApproximateQuery(out.Longitude, out.Latitude, out.Ts, 1)
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

func TestMain(dc *testing.T) {
	const addr string = "127.0.0.1:8080"
	dc.Run(
		"Insertion test",
		func(test *testing.T) {
			resp, err := http.Post("http://"+addr+"/insert", "application/json", bytes.NewReader(jsonExample))
			if err != nil {
				dc.Fatalf("%v", err)
			} else {
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					dc.Fatalf("%v", err)
				} else {
					dc.Log(string(body))
				}
			}
		},
	)

	form := url.Values(make(map[string][]string))
	form.Set("long", fmt.Sprintf("%v", out.Longitude))
	form.Set("lat", fmt.Sprintf("%v", out.Latitude))
	form.Set("day", fmt.Sprintf("%v", out.Ts))
	dc.Run(
		"PreciseQuery test",
		func(test *testing.T) {
			resp, err := http.PostForm("http://"+addr+"/query", form)
			if err != nil {
				dc.Errorf("%v", err)
			} else {
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					dc.Errorf("%v", err)
				} else {
					dc.Log(string(body))
				}
			}
		},
	)

	form.Set("range", "1")
	dc.Run(
		"ApproximateQuery In Range test",
		func(test *testing.T) {
			resp, err := http.PostForm("http://"+addr+"/query", form)
			if err != nil {
				dc.Errorf("%v", err)
			} else {
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					dc.Errorf("%v", err)
				} else {
					dc.Log(string(body))
				}
			}
		},
	)

	form.Set("range", "100")
	dc.Run(
		"ApproximateQuery Out of Range test",
		func(test *testing.T) {
			resp, err := http.PostForm("http://"+addr+"/query", form)
			if err != nil {
				dc.Errorf("%v", err)
			} else {
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					dc.Errorf("%v", err)
				} else {
					dc.Log(string(body))
				}
			}
		},
	)
}
