package main

import (
	"fmt"
	"os"
)

const version string = "0.1.0"

func main () {
	fmt.Printf("Zhizhu Download Manager v%s\n", version)
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s url output-file\n", os.Args[0])
		os.Exit(-1)
	}

	url := os.Args[1]
	actualfname := os.Args[2]
	outfname := actualfname + ".part"

	file, err := os.Open(outfname)

	initSize := int64(0)

	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println(err)
			os.Exit(-1)
		}
		file, err = os.Create(outfname)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	} else {
		initSize, err = file.Seek(0, os.SEEK_END)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		file.Close();
		file, err = os.OpenFile(outfname, os.O_APPEND | os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	}

	if initSize > 0 {
		fmt.Printf("Resuming download at %d bytes\n", initSize)
	}

	updateChan := make(chan ProgressUpdate)
	totalAmount := int64(0)

	go runDownload(updateChan, 0 ,url, file, initSize)

	for {
		update := <-updateChan
		switch update.messType {
		case SUCCESS:
			os.Rename(outfname, actualfname)
			fmt.Printf("%s finished downloading\n", actualfname)
			return
		case ERROR:
			fmt.Println(update.err)
			os.Exit(-1)
		case TOTALSIZE:
			totalAmount = update.amount
		case PROGRESS:
			fmt.Printf("%d of %d bytes downloaded\n", update.amount, totalAmount)
		}
	}
}
