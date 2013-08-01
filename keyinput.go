package main

import (
	"github.com/nsf/termbox-go"
	"path"
)

var selectId int
var inputCol int

const cursorCol int = 0
const indicatorCol int = 1
const bufferSize int = 128

var insertMode bool
var inputBuffer []rune

func listenKeyEvents(channel chan termbox.Event) {
	for {
		event := termbox.PollEvent()
		channel <- event
	}
}

func exitInsertMode() {
	insertMode = false
	width, height := termbox.Size()
	for col := 0; col < width; col++ {
		termbox.SetCell(col, height-1, ' ',
			termbox.ColorDefault, termbox.ColorDefault)
	}
	inputCol = 0
	termbox.HideCursor()
}

func handleSpecialKey(key termbox.Key) string {
	switch key {
	case termbox.KeyEsc:
		exitInsertMode()
	case termbox.KeyEnter:
		inputStr := string(inputBuffer[:inputCol])
		exitInsertMode()
		return inputStr
	}
	return ""
}

func drawAtCursor(ch rune) {
	termbox.SetCell(cursorCol, selectId+1, ch,
		termbox.ColorDefault, termbox.ColorDefault)
}

func drawIndicator(ch rune) {
	termbox.SetCell(indicatorCol, selectId+1, ch,
		termbox.ColorDefault, termbox.ColorDefault)
}

func initKeyInput() {
	selectId = 0
	drawAtCursor('+')
	insertMode = false
	inputCol = 0
	inputBuffer = make([]rune, bufferSize)
}

func handleNonInsertMode(event termbox.Event,
	ctrlChan *[]chan int,
	statii *[]DownloadStatus) bool {
	switch event.Ch {
	case 'q':
		return true
	case 'j':
		if selectId < len(*ctrlChan)-1 {
			drawAtCursor(' ')
			selectId += 1
			drawAtCursor('+')
		}
	case 'k':
		if selectId > 0 {
			drawAtCursor(' ')
			selectId -= 1
			drawAtCursor('+')
		}
	case 'p':
		(*ctrlChan)[selectId] <- PAUSE
		drawIndicator('P')
	case 'r':
		(*ctrlChan)[selectId] <- RESUME
		drawIndicator(' ')
	case 'c':
		(*ctrlChan)[selectId] <- CANCEL
		(*statii)[selectId].done = true
		drawIndicator('X')
	case 'a':
		_, height := termbox.Size()
		termbox.SetCursor(inputCol, height-1)
		insertMode = true
	}
	return false
}

func handleInsertMode(event termbox.Event,
	updateChan chan ProgressUpdate,
	ctrlChan *[]chan int,
	requests *[]DownloadRequest,
	statii *[]DownloadStatus) {
	switch event.Ch {
	case 0:
		url := handleSpecialKey(event.Key)
		if url != "" {
			basename := path.Base(url)
			id := len(*requests)
			dlreq, err := makeRequest(url, basename)
			_, height := termbox.Size()
			if err != nil {
				displayPrintln(height-2, err)
				return
			}
			displayPrintln(height-2, "")
			*requests = append(*requests, dlreq)
			*ctrlChan = append(*ctrlChan, make(chan int, 1))
			*statii = append(*statii, DownloadStatus{url, basename, 0, 0, 0, false})
			displayPrintf(id+1, "Starting download of %s\n", basename)
			go runDownload(updateChan, (*ctrlChan)[id], id, dlreq)
		}
	default:
		if inputCol < bufferSize {
			_, height := termbox.Size()
			inputBuffer[inputCol] = event.Ch
			termbox.SetCell(inputCol, height-1, event.Ch,
				termbox.ColorDefault, termbox.ColorDefault)
			inputCol = inputCol + 1
			termbox.SetCursor(inputCol, height-1)
		}
	}
}

func handleKeyEvent(event termbox.Event,
	updateChan chan ProgressUpdate,
	ctrlChan *[]chan int,
	requests *[]DownloadRequest,
	statii *[]DownloadStatus) bool {
	if event.Type == termbox.EventKey {
		if insertMode {
			handleInsertMode(event, updateChan, ctrlChan, requests, statii)
		} else {
			return handleNonInsertMode(event, ctrlChan, statii)
		}
	}
	return false
}
