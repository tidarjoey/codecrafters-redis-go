package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
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

	for {
		conn, err := l.Accept()

		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Accepted a connection from:", conn.RemoteAddr())

	reader := bufio.NewReader(conn)
	for {
		b, err := reader.ReadByte()

		if err != nil {
			if err == io.EOF {
				fmt.Println("Client disconnected:", conn.RemoteAddr())
			} else {
				fmt.Println("Read error:", err)
			}
			break
		}

		if b == '*' {
			countLine, _ := reader.ReadBytes('\n')
			count, _ := strconv.Atoi(strings.TrimSpace(string(countLine)))
			parts := make([]string, 0, count)

			for i := 0; i < count; i++ {
				prefix, _ := reader.ReadByte()

				if prefix != '$' {
					fmt.Println("Unexpected prefix:", prefix)
					break
				}

				lenLine, _ := reader.ReadBytes('\n')
				n, _ := strconv.Atoi(strings.TrimSpace(string(lenLine)))

				buf := make([]byte, n+2)
				_, err := io.ReadFull(reader, buf)

				if err != nil {
					fmt.Println("ReadFull error:", err)
					break
				}

				parts = append(parts, string(buf[:n]))
			}

			fmt.Println("Parsed RESP command:", parts)

			conn.Write([]byte("+PONG\r\n"))
			continue
		}

		rest, _ := reader.ReadBytes('\n')
		line := string(append([]byte{b}, rest...))
		line = strings.TrimRight(line, "\r\n")
		fmt.Println("Parsed inline command:", line)

		conn.Write([]byte("+PONG\r\n"))
	}
}
