package main

import (
	"fmt"
	"os"
	"io"
	"errors"
	"net/http"
)

const CHUNK_SIZE = 4096

func skipAhead(channel chan ProgressUpdate, id int, body io.Reader,
				skipAmount int64) error {
	dlAmount := int64(0)
	buf := make([]byte, CHUNK_SIZE)
	limited := io.LimitReader(body, skipAmount)
	for dlAmount < skipAmount {
		n, err := limited.Read(buf)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		dlAmount += int64(n)
		channel <- ProgressUpdate{id, PROGRESS, dlAmount, nil}
	}
	return nil
}

func downloadFile(channel chan ProgressUpdate, id int, body io.Reader,
					out *os.File, initSize int64) {
	dlAmount := initSize
	for {
		n, err := io.CopyN(out, body, CHUNK_SIZE)
		if n > 0 {
			dlAmount += n
			channel <- ProgressUpdate{id, PROGRESS, dlAmount, nil}
		}

		if err == io.EOF {
			channel <- ProgressUpdate{id, SUCCESS, 0, nil}
			return
		} else if err != nil {
			channel <- ProgressUpdate{id, ERROR, 0, err}
			return
		}
	}
}

func makeHttpRequest(url string, initSize int64) (*http.Response, error) {
	client := &http.Client{
		CheckRedirect: func (req *http.Request, via []*http.Request) error {
			if initSize > 0 {
				req.Header.Add("Range", fmt.Sprintf("bytes=%d-", initSize))
			}
			return nil
		},
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if initSize > 0 {
		req.Header.Add("Range", fmt.Sprintf("bytes=%d-", initSize))
	}

	resp, err := client.Do(req)
	return resp, err
}

func runDownload(channel chan ProgressUpdate, id int, dlreq *DownloadRequest) {
	var out *os.File
	var err error
	if dlreq.initSize == 0 {
		out, err = os.Create(dlreq.outfname)
	} else {
		out, err = os.OpenFile(dlreq.outfname, os.O_WRONLY | os.O_APPEND, 0644)
	}
	if err != nil {
		channel <- ProgressUpdate{id, ERROR, 0, err}
		return
	}
	defer out.Close()

	resp, err := makeHttpRequest(dlreq.url, dlreq.initSize)

	if err != nil {
		channel <- ProgressUpdate{id, ERROR, 0, err}
		return
	}

	if resp.StatusCode == http.StatusOK {
		channel <- ProgressUpdate{id, TOTALSIZE, resp.ContentLength, nil}
		skipAhead(channel, id, resp.Body, dlreq.initSize)
		downloadFile(channel, id, resp.Body, out, dlreq.initSize)
	} else if resp.StatusCode == http.StatusPartialContent {
		totalSize := resp.ContentLength + dlreq.initSize
		channel <- ProgressUpdate{id, TOTALSIZE, totalSize, nil}
		downloadFile(channel, id, resp.Body, out, dlreq.initSize)
	} else {
		err := errors.New(resp.Status)
		channel <- ProgressUpdate{id, ERROR, 0, err}
		return
	}
}
