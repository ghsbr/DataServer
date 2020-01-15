package main

import (
	"os"

	"github.com/bvinc/go-sqlite-lite/sqlite3"
)

func main() {
	conn, err := sqlite3.Open(":memory:")
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	if err = conn.Exec(`CREATE TABLE student(name TEXT, age INTEGER)`); err != nil {
		println(err.Error())
		os.Exit(2)
	}

	if err = conn.Exec(`INSERT INTO student (name, age) VALUES ("Giovanni", 16)`); err != nil {
		println(err.Error())
		os.Exit(3)
	}

	stmt, err := conn.Prepare(`SELECT * FROM student`)
	if err != nil {
		println(err.Error())
		os.Exit(4)
	}
	defer stmt.Close()

	var name string
	var age int
	for end, err := stmt.Step(); end && err == nil; end, err = stmt.Step() {
		stmt.Scan(&name, &age)
		println(name, age)
	}
	if err != nil {
		println(err.Error())
		os.Exit(5)
	}
}
