package main

import (
	"os"
	"bufio"
	"strings"
	"path"
)

func makeRequest(url string, fname string) DownloadRequest {
	outfname := fname + ".part"
	file, err := os.Open(outfname)
	var initSize int64
	if err != nil {
		if os.IsNotExist(err) {
			initSize = 0
		} else {
			panic(err)
		}
	} else {
		defer file.Close()
		initSize, err = file.Seek(0, os.SEEK_END)
		if err != nil {
			panic(err)
		}
	}
	return DownloadRequest{url, outfname, fname, initSize}
}

func loadRequests(reqFileName string) []DownloadRequest {
	reqFile, err := os.Open(reqFileName)
	if err != nil {
		panic(err)
	}
	defer reqFile.Close()

	table := make([]DownloadRequest, 0)
	scanner := bufio.NewScanner(reqFile)

	for scanner.Scan() {
		cols := strings.Split(scanner.Text(), " ")
		if len(cols) == 0 {
			continue
		}
		url := cols[0]
		var fname string
		if len(cols) == 1 {
			fname = path.Base(url)
		} else {
			fname = cols[1]
		}
		table.Append(makeRequest(url, fname))
	}
	return table
}
