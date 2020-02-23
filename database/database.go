package database

import (
	"fmt"
	//"log"
	"math"
	"sync"
	"time"

	"github.com/bvinc/go-sqlite-lite/sqlite3"
	"github.com/ghsbr/DataServer/data"
	//"github.com/ghsbr/DataServer"
	//"github.com/cespare/xxhash"
)

const longPerTable = 5

type Data = data.Data

//uno struct che fa da man-in-the-middle per il database
type Database struct {
	conn *sqlite3.Conn
	lock sync.RWMutex
}

//costruttore della classe Database
//ritorna un errore, se esiste, e un oggetto database
func NewDatabase(file string) (Database, bool, error) {
	conn, err := sqlite3.Open(file)
	if err != nil {
		return Database{nil, sync.RWMutex{}}, false, err
	}

	mod, err := performOneTimeSetup(conn)
	if err != nil {
		conn.Close()
		return Database{nil, sync.RWMutex{}}, false, err
	}

	return Database{conn, sync.RWMutex{}}, mod, err
}

func (db *Database) PreciseQuery(long float64, lat float64, day int64) (Data, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()
	idx, err := (func() (int64, error) {
		stmt, err := db.conn.Prepare(
			"SELECT idx FROM long"+fmt.Sprintf("%v", getIndexFromLongitude(long))+" WHERE long=? AND lat=?",
			long, lat,
		)
		defer stmt.Close()
		if err != nil {
			return 0, err
		}

		next, err := stmt.Step()
		if err != nil {
			return 0, err
		}
		if !next {
			return 0, NotFoundError{"Station not Found"}
		}

		var idx int64
		err = stmt.Scan(&idx)
		return idx, err
	})()
	if err != nil {
		return Data{}, err
	}

	stmt, err := db.conn.Prepare(
		"SELECT time,long,lat,pm25_concentration,temperature FROM d"+fmt.Sprintf("%v", truncateTime(day))+" WHERE idx=?",
		idx,
	)
	if err != nil {
		return Data{}, err
	}
	defer stmt.Close()

	next, err := stmt.Step()
	if err != nil {
		return Data{}, err
	}
	if !next {
		return Data{}, NotFoundError{"Station not Found"}
	}

	var ret Data
	err = stmt.Scan(&ret.Ts, &ret.Longitude, &ret.Latitude, &ret.Pm25Aqi, &ret.Temperature)
	if err != nil {
		return Data{}, err
	}

	return ret, nil
}

func (db *Database) ApproximateQuery(long float64, lat float64, day int64, rng float64) ([]Data, error) {
	//log.Printf("%v %v\t%v %v\n", long, rng, long-rng, long+rng)
	day = truncateTime(day)
	if getIndexFromLongitude(long-rng) == getIndexFromLongitude(long+rng) {
		return db.actualApproximateQuery(long-rng, long+rng, lat, rng, day)
	} else {
		lowerLimit := math.Max(long-rng, -180)
		//log.Printf("%v %v\n", lowerLimit, 180)
		ret, err := db.actualApproximateQuery(
			lowerLimit,
			180,
			lat,
			longPerTable,
			day,
		)
		if err != nil {
			return nil, err
		}

		upperLimit := long + rng
		//i: Indice alla tabella dopo
		for i := float64(getIndexFromLongitude(lowerLimit) - 180 + longPerTable); i <= upperLimit && i < 180; i += longPerTable {
			//log.Printf("%v %v\n", i, i+longPerTable)
			part, err := db.actualApproximateQuery(i, math.Min(i+longPerTable, upperLimit), lat, rng, day)
			if err != nil {
				return nil, err
			}

			ret = append(ret, part...)
		}
		return ret, nil
	}
}

