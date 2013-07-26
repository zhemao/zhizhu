package main

const (
	SUCCESS = iota
	ERROR = iota
	TOTALSIZE = iota
	PROGRESS = iota
)

type ProgressUpdate struct {
	id int
	messType int
	amount int64
	err error
}
