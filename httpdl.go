package main

import (
	"fmt"
	"os"
	"io"
	"time"
	"errors"
	"net/http"
)

const CHUNK_SIZE = 4096

func finishDownload(updateChan chan ProgressUpdate,
					id int, dlreq DownloadRequest, err error) {
	if err != nil {
		updateChan <- ProgressUpdate{id, ERROR, 0, err}
		return
	}
	err = os.Rename(dlreq.outpath, dlreq.actualpath)
	if err != nil {
		updateChan <- ProgressUpdate{id, ERROR, 0, err}
	} else {
		updateChan <- ProgressUpdate{id, SUCCESS, 0, nil}
	}
}

func checkPause(ctrlChan chan int, oldPause bool) (bool, error) {
	select {
	case command := <-ctrlChan:
		switch command {
		case PAUSE:
			return true, nil
		case RESUME:
			return false, nil
		case CANCEL:
			return true, errors.New("Canceled by user")
		}
	default:
		return oldPause, nil
	}
	return oldPause, nil
}

func skipAhead(updateChan chan ProgressUpdate, ctrlChan chan int,
				id int, body io.Reader, skipAmount int64) error {
	dlAmount := int64(0)
	buf := make([]byte, CHUNK_SIZE)
	limited := io.LimitReader(body, skipAmount)
	pause := false
	var err error
	for dlAmount < skipAmount {
		pause, err = checkPause(ctrlChan, pause)
		if err != nil {
			return err
		}
		if pause {
			time.Sleep(time.Second)
		} else {
			n, err := limited.Read(buf)
			if err == io.EOF {
				return nil
			} else if err != nil {
				return err
			}
			dlAmount += int64(n)
			updateChan <- ProgressUpdate{id, PROGRESS, dlAmount, nil}
		}
	}
	return nil
}

func downloadFile(updateChan chan ProgressUpdate,
				  ctrlChan chan int,
				  id int, body io.Reader,
				  out *os.File, initSize int64) error {
	dlAmount := initSize
	pause := false
	var err error
	for {
		pause, err = checkPause(ctrlChan, pause)
		if err != nil {
			return err
		}
		if pause {
			time.Sleep(time.Second)
		} else {
			n, err := io.CopyN(out, body, CHUNK_SIZE)
			if n > 0 {
				dlAmount += n
				updateChan <- ProgressUpdate{id, PROGRESS, dlAmount, nil}
			}

			if err == io.EOF {
				return nil
			} else if err != nil {
				return err
			}
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

func runDownload(updateChan chan ProgressUpdate, ctrlChan chan int,
				 id int, dlreq DownloadRequest) {
	var out *os.File

	// if the file has already been downloaded, do nothing
	_, err := os.Stat(dlreq.actualpath)
	if err == nil {
		updateChan <- ProgressUpdate{id, SUCCESS, 0, nil}
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
		updateChan <- ProgressUpdate{id, ERROR, 0, err}
		return
	}
	defer out.Close()

	resp, err := makeHttpRequest(dlreq.url, dlreq.initSize)

	if err != nil {
		updateChan <- ProgressUpdate{id, ERROR, 0, err}
		return
	}

	if resp.StatusCode == http.StatusOK {
		if resp.ContentLength == dlreq.initSize {
			finishDownload(updateChan, id, dlreq, nil)
			return
		}
		updateChan <- ProgressUpdate{id, TOTALSIZE, resp.ContentLength, nil}
		skipAhead(updateChan, ctrlChan, id, resp.Body, dlreq.initSize)
		err := downloadFile(updateChan, ctrlChan, id, resp.Body, out, dlreq.initSize)
		finishDownload(updateChan, id, dlreq, err)
	} else if resp.StatusCode == http.StatusPartialContent {
		totalSize := resp.ContentLength + dlreq.initSize
		if totalSize == 0 {
			finishDownload(updateChan, id, dlreq, nil)
			return
		}
		updateChan <- ProgressUpdate{id, TOTALSIZE, totalSize, nil}
		err := downloadFile(updateChan, ctrlChan, id, resp.Body, out, dlreq.initSize)
		finishDownload(updateChan, id, dlreq, err)
	} else {
		err := errors.New(resp.Status)
		updateChan <- ProgressUpdate{id, ERROR, 0, err}
		return
	}
}
