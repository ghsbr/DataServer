package main

import (
	"flag"
	"net"
	"os"
)

var printDebug bool

func main() {
	var address string
	flag.StringVar(
		&address, "addr", "127.0.0.1:8080",
		"Address at which the requests will be served",
	)

	flag.BoolVar(&printDebug, "debug", false, "Print Debug Messages")
	help := flag.Bool("help", true, "Show help message and exit")
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

	server, err := net.Listen(
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
	}
}

func connectionHandler(conn net.Conn) {
	defer conn.Close()
	if printDebug {
		println("Serving request sent by", conn.RemoteAddr().String())
	}
}