func (db *Database) actualApproximateQuery(longMin float64, longMax float64, lat float64, latrng float64, day int64) ([]Data, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()
	idxs, err := (func() ([]int64, error) {
		stmt, err := db.conn.Prepare(
			"SELECT idx FROM long"+fmt.Sprintf("%v", getIndexFromLongitude(longMin))+" WHERE long>=? AND long<=? AND lat>=? AND lat<=?",
			longMin, longMax, lat-latrng, lat+latrng,
		)
		if err != nil {
			return nil, err
		}
		defer stmt.Close()

		idxs := make([]int64, 0)
		for next, err := stmt.Step(); next && err == nil; next, err = stmt.Step() {
			idxs = append(idxs, 0)
			err = stmt.Scan(&idxs[len(idxs)-1])
			if err != nil {
				break
			}
		}

		return idxs, nil
	})()
	if err != nil {
		return nil, err
	}
	if len(idxs) == 0 {
		return nil, nil
	}

	//log.Printf("%v", idxs)
	stmt, err := db.conn.Prepare(
		"SELECT time,long,lat,pm25_concentration,temperature FROM d" + fmt.Sprintf("%v", day) + " WHERE idx=?",
	)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	ret := make([]Data, 0)
	for _, idx := range idxs {
		err = stmt.Bind(idx)
		if err != nil {
			return nil, err
		}

		next, err := stmt.Step()
		if err != nil {
			return nil, err
		}
		if !next {
			err = NotFoundError{fmt.Sprintf("Station %v not found", idx)}
			return nil, err
			//continue
		}

		var station Data
		err = stmt.Scan(&station.Ts, &station.Longitude, &station.Latitude, &station.Pm25Aqi, &station.Temperature)
		if err != nil {
			return nil, err
		}
		ret = append(ret, station)

		err = stmt.Reset()
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (db *Database) Insert(data Data) error {
	db.lock.Lock()
	defer db.lock.Unlock()
	stmt, err := db.conn.Prepare(
		"SELECT idx FROM long"+fmt.Sprintf("%v", getIndexFromLongitude(data.Longitude))+" WHERE long=? AND lat=?",
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
			"INSERT INTO long"+fmt.Sprintf("%v", getIndexFromLongitude(data.Longitude))+" VALUES (?, ?, ?)",
			data.Longitude, data.Latitude, id,
		)
		if err != nil {
			return err
		}
	}

	err = db.conn.Exec(
		"INSERT OR REPLACE INTO d"+fmt.Sprintf("%v", truncateTime(data.Ts))+" VALUES (?, ?, ?, ?, ?, ?)",
		id, data.Ts, data.Longitude, data.Latitude, data.Pm25Aqi, data.Temperature,
	)
	if errv, ok := err.(*sqlite3.Error); ok && errv.Code() == 1 {
		createTimeNamedTable(db, data.Ts)
		err = db.conn.Exec(
			"INSERT OR REPLACE INTO d"+fmt.Sprintf("%v", truncateTime(data.Ts))+" VALUES (?, ?, ?, ?, ?, ?)",
			id, data.Ts, data.Longitude, data.Latitude, data.Pm25Aqi, data.Temperature,
		)
	}
	return err
}

func (db *Database) Close() error {
	db.lock.Lock()
	defer db.lock.Unlock()
	return db.conn.Close()
}

func createTimeNamedTable(db *Database, time int64) error {
	//Creo una tabella per la giornata in corso
	db.conn.Exec("CREATE TABLE d" + fmt.Sprintf("%v", truncateTime(time)) + ` (
	idx INTEGER,
	time INTEGER,
	long REAL,
	lat REAL,
	pm25_concentration REAL,
	temperature INTEGER,
	PRIMARY KEY(idx, time)
)`)
	return nil
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
		for i := 0; i < 360; i += 5 {
			err = db.Exec("CREATE TABLE long" + fmt.Sprintf("%v", i) + ` (
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
	return true, nil
}

func getIndexFromLongitude(long float64) int64 {
	posIdx := int64(math.Trunc(long + 180))
	return posIdx - posIdx%longPerTable
}

func truncateTime(timestamp int64) int64 {
	return time.Unix(timestamp, 0).Truncate(time.Hour * 24).Unix()
}

type NotFoundError struct {
	msg string
}

func (e NotFoundError) Error() string {
	return e.msg
}
