package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
)

func makeRequest(url string, basename string) (DownloadRequest, error) {
	actualpath := path.Join(os.Getenv("HOME"), "Downloads", basename)
	outpath := actualpath + ".part"
	file, err := os.Open(outpath)
	var initSize int64
	if err != nil {
		if os.IsNotExist(err) {
			initSize = 0
		} else {
			return DownloadRequest{}, err
		}
	} else {
		defer file.Close()
		initSize, err = file.Seek(0, os.SEEK_END)
		if err != nil {
			return DownloadRequest{}, err
		}
	}
	return DownloadRequest{url, basename, outpath, actualpath, initSize}, nil
}

func loadRequests(reqFileName string) ([]DownloadRequest, error) {
	reqFile, err := os.Open(reqFileName)
	if os.IsNotExist(err) {
		return make([]DownloadRequest, 0), nil
	} else if err != nil {
		return nil, err
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
		var basename string
		if len(cols) == 1 {
			basename = path.Base(url)
		} else {
			basename = cols[1]
		}
		req, err := makeRequest(url, basename)
		if err != nil {
			return nil, err
		}
		table = append(table, req)
	}
	return table, nil
}

func cleanupReqFile(reqFileName string, requests *[]DownloadStatus) {
	reqFile, err := os.Create(reqFileName)
	if err != nil {
		panic(err)
	}
	for _, req := range *requests {
		if !req.done {
			fmt.Fprintf(reqFile, "%s\n", req.url)
		}
	}
	reqFile.Close()
}
