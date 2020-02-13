package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bvinc/go-sqlite-lite/sqlite3"
)

//uno struct che fa da man-in-the-middle per il database
type Database struct {
	conn *sqlite3.Conn
}

//costruttore della classe Database
//ritorna un errore, se esiste, e un oggetto database
func newDatabase(file string) (Database, bool, error) {
	conn, err := sqlite3.Open(file)
	if err != nil {
		return Database{nil}, false, err
	}

	mod, err := performOneTimeSetup(conn)
	if err != nil {
		conn.Close()
		return Database{nil}, false, err
	}

	return Database{conn}, mod, err
}

func (db Database) Insert(data Data) error {
	stmt, err := db.conn.Prepare(
		"SELECT idx FROM long"+fmt.Sprintf("%v", getLongIdx(data.Longitude))+" WHERE long=? AND lat=?",
		data.Longitude, data.Latitude,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	next, err := stmt.Step()
	if err != nil {
		return err
	}

	var id int64
	if next {
		err = stmt.Scan(&id)
		if err != nil {
			return err
		}
	} else {
		err = db.conn.Exec("INSERT INTO ids VALUES (NULL)")
		if err != nil {
			return err
		}

		id = db.conn.LastInsertRowID()
		err = db.conn.Exec(
			"INSERT INTO long"+fmt.Sprintf("%v", getLongIdx(data.Longitude))+" VALUES (?, ?, ?)",
			data.Longitude, data.Latitude, id,
		)
		if err != nil {
			return err
		}
	}

	today := time.Now().UTC().Truncate(time.Duration(time.Hour * 24)).Unix()
	err = db.conn.Exec(
		"INSERT OR REPLACE INTO d"+fmt.Sprintf("%v", today)+" VALUES (?, ?, ?, ?)",
		id, data.Ts, data.Pm25Aqi, data.Temperature,
	)
	return err
}

func (db Database) Close() error {
	return db.conn.Close()
}

func performOneTimeSetup(db *sqlite3.Conn) (bool, error) {
	//Prepariamo la query SQL
	stmt, err := db.Prepare("SELECT name FROM sqlite_master WHERE type='table' AND name='ids'")
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

	//Creo una tabella per generare
	db.Exec(`CREATE TABLE ids (
	id INTEGER PRIMARY KEY ASC AUTOINCREMENT
)`)

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

func getLongIdx(long float64) int64 {
	posIdx := int64(long + 180)
	return posIdx - posIdx%5
}
