package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// Generated by https://quicktype.io

/*
Contiene I dati di:
	=> TS - Orario in cui è stata fatta la misurazione
	=> Pm25 -  Indice inquinamento polveri sottili
			=> Indice sulla qualità dell'aria generalizzata
			=> Indice di concentrazione delle polveri sottili nell'aria
	=> Condition - Condizione meteorologica al momento della misurazione
	=> Humidity - Livello di Umidità al momento della misura
*/
type Data struct {
	Ts          int64
	Pm25Aqi     int64
	Temperature int64
	Latitude    float64
	Longitude   float64
}

func (out *Data) UnmarshalJSON(toParse []byte) error {
	var dataMap data
	if err := json.Unmarshal(toParse, &dataMap); err != nil {
		return err
	}
	fmt.Printf("%+v\n", dataMap)

	timeobj, err := time.Parse(time.RFC3339, dataMap.Ts)
	if err != nil {
		return err
	}

	*out = Data{
		timeobj.UTC().Unix(),
		dataMap.Pm25.Aqi,
		dataMap.Temperature,
		dataMap.Coordinates.Latitude,
		dataMap.Coordinates.Longitude,
	}
	return nil
}

type data struct {
	Ts   string `json:"ts"`
	Pm25 struct {
		Aqi int64 `json:"aqi"`
	} `json:"pm25"`
	Temperature int64 `json:"temperature"`
	Coordinates struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"coordinates"`
}
