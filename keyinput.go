package main

import (
	"github.com/nsf/termbox-go"
)

var selectId int;
const cursorCol int = 0;
const indicatorCol int = 1;

func listenKeyEvents(channel chan termbox.Event) {
	for {
		event := termbox.PollEvent()
		channel <- event;
	}
}

func handleSpecialKey(key termbox.Key) bool {
	switch key {
	case termbox.KeyEsc:
		return true
	}
	return false
}

func drawAtCursor(ch rune) {
	termbox.SetCell(cursorCol, selectId + 1, ch,
					termbox.ColorDefault, termbox.ColorDefault)
}

func drawIndicator(ch rune) {
	termbox.SetCell(indicatorCol, selectId + 1, ch,
					termbox.ColorDefault, termbox.ColorDefault)
}

func initSelector() {
	selectId = 0
	drawAtCursor('+')
}

func handleKeyEvent(event termbox.Event,
					ctrlChan *[]chan int) bool {
	if event.Type == termbox.EventKey {
		switch event.Ch {
		case 0:
			return handleSpecialKey(event.Key)
		case 'q':
			return true
		case 'j':
			if selectId < len(*ctrlChan) - 1 {
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
			drawIndicator('X')
		}
	}
	return false
}
