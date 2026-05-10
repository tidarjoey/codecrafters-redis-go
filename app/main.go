package main

import (
	"fmt"
	"os"

	"github.com/codecrafters-io/redis-starter-go/internal/server"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

func main() {
	s := server.New(store.New())
	if err := s.Listen("0.0.0.0:6379"); err != nil {
		fmt.Println("Server error:", err)
		os.Exit(1)
	}
}
