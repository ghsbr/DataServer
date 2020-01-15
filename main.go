package main

import (
	"net"
	"os"
)

func main() {
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
}
