package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)

var _ = net.Listen
var _ = os.Exit

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	fmt.Println("Listening on 0.0.0.0:6379...")

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	fmt.Println("Accepted a connection from:", conn.RemoteAddr())

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client disconnected:", conn.RemoteAddr())
			} else {
				fmt.Println("Read error: ", err)
			}
			break
		}

		fmt.Printf("Received: %q\n", line)

		conn.Write([]byte("+PONG\r\n"))
		fmt.Println("Sent: +PONG")
		continue
	}
}
