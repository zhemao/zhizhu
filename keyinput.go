package main

import (
	"github.com/nsf/termbox-go"
)

var select_id int;
const cursor_col int = 0;
const indicator_col int = 1;

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
	termbox.SetCell(cursor_col, select_id + 1, ch,
					termbox.ColorDefault, termbox.ColorDefault)
}

func drawIndicator(ch rune) {
	termbox.SetCell(indicator_col, select_id + 1, ch,
					termbox.ColorDefault, termbox.ColorDefault)
}

func initSelector() {
	select_id = 0
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
			if select_id < len(*ctrlChan) - 1 {
				drawAtCursor(' ')
				select_id += 1
				drawAtCursor('+')
			}
		case 'k':
			if select_id > 0 {
				drawAtCursor(' ')
				select_id -= 1
				drawAtCursor('+')
			}
		case 'p':
			(*ctrlChan)[select_id] <- PAUSE
			drawIndicator('P')
		case 'r':
			(*ctrlChan)[select_id] <- RESUME
			drawIndicator(' ')
		case 'c':
			(*ctrlChan)[select_id] <- CANCEL
			drawIndicator('X')
		}
	}
	return false
}
