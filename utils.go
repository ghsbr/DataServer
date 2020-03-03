package main

//Quando il tipo del puntatore in cui scrivere il parametro del
//POST form non viene riconosciuto o è sbagliato, viene ritornato questo errore. 
type TypeError struct {
	msg string
}

func (err TypeError) Error() string {
	return err.msg
}
//Viene ritornato questo errore quando il parametro a cui si fa
//riferimento non esiste nel POST form
type ParameterError struct {
	msg string
}

func (err ParameterError) Error() string {
	return err.msg
}
