package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

type client chan<- string

var (
	entering = make(chan client)
	leaving = make(chan client)
	messages = make(chan string) // all incoming client message
)

func broadcaster() {
	clients := make(map[client]bool)
	for {
		select {
		case msg := <- messages:
			for cli := range clients {
				cli <- msg
			}
		case cli := <- entering:
			clients[cli] = true
		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}

// change client struct to get the name from the client
// add timer to each client request
// create a new func to reader client's response

func clientWriter(conn net.Conn, ch <-chan string) {
	for ms := range ch {
		fmt.Fprintln(conn, ms)
	}
}

func clientReader(conn net.Conn, clientInput chan<- string) {
	input := bufio.NewScanner(conn)
	for input.Scan() {
		clientInput <- input.Text()
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	clientWriteCh := make(chan string, 10)
	clientInputCh := make(chan string)
	go clientWriter(conn, clientWriteCh)
	go clientReader(conn, clientInputCh)

	who := conn.RemoteAddr().String()
	clientWriteCh <- fmt.Sprintf("You are %v", who)
	messages <- fmt.Sprintf("%v has arrived", who)
	entering <- clientWriteCh

	loop:
	for {
		select {
		case input := <-clientInputCh:
			messages <- input
		case <-time.After(time.Second *20):
			break loop
		}
	}
	leaving <- clientWriteCh
	messages <- fmt.Sprintf("%v has left", who)
}

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