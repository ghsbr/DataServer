package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
)

var printDebug bool

func main() {
	var addr string
	flag.StringVar(
		&addr, "addr", "127.0.0.1:8080",
		" at which the requests will be served",
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

	db, mod, err := NewDatabase( /*"data.db"*/ ":memory:")
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()
	if printDebug {
		fmt.Printf("Was setup performed? %v\n", mod)
	}

	//Query Handler
	query := func(resp http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			err := req.ParseForm()
			if err != nil {
				log.Println(err)
				resp.Write([]byte("Form not valid"))
				return
			}

			getFromForm := func(paramName string, out interface{}) error {
				if str := req.PostForm.Get(paramName); str != "" {
					t := reflect.TypeOf(out)
					if t == nil {
						return TypeError{"Cannot get type of second parameter"}
					}

					if t.String() == "*uint64" {
						if out == nil {
							out = new(uint64)
						}
						typedOut := out.(*uint64)
						*typedOut, err = strconv.ParseUint(str, 10, 64)
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

			var day uint64
			err = getFromForm("day", &day)
			if err != nil {
				log.Println(err)
				resp.Write([]byte(err.Error()))
			}
			/*if daystr := req.PostForm.Get("day"); daystr != "" {
				day, err = strconv.ParseUint(daystr, 10, 64)
				if err != nil {
					log.Println(err)
					resp.Write([]byte("Error: Day is not a number"))
					return
				}
			} else {
				log.Println("Day not present")
				resp.Write([]byte("Error: Day not present"))
				return
			}*/

			var long float64
			err = getFromForm("long", &long)
			if err != nil {
				log.Println(err)
				resp.Write([]byte(err.Error()))
			}
			/*if longstr := req.PostForm.Get("long"); longstr != "" {
				long, err = strconv.ParseFloat(longstr, 64)
				if err != nil {
					log.Println(err)
					resp.Write([]byte("Error: long is not a float"))
					return
				}
			} else {
				log.Println("long not present")
				resp.Write([]byte("Error: long not present"))
				return
			}*/

			var lat float64
			err = getFromForm("lat", &lat)
			if err != nil {
				log.Println(err)
				resp.Write([]byte(err.Error()))
			}
			/*if latstr := req.PostForm.Get("lat"); latstr != "" {
				lat, err = strconv.ParseFloat(latstr, 64)
				if err != nil {
					log.Println(err)
					resp.Write([]byte("Error: lat is not a float"))
					return
				}
			} else {
				log.Println("lat not present")
				resp.Write([]byte("Error: lat not present"))
				return
			}*/

			var data interface{}
			if rangestr := req.PostForm.Get("range"); rangestr != "" {
				rng, err := strconv.ParseFloat(rangestr, 64)
				if err != nil {
					log.Println(err)
					resp.Write([]byte("Error: lat is not a float"))
					return
				}

				data, err = db.ApproximateQuery(long, lat, day, rng)
			} else {
				data, err = db.PreciseQuery(long, lat, day)
			}
			if err != nil {
				resp.Write([]byte(err.Error()))
				log.Println(err)
				return
			}

			data, err = json.Marshal(data)
			if err != nil {
				resp.Write([]byte(err.Error()))
				log.Println(err)
			} else {
				resp.Write(data.([]byte))
			}
		} else {
			_, err := resp.Write([]byte("Error: Wrong Method"))
			if err != nil {
				log.Println(err)
			} else {
				log.Println("Wrong Request Method")
			}
		}
	}

	//Insert Handler
	insert := func(resp http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				log.Println(err)
				resp.Write([]byte(err.Error()))
				return
			}

			//TODO: Validate json

			var inData Data
			err = json.Unmarshal(body, &inData)
			if err != nil {
				log.Println(err)
				resp.Write([]byte(err.Error()))
				return
			}

			err = db.Insert(inData)
			if err != nil {
				log.Println(err)
				resp.Write([]byte(err.Error()))
			} else {
				resp.Write([]byte("Ok"))
				log.Println("json written correctly in database")
			}

		} else {
			_, err := resp.Write([]byte("Error: Wrong Method"))
			if err != nil {
				log.Println(err)
			} else {
				log.Println("Wrong Request Method")
			}
		}
	}

	http.HandleFunc("/insert", insert)
	http.HandleFunc("/query", query)
	log.Fatalln(http.ListenAndServe(addr, nil).Error())
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
