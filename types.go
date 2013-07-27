package main

const (
	SUCCESS = iota
	ERROR = iota
	TOTALSIZE = iota
	PROGRESS = iota
	QUIT = iota
)

type ProgressUpdate struct {
	id int
	messType int
	amount int64
	err error
}

type DownloadRequest struct {
	url string
	outfname string
	actualfname string
	initSize int64
}

type DownloadStatus struct {
	url string
	fname string
	dlAmount int64
	totalAmount int64
	done bool
}
