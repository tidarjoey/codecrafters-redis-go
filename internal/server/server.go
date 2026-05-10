package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

type Server struct {
	store *store.Store
}

func New(s *store.Store) *Server {
	return &Server{store: s}
}

func (s *Server) Listen(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	fmt.Println("Listening on", addr)
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Accepted a connection from:", conn.RemoteAddr())

	r := bufio.NewReader(conn)
	for {
		parts, err := resp.ReadCommand(r)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Client disconnected:", conn.RemoteAddr())
			} else {
				fmt.Println("Read error:", err)
			}
			return
		}
		if len(parts) == 0 {
			continue
		}
		s.dispatch(conn, parts)
	}
}

func (s *Server) dispatch(w io.Writer, parts []string) {
	cmd := strings.ToUpper(parts[0])
	switch cmd {
	case "PING":
		resp.WriteSimple(w, "PONG")
	case "ECHO":
		if len(parts) >= 2 {
			resp.WriteBulk(w, strings.Join(parts[1:], " "))
		} else {
			resp.WriteNullBulk(w)
		}
	case "SET":
		if len(parts) < 3 {
			resp.WriteError(w, "ERR wrong number of arguments for 'set' command")
			return
		}
		key, value := parts[1], parts[2]
		var expiry time.Time
		for i := 3; i+1 < len(parts); i += 2 {
			switch strings.ToUpper(parts[i]) {
			case "PX":
				ms, _ := strconv.ParseInt(parts[i+1], 10, 64)
				expiry = time.Now().Add(time.Duration(ms) * time.Millisecond)
			case "EX":
				sec, _ := strconv.ParseInt(parts[i+1], 10, 64)
				expiry = time.Now().Add(time.Duration(sec) * time.Second)
			}
		}
		s.store.Set(key, value, expiry)
		resp.WriteSimple(w, "OK")
	case "RPUSH":
		if len(parts) < 3 {
			resp.WriteError(w, "ERR wrong number of arguments for 'rpush' command")
			return
		}
		n, err := s.store.RPush(parts[1], parts[2:]...)
		if err != nil {
			resp.WriteError(w, err.Error())
			return
		}
		resp.WriteInteger(w, n)
	case "GET":
		if len(parts) < 2 {
			resp.WriteError(w, "ERR wrong number of arguments for 'get' command")
			return
		}
		val, ok, err := s.store.Get(parts[1])
		if err != nil {
			resp.WriteError(w, err.Error())
			return
		}
		if ok {
			resp.WriteBulk(w, val)
		} else {
			resp.WriteNullBulk(w)
		}
	default:
		resp.WriteSimple(w, "OK")
	}
}
