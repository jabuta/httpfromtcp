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
	fmt.Printf("reading data from %s\n", inputFilePath)
	fmt.Print("==============================\n")
	linesChanel := getLinesChannel(f)
	for line := range linesChanel {
		fmt.Printf("read: %s\n", line)
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	linesChan := make(chan string)

	go func() {
		defer close(linesChan)
		defer f.Close()
		currentLine := ""
		for {
			buffer := make([]byte, 8, 8)
			n, err := f.Read(buffer)
			if err != nil {
				if currentLine != "" {
					linesChan <- currentLine
					currentLine = ""
				}
				if errors.Is(err, io.EOF) {
					break
				}
			}
			str := string(buffer[:n])
			parts := strings.Split(str, "\n")
			for i := 0; i < len(parts)-1; i++ {
				linesChan <- currentLine + parts[i]
				currentLine = ""
			}
			currentLine += parts[len(parts)-1]
		}

	}()
	return linesChan
}
