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

func allDone(statii *[]DownloadStatus) bool {
	for _, stat := range *statii {
		if !stat.done {
			return false
		}
	}
	return true
}

func main () {
	defer exitProgram()

	reqFileName := os.ExpandEnv("$HOME/.zhizhu/requests.txt")
	requests, err := loadRequests(reqFileName)
	if err != nil {
		panic(err)
	}

	updateChan := make(chan ProgressUpdate)

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	displayPrintf(0, "Zhizhu Download Manager v%s\n", version)

	statii := make([]DownloadStatus, len(requests))

	for i, dlreq := range requests {
		go runDownload(updateChan, i, dlreq)
		displayPrintf(i + 1, "Starting download of %s\n", dlreq.actualfname)
		statii[i] = DownloadStatus{dlreq.url, dlreq.actualfname, 0, 0, false}
	}
	go listenKeyEvents(updateChan)


	for {
		update := <-updateChan

		switch update.messType {
		case SUCCESS:
			actualfname := requests[update.id].actualfname
			displayPrintf(update.id, "%s finished downloading\n", actualfname)
			statii[update.id].done = true
			if allDone(&statii) {
				return
			}
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
