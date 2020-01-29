package main

import (
	"flag"
	//"net"
	"os"
	"strconv"
	"time"

	"github.com/bvinc/go-sqlite-lite/sqlite3"
)

var printDebug bool

const longPerTable = 5

func main() {
	var address string
	flag.StringVar(
		&address, "addr", "127.0.0.1:8080",
		"Address at which the requests will be served",
	)

	flag.BoolVar(&printDebug, "debug", true, "Print Debug Messages")
	help := flag.Bool("help", false, "Show help message and exit")
	flag.Parse()

	if *help {
		print("DataServer v0.0.1\n\n")
		flag.CommandLine.SetOutput(os.Stdout)
		flag.PrintDefaults()
		os.Exit(0)
	}

	if printDebug {
		println("Debug Messages on")
	}

	conn, err := sqlite3.Open("data.db")
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	mod, err := performOneTimeSetup(conn)
	if err != nil {
		println(err.Error())
		os.Exit(2)
	}
	if printDebug {
		println("Was setup performed? " + strconv.FormatBool(mod))
	}

	/*server, err := net.Listen(
		"tcp",
		"127.0.0.1:8080",
	)
	if err != nil {
		println(err.Error())
		os.Exit(2)
	}
	defer server.Close()

	println("Serving at", server.Addr().String())

	for {
		conn, err := server.Accept()
		if err != nil {
			println(err.Error())
			break
		}
		go connectionHandler(conn)
	}*/
}

/*func connectionHandler(conn net.Conn) {
	defer conn.Close()
	if printDebug {
		println("Serving request sent by", conn.RemoteAddr().String())
	}
}*/

//TODO: usare una Session
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
	condition TEXT,
	humidity INTEGER,
	pressure INTEGER,
	wind_speed REAL,
	wind_direction INTEGER,
	temperature INTEGER,
	PRIMARY KEY(idx, time)
)`)
	if err != nil {
		return true, nil
	}
	return true, nil
}
