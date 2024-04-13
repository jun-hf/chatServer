package main

import (
	"fmt"
	"log"
	"net"
)

type client chan<- string

var (
	entering = make(chan client)
	leaving = make(chan client)
	messages = make(chan string) // all incoming client message
)



func main() {
	listen, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
	go broadcaster()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Printf("failed to accept connection %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(con net.Conn) {
	fmt.Fprintln(con, "Hello")
}