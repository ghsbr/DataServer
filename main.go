package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
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

	/*	var jsonExample []byte
		{
			file, err := os.Open("data_example.json")
			if err != nil {
				println(err.Error())
				os.Exit(3)
			}

			if stat, err := file.Stat(); err == nil {
				jsonExample = make([]byte, stat.Size())
				file.Read(jsonExample)
			} else {
				println(err.Error())
				os.Exit(4)
			}
			file.Close()
		}
		if printDebug {
			println(string(jsonExample))
		}

		var out Data
		if err = json.Unmarshal(jsonExample, &out); err != nil {
			println(err.Error())
			os.Exit(5)
		}
		if printDebug {
			fmt.Printf("%+v\n", out)
		}

		db.Insert(out)*/

	log.Fatalln(http.ListenAndServe(addr, nil).Error())
}

func query(resp http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		err := req.ParseForm()
		if err != nil {
			log.Println(err)
		}

		var day uint64
		if daystr := req.PostForm.Get("day"); daystr != "" {
			day, err = strconv.ParseUint(daystr, 10, 64)
			if err != nil {
				log.Println(err)
				resp.Write([]byte("Error: Day is not a number"))
				return
			}
		} else {
			resp.Write([]byte("Error: Day not present"))
		}

	} else {
		b, err := resp.Write([]byte("Error: Wrong Method"))
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("Written %v bytes succesfully", b)
		}
	}
}
