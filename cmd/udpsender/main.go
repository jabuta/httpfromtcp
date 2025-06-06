package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const address = "localhost:42069"

func main() {
	udpAddres, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Fatalf("failed to resolve UDP address: %s\n", err)
	}
	conn, err := net.DialUDP(udpAddres.Network(), nil, udpAddres)
	if err != nil {
		log.Fatalf("failed to establish UDP connection: %s\n", err)
	}
	defer conn.Close()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">")
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("error reading input: ", err)
		}
		_, err = conn.Write([]byte(message))
		if err != nil {
			fmt.Println("error sending data: ", err)
		}
		fmt.Printf("Message sent: %s", message)
	}
}
