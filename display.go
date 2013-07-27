package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
)

const (
	KB = 1024
	MB = 1024 * 1024
	GB = 1024 * 1024 * 1024
)

/*
 * Converts the integer size in bytes to a human-readable string
 */
func displaySize(size int64) string {
	if size < KB {
		return fmt.Sprintf("%dB", size)
	}
	if size < MB {
		return fmt.Sprintf("%.2fK", float64(size) / KB)
	}
	if size < GB {
		return fmt.Sprintf("%.2fM", float64(size) / MB)
	}
	return fmt.Sprintf("%.2fG", float64(size) / GB)
}

func displayString(row int, str string) {
	for col, ch := range str {
		termbox.SetCell(col, row, ch,
						termbox.ColorDefault, termbox.ColorDefault)
	}
	width, _ := termbox.Size()
	for col := len(str); col < width; col++ {
		termbox.SetCell(col, row, ' ', termbox.ColorDefault, termbox.ColorDefault)
	}
}

func displayPrintln(row int, obj interface{}) {
	rowStr := fmt.Sprintln(obj)
	displayString(row, rowStr)
}

func displayPrintf(row int, format string, args...interface{}) {
	rowStr := fmt.Sprintf(format, args...)
	displayString(row, rowStr)
}

func displayProgress(id int, status *DownloadStatus) {
	percent := status.dlAmount * 100 / status.totalAmount
	displayPrintf(id, "%s %d%% | %s of %s downloaded\n",
				  status.fname,
				  percent,
				  displaySize(status.dlAmount),
				  displaySize(status.totalAmount))
}
