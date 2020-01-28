package main

import (
	"fmt"
	"os"
	"encoding/json"
)

func main() {
	// open The requested Json File and controlif we have some troubles
	jsonFile, err := os.Open("data_example.json")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Opened jsonFile")
	// defer the closing of the file, so we can parse it later
	defer jsonFile.Close()
	type Italia struct
	{
		Italia[] Veneto 'json:"data_Example"'
	}
	type Veneto struct
	{
		Time string 'json:"ts"'
	}
}
