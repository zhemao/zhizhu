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
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s request-file\n", os.Args[0])
		os.Exit(-1)
	}

	defer exitProgram()

	reqFileName := os.Args[1]
	requests := loadRequests(reqFileName)

	updateChan := make(chan ProgressUpdate)

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	displayPrintf(0, "Zhizhu Download Manager v%s\n", version)

	for i, dlreq := range requests {
		go runDownload(updateChan, i, &dlreq)
	}
	go listenKeyEvents(updateChan)

	totalAmounts := make([]int64, len(requests))

	for {
		update := <-updateChan

		//termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		switch update.messType {
		case SUCCESS:
			outfname := requests[update.id].outfname
			actualfname := requests[update.id].actualfname
			os.Rename(outfname, actualfname)
			displayPrintf(update.id, "%s finished downloading\n", actualfname)
			return
		case ERROR:
			displayPrintln(update.id + 1, update.err)
			os.Exit(-1)
		case TOTALSIZE:
			totalAmounts[update.id] = update.amount
		case PROGRESS:
			displayProgress(update.id + 1, update.amount, totalAmounts[update.id])
		case QUIT:
			return
		}
		termbox.Flush()
	}
}
