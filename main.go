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

func main () {
	defer exitProgram()

	reqFileName := os.ExpandEnv("$HOME/.zhizhu_requests.txt")
	requests, err := loadRequests(reqFileName)
	if err != nil {
		panic(err)
	}

	updateChan := make(chan ProgressUpdate)
	statii := make([]DownloadStatus, len(requests))
	finished := make([]bool, len(requests))

	defer cleanupReqFile(reqFileName, &requests, &finished)

	err = termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	displayPrintf(0, "Zhizhu Download Manager v%s\n", version)

	for i, dlreq := range requests {
		displayPrintf(i + 1, "Starting download of %s\n", dlreq.basename)
		statii[i] = DownloadStatus{dlreq.url, dlreq.basename, 0, 0}
		finished[i] = false
		go runDownload(updateChan, i, dlreq)
	}
	go listenKeyEvents(updateChan)

	termbox.Flush()

	for {
		update := <-updateChan

		switch update.messType {
		case SUCCESS:
			fname := statii[update.id].fname
			displayPrintf(update.id + 1, "%s finished downloading\n", fname)
			finished[update.id] = true
		case ERROR:
			displayPrintln(update.id + 1, update.err)
			os.Exit(-1)
		case TOTALSIZE:
			statii[update.id].totalAmount = update.amount
		case PROGRESS:
			statii[update.id].dlAmount = update.amount
			displayProgress(update.id + 1, &(statii[update.id]))
		case QUIT:
			return
		}
		termbox.Flush()
	}
}
