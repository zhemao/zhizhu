package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"os"
	"time"
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

const oldWeight float64 = 0.6
const newWeight float64 = 1.0 - oldWeight

func handleProgressUpdate(update ProgressUpdate, statii *[]DownloadStatus) {
	switch update.messType {
	case SUCCESS:
		fname := (*statii)[update.id].fname
		displayPrintf(update.id+1, "%s finished downloading\n", fname)
		(*statii)[update.id].done = true
	case ERROR:
		displayPrintln(update.id+1, update.err)
	case CANCELED:
		displayPrintln(update.id+1, "Canceled")
		(*statii)[update.id].done = true
	case TOTALSIZE:
		(*statii)[update.id].totalAmount = update.amount
		displayProgress(update.id, &((*statii)[update.id]))
	case PROGRESS:
		status := &((*statii)[update.id])
		status.dlAmount = update.amount
		displayProgress(update.id, status)
	}
}

func trackDownloadSpeed(statii *[]DownloadStatus) {
	lastSizes := make([]int64, len(*statii))
	tickChan := time.Tick(time.Second)

	for {
		select {
		case <-tickChan:
			for i, _ := range lastSizes {
				status := &((*statii)[i])
				newSpeed := status.dlAmount - lastSizes[i]
				status.avgSpeed = int64(oldWeight*float64(status.avgSpeed) +
					newWeight*float64(newSpeed))
				lastSizes[i] = status.dlAmount
			}
		}
	}
}

func main() {
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
	initKeyInput()

	for i, dlreq := range requests {
		displayPrintf(i+1, "Starting download of %s\n", dlreq.basename)
		statii[i] = DownloadStatus{dlreq.url, dlreq.basename, 0, 0, 0, false}
		ctrlChan[i] = make(chan int, 1)
		go runDownload(updateChan, ctrlChan[i], i, dlreq)
	}
	go listenKeyEvents(keyEventChan)
	go trackDownloadSpeed(&statii)

	termbox.Flush()

	for {
		select {
		case update := <-updateChan:
			handleProgressUpdate(update, &statii)
		case event := <-keyEventChan:
			if handleKeyEvent(event, updateChan, &ctrlChan, &requests, &statii) {
				return
			}
		}

		termbox.Flush()
	}
}
