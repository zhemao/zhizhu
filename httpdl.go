package main

import (
	"fmt"
	"os"
	"io"
	"errors"
	"net/http"
)

const CHUNK_SIZE = 4096

func finishDownload(channel chan ProgressUpdate, id int,
					dlreq DownloadRequest, err error) {
	if err != nil {
		channel <- ProgressUpdate{id, ERROR, 0, err}
		return
	}
	err = os.Rename(dlreq.outpath, dlreq.actualpath)
	if err != nil {
		channel <- ProgressUpdate{id, ERROR, 0, err}
	} else {
		channel <- ProgressUpdate{id, SUCCESS, 0, nil}
	}
}

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
					out *os.File, initSize int64) error {
	dlAmount := initSize
	for {
		n, err := io.CopyN(out, body, CHUNK_SIZE)
		if n > 0 {
			dlAmount += n
			channel <- ProgressUpdate{id, PROGRESS, dlAmount, nil}
		}

		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
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

func runDownload(channel chan ProgressUpdate, id int, dlreq DownloadRequest) {
	var out *os.File

	// if the file has already been downloaded, do nothing
	_, err := os.Stat(dlreq.actualpath)
	if err == nil {
		channel <- ProgressUpdate{id, SUCCESS, 0, nil}
		return
	}

	// if we're starting fresh, create new file
	if dlreq.initSize == 0 {
		out, err = os.Create(dlreq.outpath)
	} else {
		// otherwise, append to the old one
		out, err = os.OpenFile(dlreq.outpath, os.O_WRONLY | os.O_APPEND, 0644)
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
		if resp.ContentLength == dlreq.initSize {
			finishDownload(channel, id, dlreq, nil)
			return
		}
		channel <- ProgressUpdate{id, TOTALSIZE, resp.ContentLength, nil}
		skipAhead(channel, id, resp.Body, dlreq.initSize)
		err := downloadFile(channel, id, resp.Body, out, dlreq.initSize)
		finishDownload(channel, id, dlreq, err)
	} else if resp.StatusCode == http.StatusPartialContent {
		totalSize := resp.ContentLength + dlreq.initSize
		if totalSize == 0 {
			finishDownload(channel, id, dlreq, nil)
			return
		}
		channel <- ProgressUpdate{id, TOTALSIZE, totalSize, nil}
		err := downloadFile(channel, id, resp.Body, out, dlreq.initSize)
		finishDownload(channel, id, dlreq, err)
	} else {
		err := errors.New(resp.Status)
		channel <- ProgressUpdate{id, ERROR, 0, err}
		return
	}
}
