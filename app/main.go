package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	_ = net.Listen
	_ = os.Exit
)

// in-memory store protected by a mutex for concurrent goroutines
var (
	store   = make(map[string]string)
	storeMu sync.RWMutex
)

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

		// RESP Array form
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

			if len(parts) == 0 {
				continue
			}

			cmd := strings.ToUpper(parts[0])
			switch cmd {
			case "PING":
				_, _ = conn.Write([]byte("+PONG\r\n"))
			case "ECHO":
				if len(parts) >= 2 {
					payload := strings.Join(parts[1:], " ")
					_ = writeBulk(conn, payload)
				} else {
					_ = writeNullBulk(conn)
				}
			case "SET":
				if len(parts) >= 3 {
					key := parts[1]
					value := parts[2]
					storeMu.Lock()
					store[key] = value
					storeMu.Unlock()
					_, _ = conn.Write([]byte("+OK\r\n"))
				} else {
					_, _ = conn.Write([]byte("-ERR wrong number of arguments for 'set' command\r\n"))
				}
			case "GET":
				if len(parts) >= 2 {
					key := parts[1]
					storeMu.RLock()
					val, ok := store[key]
					storeMu.RUnlock()
					if ok {
						_ = writeBulk(conn, val)
					} else {
						_ = writeNullBulk(conn)
					}
				} else {
					_, _ = conn.Write([]byte("-ERR wrong number of arguments for 'get' command\r\n"))
				}
			default:
				_, _ = conn.Write([]byte("+OK\r\n"))
			}
			continue
		}

		// Inline form
		rest, _ := reader.ReadBytes('\n')
		line := string(append([]byte{b}, rest...))
		line = strings.TrimRight(line, "\r\n")
		fmt.Println("Parsed inline command:", line)

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		switch strings.ToUpper(fields[0]) {
		case "PING":
			_, _ = conn.Write([]byte("+PONG\r\n"))
		case "ECHO":
			if len(fields) >= 2 {
				payload := strings.Join(fields[1:], " ")
				_ = writeBulk(conn, payload)
			} else {
				_ = writeNullBulk(conn)
			}
		case "SET":
			if len(fields) >= 3 {
				key := fields[1]
				value := fields[2]
				storeMu.Lock()
				store[key] = value
				storeMu.Unlock()
				_, _ = conn.Write([]byte("+OK\r\n"))
			} else {
				_, _ = conn.Write([]byte("-ERR wrong number of arguments for 'set' command\r\n"))
			}
		case "GET":
			if len(fields) >= 2 {
				key := fields[1]
				storeMu.RLock()
				val, ok := store[key]
				storeMu.RUnlock()
				if ok {
					_ = writeBulk(conn, val)
				} else {
					_ = writeNullBulk(conn)
				}
			} else {
				_, _ = conn.Write([]byte("-ERR wrong number of arguments for 'get' command\r\n"))
			}
		default:
			_, _ = conn.Write([]byte("+OK\r\n"))
		}
	}
}

// writeBulk writes a non-nil bulk string (binary-safe).
func writeBulk(w io.Writer, s string) error {
	_, err := fmt.Fprintf(w, "$%d\r\n%s\r\n", len(s), s)
	return err
}

// writeNullBulk writes a nil bulk string.
func writeNullBulk(w io.Writer) error {
	_, err := io.WriteString(w, "$-1\r\n")
	return err
}
