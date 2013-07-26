package main

import (
	"fmt"
	"os"
	"io"
	"errors"
	"net/http"
)

const version string = "0.1.0"

type ProgressUpdate struct {
	id int
	messType int
	amount int64
	err error
}

const (
	ERROR = iota
	SUCCESS = iota
	TOTALSIZE = iota
	PROGRESS = iota
	SKIP = iota
	HTTPSTATUS = iota
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
	buf := make([]byte, CHUNK_SIZE)
	dlAmount := initSize
	for {
		n, err := body.Read(buf)
		if n > 0 {
			dlAmount += int64(n)
			_, err := out.Write(buf[:n])
			if err != nil {
				channel <- ProgressUpdate{id, ERROR, 0, err}
				return
			}
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

func runDownload(channel chan ProgressUpdate, id int, url string, out *os.File, 
					initSize int64) {
	defer out.Close()
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
		channel <- ProgressUpdate{id, ERROR, 0, err}
		return
	}

	if initSize > 0 {
		req.Header.Add("Range", fmt.Sprintf("bytes=%d-", initSize))
	}

	resp, err := client.Do(req)
	if err != nil {
		channel <- ProgressUpdate{id, ERROR, 0, err}
		return
	}

	channel <- ProgressUpdate{id, HTTPSTATUS, int64(resp.StatusCode), nil}
	if resp.StatusCode == http.StatusOK {
		channel <- ProgressUpdate{id, TOTALSIZE, resp.ContentLength, nil}
		channel <- ProgressUpdate{id, SKIP, initSize, nil}
		skipAhead(channel, id, resp.Body, initSize)
		downloadFile(channel, id, resp.Body, out, initSize)
	} else if resp.StatusCode == http.StatusPartialContent {
		totalSize := resp.ContentLength + initSize
		channel <- ProgressUpdate{id, TOTALSIZE, totalSize, nil}
		downloadFile(channel, id, resp.Body, out, initSize)
	} else {
		channel <- ProgressUpdate{id, ERROR, 0, errors.New(resp.Status)}
	}
}

func main () {
	fmt.Printf("Zhizhu Download Manager v%s\n", version)
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s url output-file\n", os.Args[0])
		os.Exit(-1)
	}

	url := os.Args[1]
	outfname := os.Args[2]

	file, err := os.Open(outfname)

	initSize := int64(0)

	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println(err)
			os.Exit(-1)
		}
		file, err = os.Create(outfname)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	} else {
		initSize, err = file.Seek(0, os.SEEK_END)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		file.Close();
		file, err = os.OpenFile(outfname, os.O_APPEND | os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	}

	if initSize > 0 {
		fmt.Printf("Resuming download at %d bytes\n", initSize)
	}

	updateChan := make(chan ProgressUpdate)
	totalAmount := int64(0)

	go runDownload(updateChan, 0 ,url, file, initSize)

	for {
		update := <-updateChan
		switch update.messType {
		case SUCCESS:
			fmt.Printf("%s finished downloading\n", outfname)
			return
		case ERROR:
			fmt.Println(update.err)
			os.Exit(-1)
		case HTTPSTATUS:
			fmt.Printf("HTTP status %d\n", update.amount)
		case TOTALSIZE:
			totalAmount = update.amount
		case SKIP:
			fmt.Printf("Skipping ahead by %d\n", update.amount)
		case PROGRESS:
			fmt.Printf("%d of %d bytes downloaded\n", update.amount, totalAmount)
		}
	}
}
