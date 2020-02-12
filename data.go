package main

import (
	"encoding/json"
	"fmt"
	"time"
)

/*Data Contiene I dati di:
=> TS - Orario in cui è stata fatta la misurazione
=> Pm25 -  Indice inquinamento polveri sottili
		=> Indice sulla qualità dell'aria generalizzata
		=> Indice di concentrazione delle polveri sottili nell'aria
=> Condition - Condizione meteorologica al momento della misurazione
=> Humidity - Livello di Umidità al momento della misura
=> Pressure - Livello di pressione al momento della misurazione
=> Wind - Informazioni sul vento al momento della misurazione
	=> Speed - Velocità del vento
	=> Direction - Direzione del vento
=> Temperature - Temperatura al momento della misurazione
=> Latitude - Latitudine di riferimento per la misurazione corrente
=> Longitude - Longitudine di riferimento per la misurazione corrente
*/
type Data struct {
	Ts          int64
	Pm25Aqi     int64
	Temperature int64
	Latitude    float64
	Longitude   float64
}

/*UnmarshalJSON :
Il programma carica il JSON e ritorna un errore nel caso in cui ci isano problemi, altrimenti lo stampa
Successivamente corregge il formato del tempo e ritorna un errore nel caso di problemi
Alla fine gli viene detto dove andare ad inserire i dati che troverà nel JSON
*/
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

/*
	Al programma viene detto dove andare a prendere i dati richiesti dentro il file che è stato caricato
	e dove andare poi a mettere la relativa copia
*/
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
