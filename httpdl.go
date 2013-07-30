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

func finishDownload(updateChan chan ProgressUpdate, id int,
					dlreq DownloadRequest, status int, err error) {
	if status == CANCELED {
		os.Remove(dlreq.outpath)
	} else if status == SUCCESS {
		err = os.Rename(dlreq.outpath, dlreq.actualpath)
		if err != nil {
			updateChan <- ProgressUpdate{id, ERROR, 0, err}
			return
		}
	}
	updateChan <- ProgressUpdate{id, status, 0, err}
}

func checkPause(ctrlChan chan int, oldPause bool) (bool, bool) {
	select {
	case command := <-ctrlChan:
		switch command {
		case PAUSE:
			return true, false
		case RESUME:
			return false, false
		case CANCEL:
			return true, true
		}
	default:
		return oldPause, false
	}
	return oldPause, false
}

func skipAhead(updateChan chan ProgressUpdate, ctrlChan chan int,
				id int, body io.Reader, skipAmount int64) (int, error) {
	dlAmount := int64(0)
	buf := make([]byte, CHUNK_SIZE)
	limited := io.LimitReader(body, skipAmount)
	pause := false
	cancel := false
	for dlAmount < skipAmount {
		pause, cancel = checkPause(ctrlChan, pause)
		if cancel {
			return CANCELED, nil
		}
		if pause {
			time.Sleep(time.Second)
		} else {
			n, err := limited.Read(buf)
			if err == io.EOF {
				return SUCCESS, nil
			} else if err != nil {
				return ERROR, err
			}
			dlAmount += int64(n)
			updateChan <- ProgressUpdate{id, PROGRESS, dlAmount, nil}
		}
	}
	return SUCCESS, nil
}

func downloadFile(updateChan chan ProgressUpdate,
				  ctrlChan chan int,
				  id int, body io.Reader,
				  out *os.File, initSize int64) (int, error) {
	dlAmount := initSize
	pause := false
	cancel := false
	for {
		pause, cancel = checkPause(ctrlChan, pause)
		if cancel {
			return CANCELED, nil
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
				return SUCCESS, nil
			} else if err != nil {
				return ERROR, err
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
			finishDownload(updateChan, id, dlreq, SUCCESS, nil)
			return
		}
		updateChan <- ProgressUpdate{id, TOTALSIZE, resp.ContentLength, nil}
		status, err := skipAhead(updateChan, ctrlChan, id,
								 resp.Body, dlreq.initSize)
		if err != nil {
			finishDownload(updateChan, id, dlreq, status, err)
		}
		status, err = downloadFile(updateChan, ctrlChan, id,
								   resp.Body, out, dlreq.initSize)
		finishDownload(updateChan, id, dlreq, status, err)
	} else if resp.StatusCode == http.StatusPartialContent {
		totalSize := resp.ContentLength + dlreq.initSize
		if totalSize == 0 {
			finishDownload(updateChan, id, dlreq, SUCCESS, nil)
			return
		}
		updateChan <- ProgressUpdate{id, TOTALSIZE, totalSize, nil}
		status, err := downloadFile(updateChan, ctrlChan, id,
									resp.Body, out, dlreq.initSize)
		finishDownload(updateChan, id, dlreq, status, err)
	} else {
		err := errors.New(resp.Status)
		updateChan <- ProgressUpdate{id, ERROR, 0, err}
		return
	}
}
