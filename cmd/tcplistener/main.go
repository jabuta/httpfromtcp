package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

const port = ":42069"

func main() {
	l, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("could not listen on port %s: %s\n", port, err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalf("failed to accept connection on port %s: %s", port, err)
		}
		fmt.Printf("connection accepted on %s\n", port)
		fmt.Print("==============================\n")
		linesChanel := getLinesChannel(conn)
		for line := range linesChanel {
			fmt.Printf("read: %s\n", line)
		}
		fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	linesChan := make(chan string)

	go func() {
		defer close(linesChan)
		defer f.Close()
		currentLine := ""
		for {
			buffer := make([]byte, 8)
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
