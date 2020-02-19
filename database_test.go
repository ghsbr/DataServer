package main

import (
	"testing"
	"fmt"
	"os"
	"encoding/json"
)

var db Database

func TestNewDatabase(dc *testing.T){
	//t.Errror()
	db,_,err :=  NewDatabase(":memory:")
	if err != nil{
		dc.Errorf("Error found %v", err)
	}
}

func TestInsert(dc *testing.T){
	var jsonExample []byte
	{
		file, err := os.Open("data_example.json")
		if err != nil {
			println(err.Error())
			os.Exit(3)
		}
		if stat, err := file.Stat(); err == nil {
			jsonExample = make([]byte, stat.Size())
			file.Read(jsonExample)
		} else {
			println(err.Error())
			os.Exit(4)
		}
		file.Close()
	}
	if printDebug {
		println(string(jsonExample))
	}
}

func pTestInsert(dc *testing.T){
	var out Data
	if err := json.Unmarshal(jsonExample, &out); err != nil {
		println(err.Error())
		os.Exit(5)
	}
	if printDebug {
		fmt.Printf("%+v\n", out)
	}

	err := db.Insert(out)
	if  err != nil{
		dc.Errorf("Error found %v", err)
	} 
}

/*func TestPreciseQuery(dc *testing.T){
	//t.Errror()
	db,_,err :=  db.PreciseQuery(long float64, lat float64, day int64)
	if  err != nil{
		dc.Errorf("Error found %v", err)
	} 

	if err = db.Close(); err != nil{
		dc.Errorf("Error found %v", err)
	}
}*/

func TestClose(dc *testing.T){
	//t.Errror()
	if err = db.Close(); err != nil{
		dc.Errorf("Error found %v", err)
	}
}