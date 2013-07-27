package main

import (
	"github.com/nsf/termbox-go"
)

func listenKeyEvents(channel chan ProgressUpdate) {
	for {
		event := termbox.PollEvent()
		if event.Type == termbox.EventKey {
			switch event.Key {
			case termbox.KeyEsc, termbox.KeyCtrlQ:
				channel <- ProgressUpdate{-1, QUIT, 0, nil}
			}
		}
	}
}
