package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ghsbr/DataServer/data"
	"github.com/ghsbr/DataServer/database"
	"github.com/qri-io/jsonschema"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
)

var (
	Log       *log.Logger
	schema    jsonschema.RootSchema
	rawSchema []byte = []byte(`{
  "definitions": {},
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "http://example.com/root.json",
  "type": "object",
  "title": "The Root Schema",
  "required": [
    "ts",
    "pm25",
    "temperature",
    "coordinates"
  ],
  "properties": {
    "ts": {
      "$id": "#/properties/ts",
      "type": "string",
      "title": "The Ts Schema",
      "default": "",
      "examples": [
        "2020-01-15T21:00:00.000Z"
      ],
      "pattern": "^(.*)$"
    },
    "pm25": {
      "$id": "#/properties/pm25",
      "type": "object",
      "title": "The Pm25 Schema",
      "required": [
        "aqi"
      ],
      "properties": {
        "aqi": {
          "$id": "#/properties/pm25/properties/aqi",
          "type": "integer",
          "title": "The Aqi Schema",
          "default": 0,
          "examples": [
            421
          ]
        }
      }
    },
    "temperature": {
      "$id": "#/properties/temperature",
      "type": "integer",
      "title": "The Temperature Schema",
      "default": 0,
      "examples": [
        17
      ]
    },
    "coordinates": {
      "$id": "#/properties/coordinates",
      "type": "object",
      "title": "The Coordinates Schema",
      "required": [
        "latitude",
        "longitude"
      ],
      "properties": {
        "latitude": {
          "$id": "#/properties/coordinates/properties/latitude",
          "type": "number",
          "title": "The Latitude Schema",
          "default": 0.0,
          "examples": [
            23.79611913
          ]
        },
        "longitude": {
          "$id": "#/properties/coordinates/properties/longitude",
          "type": "number",
          "title": "The Longitude Schema",
          "default": 0.0,
          "examples": [
            90.41756094
          ]
        }
      }
    }
  }
}`)
)

type (
	Data     = data.Data
	Database = database.Database
)

func main() {
	addr := flag.String(
		"addr", "localhost:8080",
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

	//Crea databaseHandler
	db, mod, err := database.NewDatabase("dataserver", "dataserver", "dataserver", Log)
	if err != nil {
		Log.Fatalln(err)
	}
	defer db.Close()
	Log.Printf("Was setup performed? %v\n", mod)

	//Prepara JSON Schema
	err = json.Unmarshal(rawSchema, &schema)
	if err != nil {
		Log.Fatalln(err)
	}

	//Query Handler
	query := makeQuery(db)

	//Insert Handler
	insert := makeInsert(db)

	//Aggiungi Handlers e inizia ad ascoltare sull'indirizzo addr
	http.HandleFunc("/insert", insert)
	http.HandleFunc("/query", query)
	Log.Printf("Trying to serve on address: %v\n", *addr)
	Log.Fatalln(http.ListenAndServe(*addr, nil).Error())
}

func makeQuery(db *Database) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		//TODO: controllare vari returns
		//Se il metodo della richiesta è POST elaborala
		resp.Header().Add("Access-Control-Allow-Origin", "*")
		if req.Method == "POST" {
			err := req.ParseForm()
			if err != nil {
				Log.Println(err)
				resp.Write([]byte("Form not valid"))
				return
			}

			//Closure per estrapolare dati dal POST form
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

			//Ottieni un punto temporale
			var day int64
			err = getFromForm("day", &day)
			if err != nil {
				Log.Println(err)
				resp.Write([]byte(err.Error()))
				return
			}

			//Ottieni la longitudine
			var long float64
			err = getFromForm("long", &long)
			if err != nil {
				Log.Println(err)
				resp.Write([]byte(err.Error()))
				return
			}

			//Ottieni la latitudine
			var lat float64
			err = getFromForm("lat", &lat)
			if err != nil {
				Log.Println(err)
				resp.Write([]byte(err.Error()))
				return
			}

			//Se il parametro range esiste allora crea una ApproximateQuery con esso
			//altrimenti fai una PreciseQuery e salvalo dentro data
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

			//Se tutto viene completato senza errori, serializza i dati e rispondi.
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
		// Se il metodo corrisponde a POST elaboralo
		resp.Header().Add("Access-Control-Allow-Origin", "*")
		if req.Method == "POST" {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				Log.Println(err)
				resp.Write([]byte(err.Error()))
				return
			}

			//Confronta il JSON con lo Schema
			{
				errs, err := schema.ValidateBytes(body)
				if err != nil {
					Log.Println(err)
					resp.Write([]byte("Bad JSON:\n" + err.Error()))
					return
				}
				if len(errs) > 0 {
					Log.Println("Bad JSON:")
					resp.Write([]byte("Bad JSON:\n"))
					for _, err = range errs[:len(errs)-1] {
						Log.Println(err)
						resp.Write([]byte(err.Error() + "\n"))
					}
					Log.Println(err)
					resp.Write([]byte(err.Error()))
					return
				}
			}

			//Se il JSON corrisponde allo Schema deserializzalo e inseriscilo nel database
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
