package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const inputFilePath = "messages.txt"

func main() {
	f, err := os.Open("messages.txt")
	if err != nil {
		log.Fatalf("could not open %s: %s\n", inputFilePath, err)
	}
	defer f.Close()

	fmt.Printf("reading data from %s\n", inputFilePath)
	fmt.Print("==============================\n")
	var allfile = ""
	for {
		b := make([]byte, 8, 8)
		_, err := f.Read(b)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				log.Fatalf("encountered error reading: %s\n", err)
			}
		}
		allfile += string(b)
	}
	lines := strings.Split(allfile, "\n")
	for _, line := range lines {
		fmt.Printf("read: %s\n", line)
	}
}
