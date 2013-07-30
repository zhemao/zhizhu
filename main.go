package main

import (
	"fmt"
	"os"
	"time"
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

const oldWeight float64 = 0.8
const newWeight float64 = 1.0 - oldWeight

func handleProgressUpdate(update ProgressUpdate, statii *[]DownloadStatus) {
	switch update.messType {
	case SUCCESS:
		fname := (*statii)[update.id].fname
		displayPrintf(update.id + 1, "%s finished downloading\n", fname)
		(*statii)[update.id].done = true
	case ERROR:
		displayPrintln(update.id + 1, update.err)
	case CANCELED:
		displayPrintln(update.id + 1, "Canceled")
		(*statii)[update.id].done = true
	case TOTALSIZE:
		(*statii)[update.id].totalAmount = update.amount
		(*statii)[update.id].lastUpdate = time.Now().UnixNano()
		displayProgress(update.id, &((*statii)[update.id]))
	case PROGRESS:
		status := &((*statii)[update.id])
		curTime := time.Now().UnixNano()
		oldTime := status.lastUpdate
		oldAmount := status.dlAmount

		bps := float64(update.amount - oldAmount) / float64(curTime - oldTime) * 1e9
		oldSpeed := status.avgSpeed
		newSpeed := oldWeight * oldSpeed + newWeight * bps

		status.dlAmount = update.amount
		status.lastUpdate = curTime
		status.avgSpeed = newSpeed

		displayProgress(update.id, &((*statii)[update.id]))
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
	ctrlChan := make([]chan int, len(requests))
	statii := make([]DownloadStatus, len(requests))

	defer cleanupReqFile(reqFileName, &statii)

	err = termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	displayString(0, 0, fmt.Sprintf("Zhizhu Download Manager v%s\n", version))
	initSelector()

	for i, dlreq := range requests {
		displayPrintf(i + 1, "Starting download of %s\n", dlreq.basename)
		statii[i] = DownloadStatus{dlreq.url, dlreq.basename, 0, 0,
									time.Now().UnixNano(), 0.0, false}
		ctrlChan[i] = make(chan int, 1)
		go runDownload(updateChan, ctrlChan[i], i, dlreq)
	}
	go listenKeyEvents(keyEventChan)

	termbox.Flush()

	for {
		select {
		case update := <-updateChan:
			handleProgressUpdate(update, &statii)
		case event := <-keyEventChan:
			if handleKeyEvent(event, &ctrlChan) {
				return
			}
		}

		termbox.Flush()
	}
}
