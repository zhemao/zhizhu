package main

const (
	SUCCESS = iota
	ERROR = iota
	TOTALSIZE = iota
	PROGRESS = iota
	CANCELED = iota
)

const (
	PAUSE = iota
	RESUME = iota
	CANCEL = iota
)

type ProgressUpdate struct {
	id int
	messType int
	amount int64
	err error
}

type DownloadRequest struct {
	url string
	basename string
	outpath string
	actualpath string
	initSize int64
}

type DownloadStatus struct {
	url string
	fname string
	dlAmount int64
	totalAmount int64
	done bool
}
