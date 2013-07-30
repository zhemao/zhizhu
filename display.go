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

const INDENT int = 3

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

func displayString(row int, startCol int, str string) {
	width, _ := termbox.Size()
	for col, ch := range str {
		if col < width {
			termbox.SetCell(col + startCol, row, ch,
							termbox.ColorDefault, termbox.ColorDefault)
		}
	}
	for col := len(str) + startCol; col < width; col++ {
		termbox.SetCell(col, row, ' ', termbox.ColorDefault, termbox.ColorDefault)
	}
}

func displayPrintln(row int, obj interface{}) {
	rowStr := fmt.Sprintln(obj)
	displayString(row, INDENT, rowStr)
}

func displayPrintf(row int, format string, args...interface{}) {
	rowStr := fmt.Sprintf(format, args...)
	displayString(row, INDENT, rowStr)
}

func displayProgress(id int, status *DownloadStatus) {
	percent := status.dlAmount * 100 / status.totalAmount
	displayPrintf(id + 1, "%s %d%% | %s of %s downloaded | %s / s\n",
				  status.fname,
				  percent,
				  displaySize(status.dlAmount),
				  displaySize(status.totalAmount),
			      displaySize(int64(status.avgSpeed)))
}
