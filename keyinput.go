package main

import (
	"github.com/nsf/termbox-go"
)

func listenKeyEvents(channel chan termbox.Event) {
	for {
		event := termbox.PollEvent()
		channel <- event;
	}
}

func handleKeyEvent(event termbox.Event,
					requests *[]DownloadRequest,
					statii *[]DownloadStatus) bool {
	if event.Type == termbox.EventKey {
		switch event.Key {
		case termbox.KeyEsc:
			return true
		}
	}
	return false
}
