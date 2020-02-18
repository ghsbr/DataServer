package main

import (
	"testing"
)
func TestNewDatabase(dc *testing.T){
	//t.Errror()
	db,_,err :=  NewDatabase(":memory:")
	if err != nil{
		dc.Errorf("Error found %v", err)
	}
	
	if err = db.Close(); err != nil{
		dc.Errorf("Error found %v", err)
	}

}
func TestPreciseQuery(dc *testing.T){
	//t.Errror()
	db,_,err :=  PreciseQuery(":memory:")
	if  err != nil{
		dc.Errorf("Error found %v", err)
	} 

	if err = db.Close(); err != nil{
		dc.Errorf("Error found %v", err)
	}
}