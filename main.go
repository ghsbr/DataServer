package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
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

	db, mod, err := NewDatabase( /*"data.db"*/ ":memory:")
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	defer db.Close()
	if printDebug {
		fmt.Printf("Was setup performed? %v", mod)
	}

	var jsonExample []byte
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

	db.Insert(out)

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
			continue
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
