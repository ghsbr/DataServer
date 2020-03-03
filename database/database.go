package database

import (
	"database/sql"
	"fmt"
	"log"
	"math"

	"github.com/cespare/xxhash"
	"github.com/ghsbr/DataServer/data"
	"github.com/lib/pq"
)

const longPerTable = 5

type Data = data.Data

var (
	Log    *log.Logger
	hasher = xxhash.New()
)
//setta il logger per questo pacchetto.
func setLogger(mainLog *log.Logger) {
	Log = log.New(mainLog.Writer(), "[DataServer/Database] ", mainLog.Flags())
}

//uno struct che fa da man-in-the-middle per il database.
type Database struct {
	conn *sql.DB
}

//costruttore della classe Database
//ritorna un errore, se esiste, e un oggetto database.
func NewDatabase(user, password, dbname string, mainLog *log.Logger) (*Database, bool, error) {
	//Imposta il logger
	setLogger(mainLog)
	//Fa una prova di autenticazione nel database, tenendo aperta la connessione nel caso in cui vada
	//a buon fine.
	conn, err := sql.Open(
		"postgres",
		fmt.Sprintf("user=%v password=%v dbname=%v host=/run/postgresql", user, password, dbname),
	)
	if err != nil {
		return nil, false, err
	}
	err = conn.Ping()
	if err != nil {
		return nil, false, err
	}

	//Controlla se il setup del database è gia stato effettuato, in caso negativo viene eseguito.
	mod, err := performOneTimeSetup(conn)
	if err != nil {
		conn.Close()
		return nil, false, err
	}

	return &Database{conn}, mod, err
}
// Controlla che la stazione a quelle determinate coordinate, longitudine e latitudine, esista realmente.
//Prende tutti i parametri passati e controlla nel database che effettivamente
//esista una corrispondenza sulla base del relativo incrocio. 
func (db *Database) PreciseQuery(long float64, lat float64, day int64) (Data, error) {
	//Calcola l'indice della stazione.
	var idx int64
	{
		bytes := floatToBytes(long)
		hasher.Sum(bytes[:])
		bytes = floatToBytes(lat)
		hasher.Sum(bytes[:])
		idx = int64(hasher.Sum64())
		hasher.Reset()
	}

	//Query al database.
	stmt, err := db.conn.Prepare(
		"SELECT time,long,lat,pm25_concentration,temperature FROM d" + fmt.Sprintf("%v", truncateTime(day)) + " WHERE idx=$1",
	)
	if err != nil {
		return Data{}, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(idx)

	//TODO: Migliorare error.handling
	if row != nil {
		var ret Data
		err = row.Scan(&ret.Ts, &ret.Longitude, &ret.Latitude, &ret.Pm25Aqi, &ret.Temperature)
		if err != nil {
			return Data{}, err
		}

		return ret, nil
	} else {
		return Data{}, NotFoundError{fmt.Sprintf("Station %v not found", idx)}
	}
}
//Chiede latitudine, longitudine, un punto temporale e un range di coordinate e cerca
//nel range tutte le stazioni, prendendone la relativa scansione più vicina temporalmente al punto temporale passato 
func (db *Database) ApproximateQuery(long float64, lat float64, day int64, rng float64) ([]Data, error) {
	//TODO: Controllare che rng sia positivo
	Log.Printf("%v %v\t%v %v\n", long, rng, long-rng, long+rng)
	//Controlla se l'inizio e la fine del range sono inclusi nella medesima tabella.
	if getIndexFromLongitude(long-rng) == getIndexFromLongitude(long+rng) {
		return db.actualApproximateQuery(long-rng, long+rng, lat, rng, day)
	} else {
		//Se non si trovano nella stessa tabella, procedi nella raccolta dei dati da più tabelle
		//TODO: Farlo con le goroutines
		lowerLimit := math.Max(long-rng, -180)
		Log.Printf("%v %v\n", lowerLimit, 180)
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
		//i: Coordinata alla tabella dopo
		for i := float64(getIndexFromLongitude(lowerLimit) + longPerTable); i <= upperLimit && i < 180; i += longPerTable {
			Log.Printf("%v %v\n", i, i+longPerTable)
			part, err := db.actualApproximateQuery(i, math.Min(i+longPerTable, upperLimit), lat, rng, day)
			if err != nil {
				return nil, err
			}

			//Aggiugni ogni risultato alla lista che li raccoglie
			ret = append(ret, part...)
		}
		return ret, nil
	}
}

func (db *Database) actualApproximateQuery(longMin float64, longMax float64, lat float64, latrng float64, day int64) ([]Data, error) {
	//Raccogli gli indici di tutte le stazioni prendendo come longitudine da longMin a longMax e come latitudine
	// da lat-latrng e lat+latrng.
	idxs, err := (func() ([]int64, error) {
		stmt, err := db.conn.Prepare(
			"SELECT idx FROM \"long" + fmt.Sprintf("%v", getIndexFromLongitude(longMin)) +
				"\" WHERE long >= $1 AND long <= $2 AND lat >= $3 AND lat <= $4",
		)
		if err != nil {
			return nil, err
		}
		defer stmt.Close()

		rows, err := stmt.Query(longMin, longMax, lat-latrng, lat+latrng)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		idxs := make([]int64, 0)
		for next := rows.Next(); next && err == nil; next = rows.Next() {
			idxs = append(idxs, 0)
			err = rows.Scan(&idxs[len(idxs)-1])
		}

		if err != nil {
			return nil, err
		} else {
			return idxs, nil
		}
	})()
	if err != nil {
		return nil, err
	}
	if len(idxs) == 0 {
		return nil, nil
	}

	Log.Printf("%v", idxs)
	// Se c'è più di un indice prendiamo la rilevazione più vicina temporalmente a quello richiesto
	stmt, err := db.conn.Prepare(
		"SELECT time,long,lat,pm25_concentration,temperature FROM d" + fmt.Sprintf("%v", truncateTime(day)) + " WHERE idx = $1",
	)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	ret := make([]Data, 0)
	for _, idx := range idxs {
		Log.Printf("%v\n", idx)
		rows, err := stmt.Query(idx)
		if err != nil {
			return nil, err
		}

		err = (func(rows *sql.Rows) error {
			defer rows.Close()
			var (
				closer  Data
				current Data
			)
			//Su tutte le rilevazioni prendi quella più vicina temporalmente a quello richiesto
			for next := rows.Next(); next; next = rows.Next() {
				if err = rows.Err(); err != nil {
					return err
				}
				Log.Printf("%v\n", next)

				err = rows.Scan(&current.Ts, &current.Longitude, &current.Latitude, &current.Pm25Aqi, &current.Temperature)
				if err != nil {
					return err
				}

				Log.Printf("%v\t%v\n", abs(current.Ts-day), abs(closer.Ts-day))
				if abs(current.Ts-day) < abs(closer.Ts-day) {
					closer = current
				}
			}
			ret = append(ret, closer)
			return nil
		})(rows)
		if err != nil {
			return nil, err
		}
	}

	return ret, nil
}

//Inserisci data nel database
func (db *Database) Insert(data Data) error {
	//TODO: Calcolare indice anzichè cercarlo
	// Cerco la presenza dell'indice
	stmt, err := db.conn.Prepare(
		"SELECT idx FROM long" + fmt.Sprintf("%v", getIndexFromLongitude(data.Longitude)) + " WHERE long = $1 AND lat = $2",
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	//Inizio transazione
	tx, err := db.conn.Begin()
	if err != nil {
		Log.Printf("%v\n", err)
		return err
	}

	//Se l'indice è presente salvalo, altrimenti procedi ad inserirlo
	row := stmt.QueryRow(data.Longitude, data.Latitude)
	var id int64
	err = row.Scan(&id)
	if _, ok := err.(*pq.Error); err != nil && !ok {
		if _, ok := err.(*pq.Error); !ok {
			bytes := floatToBytes(data.Longitude)
			hasher.Sum(bytes[:])
			bytes = floatToBytes(data.Latitude)
			hasher.Sum(bytes[:])
			id = int64(hasher.Sum64())
			hasher.Reset()

			_, err = tx.Exec(
				"INSERT INTO long"+fmt.Sprintf("%v", getIndexFromLongitude(data.Longitude))+" VALUES ($1, $2, $3)",
				data.Longitude, data.Latitude, id,
			)
			if err != nil {
				tx.Rollback()
				return err
			}
		} else {
			tx.Rollback()
			return err
		}
	}

	//Se non esiste crea la tabella del relativo giorno
	err = createTimeNamedTable(tx, data.Ts)
	if err != nil {
		Log.Println("Let's go")
		tx.Rollback()
		return err
	}

	//Inserisci nella tabella il dato data
	_, err = tx.Exec(
		"INSERT INTO d"+fmt.Sprintf("%v", truncateTime(data.Ts))+" VALUES ($1, $2, $3, $4, $5, $6)",
		id, data.Ts, data.Longitude, data.Latitude, data.Pm25Aqi, data.Temperature,
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	//Se non sono avvenuti errori, completa la transazione
	err = tx.Commit()

	return err
}
//Chiude la connessione
func (db *Database) Close() error {
	return db.conn.Close()
}

func createTimeNamedTable(db *sql.Tx, time int64) error {
	//Creo una tabella per la giornata time
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS d" + fmt.Sprintf("%v", truncateTime(time)) + `(
    idx bigint,
    time bigint,
    long double precision,
    lat double precision,
    pm25_concentration double precision,
    temperature integer,
    PRIMARY KEY(idx, time)
)`)
	return err
}

func performOneTimeSetup(db *sql.DB) (bool, error) {
	/*Eseguiamo la query e controlliamo exists per controllare se una riga
	 *effettivamente esiste. In tal caso non agiremo e ritorneremo false
	 *in caso contrario procederemo a creare le tabelle e a ritornare true*/
	_, err := db.Exec("SELECT idx FROM long0")
	if err == nil {
		return false, nil
	}

	pqerr, ok := err.(*pq.Error)
	if !ok || pqerr.Code != "42P01" {
		return false, err
	}

	//Creo una tabella ogni 5 "gradi"
	tx, err := db.Begin()
	if err != nil {
		return false, err
	}

	for i := -180; i < 180; i += 5 {
		_, err := tx.Exec("CREATE TABLE \"long" + fmt.Sprintf("%v", i) + `" (
		long double precision,
		lat double precision,
		idx bigint,
		PRIMARY KEY (long, lat)
	)`)
		if err != nil {
			tx.Rollback()
			return true, err
		}
	}
	err = tx.Commit()
	return true, err
}
