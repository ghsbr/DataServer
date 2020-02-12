package main

import (
	"github.com/bvinc/go-sqlite-lite/sqlite3"
)

type Database struct{
	conn *sqlite3.Conn

}

func newDatabase(string file) (Database, error){
	
	conn, err := sqlite3.Open(file)
	performOneTimeSetup(conn)

	if err == nil {
		return Database{conn}, nil
	}
	else{
		return Database{nil}, err
	}
}

func performOneTimeSetup(db *sqlite3.Conn) (bool, error) {
	//Prepariamo la query SQL
	stmt, err := db.Prepare("SELECT name FROM sqlite_master WHERE type='table' AND name='long0'")
	if err != nil {
		return false, err
	}

	/*Eseguiamo la query e controlliamo exists per controllare se una riga
	 *effettivamente esiste. In tal caso non agiremo e ritorneremo false
	 *in caso contrario procederemo a creare le tabelle e a ritornare true*/
	exists, err := stmt.Step()
	stmt.Close()
	if err != nil {
		return false, err
	} else if exists {
		return false, nil
	}

	//Creo una tabella ogni 5 "gradi"
	{
		var i int64
		for i = 0; i < 360; i += 5 {
			if printDebug {
				println("Creating: long" + strconv.FormatInt(i, 10) + " for " + strconv.FormatInt(i-180, 10))
			}
			err = db.Exec("CREATE TABLE long" + strconv.FormatInt(i, 10) + ` (
	long REAL,
	lat REAL,
	idx INTEGER,
	PRIMARY KEY (long, lat)
)`)
			if err != nil {
				return true, err
			}
		}
	}

	//Ottengo l'unix timestamp per la giornata di oggi
	today := time.Now().UTC().Truncate(time.Duration(time.Hour * 24)).Unix()
	if printDebug {
		println(time.Unix(today, 0).Format(time.UnixDate))
		println("Creating: d" + strconv.FormatInt(today, 10))
	}

	//Creo una tabella per la giornata in corso
	err = db.Exec("CREATE TABLE d" + strconv.FormatInt(today, 10) + ` (
	idx INTEGER,
	time INTEGER,
	pm25_concentration REAL,
	temperature INTEGER,
	PRIMARY KEY(idx, time)
)`)
	if err != nil {
		return true, err
	}
	return true, nil
}