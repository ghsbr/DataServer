package main

import (
	"os"

	"github.com/bvinc/go-sqlite-lite/sqlite3" //sqlite
)

func main() {
	conn, err := sqlite3.Open(":memory:") //creazione memoria database
	if err != nil {			      
		println(err.Error())		
		os.Exit(1)			
	}
	defer conn.Close()		//chiudu il database a fine esecuzione

	//crea tab studente con campi nome e età, ritorna in ogni caso
	if err = conn.Exec(`CREATE TABLE student(name TEXT, age INTEGER)`); err != nil {
		println(err.Error())
		os.Exit(2)
	}
	
	//inserisci studente che ha nome ed età 
	if err = conn.Exec(`INSERT INTO student (name, age) VALUES ("Giovanni", 16)`); err != nil {
		println(err.Error())
		os.Exit(3)
	}
	
	//seleziona tutte le colonne dalla tabella student
	stmt, err := conn.Prepare(`SELECT * FROM student`)
	if err != nil {
		println(err.Error())
		os.Exit(4)
	}
	defer stmt.Close() //chiudi a fine esecuzione

	//Printa tutte le righe
	var name string
	var age int
	for end, err := stmt.Step(); end && err == nil; end, err = stmt.Step() {
		stmt.Scan(&name, &age) //leggi riga
		println(name, age)
	}
	if err != nil {
		println(err.Error())
		os.Exit(5)
	}
}
