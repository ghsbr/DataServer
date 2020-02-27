package main

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
