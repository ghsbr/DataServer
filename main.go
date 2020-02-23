package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ghsbr/DataServer/data"
	"github.com/ghsbr/DataServer/database"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
)

var Log *log.Logger

type (
	Data     = data.Data
	Database = database.Database
)

func main() {
	addr := flag.String(
		"addr", "127.0.0.1:8080",
		" at which the requests will be served",
	)

	printDebug := flag.Bool("debug", false, "Print Debug Messages")
	help := flag.Bool("help", false, "Show help message and exit")
	flag.Parse()

	if *help {
		fmt.Print("DataServer v0.0.1\n\n")
		flag.CommandLine.SetOutput(os.Stdout)
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *printDebug {
		Log = log.New(os.Stdout, "[DataServer] ", 0)
	} else {
		Log = log.New(ioutil.Discard, "", 0)
	}
	database.SetLogger(Log)

	db, mod, err := database.NewDatabase("data.db")
	if err != nil {
		Log.Fatalln(err)
	}
	defer db.Close()
	Log.Printf("Was setup performed? %v\n", mod)

	//Query Handler
	query := makeQuery(&db)

	//Insert Handler
	insert := makeInsert(&db)

	http.HandleFunc("/insert", insert)
	http.HandleFunc("/query", query)
	Log.Printf("Trying to serve on address: %v\n", *addr)
	Log.Fatalln(http.ListenAndServe(*addr, nil).Error())
}

type TypeError struct {
	msg string
}

func (err TypeError) Error() string {
	return err.msg
}

type ParameterError struct {
	msg string
}

func (err ParameterError) Error() string {
	return err.msg
}

func makeQuery(db *Database) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			err := req.ParseForm()
			if err != nil {
				Log.Println(err)
				resp.Write([]byte("Form not valid"))
				return
			}

			getFromForm := func(paramName string, out interface{}) error {
				if str := req.PostForm.Get(paramName); str != "" {
					t := reflect.TypeOf(out)
					if t == nil {
						return TypeError{"Cannot get type of second parameter"}
					}

					if t.String() == "*int64" {
						if out == nil {
							out = new(uint64)
						}
						typedOut := out.(*int64)
						*typedOut, err = strconv.ParseInt(str, 10, 64)
					} else if t.String() == "*float64" {
						if out == nil {
							out = new(float64)
						}
						typedOut := out.(*float64)
						*typedOut, err = strconv.ParseFloat(str, 64)
					} else {
						return TypeError{"Type of second parameter is not usable"}
					}

					if err != nil {
						return err
					}
				} else {
					return ParameterError{fmt.Sprintf("%v not present\n", paramName)}
				}
				return nil
			}

			var day int64
			err = getFromForm("day", &day)
			if err != nil {
				Log.Println(err)
				resp.Write([]byte(err.Error()))
			}

			var long float64
			err = getFromForm("long", &long)
			if err != nil {
				Log.Println(err)
				resp.Write([]byte(err.Error()))
			}

			var lat float64
			err = getFromForm("lat", &lat)
			if err != nil {
				Log.Println(err)
				resp.Write([]byte(err.Error()))
			}

			var data interface{}
			var rng float64
			if err = getFromForm("range", &rng); err != nil {
				Log.Printf(
					"Performing PreciseQuery on {Ts: %v, Long: %v, Lat: %v}\n",
					day, long, lat,
				)
				data, err = db.PreciseQuery(long, lat, day)
			} else {
				Log.Printf(
					"Performing ApproximateQuery on {Ts: %v, Long: %v, Lat: %v, Range: %v}\n",
					day, long, lat, rng,
				)
				data, err = db.ApproximateQuery(long, lat, day, rng)
			}
			if err != nil {
				resp.Write([]byte(err.Error()))
				Log.Println(err)
				return
			}

			data, err = json.Marshal(data)
			if err != nil {
				resp.Write([]byte(err.Error()))
				Log.Println(err)
			} else {
				resp.Write(data.([]byte))
			}
		} else {
			_, err := resp.Write([]byte("Error: Wrong Method"))
			if err != nil {
				Log.Println(err)
			} else {
				Log.Println("Wrong Request Method")
			}
		}
	}
}

func makeInsert(db *Database) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				Log.Println(err)
				resp.Write([]byte(err.Error()))
				return
			}

			//TODO: Validate json

			var inData Data
			err = json.Unmarshal(body, &inData)
			if err != nil {
				Log.Println(err)
				resp.Write([]byte(err.Error()))
				return
			}

			err = db.Insert(inData)
			if err != nil {
				Log.Println(err)
				resp.Write([]byte(err.Error()))
			} else {
				resp.Write([]byte("Ok"))
				Log.Println("json written correctly to database")
			}

		} else {
			_, err := resp.Write([]byte("Error: Wrong Method"))
			if err != nil {
				Log.Println(err)
			} else {
				Log.Println("Wrong Request Method")
			}
		}
	}
}
