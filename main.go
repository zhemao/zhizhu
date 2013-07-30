package main

import (
	"fmt"
	"os"
	"github.com/nsf/termbox-go"
)

const version string = "0.1.0"

func exitProgram() {
	if r := recover(); r != nil {
		fmt.Println(r)
		os.Exit(-1)
	} else {
		os.Exit(0)
	}
}

func handleProgressUpdate(update ProgressUpdate, statii *[]DownloadStatus) {
	switch update.messType {
	case SUCCESS:
		fname := (*statii)[update.id].fname
		displayPrintf(update.id + 1, 3, "%s finished downloading\n", fname)
		(*statii)[update.id].done = true
	case ERROR:
		displayPrintln(update.id + 1, 3, update.err)
		os.Exit(-1)
	case TOTALSIZE:
		(*statii)[update.id].totalAmount = update.amount
	case PROGRESS:
		(*statii)[update.id].dlAmount = update.amount
		displayProgress(update.id + 1, &((*statii)[update.id]))
	}
}

func main () {
	defer exitProgram()

	reqFileName := os.ExpandEnv("$HOME/.zhizhu_requests.txt")
	requests, err := loadRequests(reqFileName)
	if err != nil {
		panic(err)
	}

	updateChan := make(chan ProgressUpdate)
	keyEventChan := make(chan termbox.Event)
	statii := make([]DownloadStatus, len(requests))

	defer cleanupReqFile(reqFileName, &statii)

	err = termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	displayPrintf(0, "Zhizhu Download Manager v%s\n", version)

	for i, dlreq := range requests {
		displayPrintf(i + 1, "Starting download of %s\n", dlreq.basename)
		statii[i] = DownloadStatus{dlreq.url, dlreq.basename, 0, 0, false}
		go runDownload(updateChan, i, dlreq)
	}
	go listenKeyEvents(keyEventChan)

	termbox.Flush()

	for {
		select {
		case update := <-updateChan:
			handleProgressUpdate(update, &statii)
		case event := <-keyEventChan:
			if handleKeyEvent(event, &requests, &statii) {
				return
			}
		}

		termbox.Flush()
	}
}
