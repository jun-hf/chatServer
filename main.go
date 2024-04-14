package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

type client struct {
	c    chan string
	name string
}

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string) // all incoming client message
)

func broadcaster() {
	clients := make(map[client]bool)
	for {
		select {
		case msg := <-messages:
			for cli := range clients {
				select {
				case cli.c <- msg:
				default:
					// skip client 
				}
			}
		case cli := <-entering:
			clients[cli] = true
		case cli := <-leaving:
			delete(clients, cli)
			close(cli.c)
		}
	}
}

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

	timeout := time.Minute * 10
	clientName := "no name"
	select {
	case name := <-clientInputCh:
		clientName = name
	case <-time.After(timeout):
		conn.Close()
		return
	}

	newClient := client{clientWriteCh, clientName}
	clientWriteCh <- fmt.Sprintf("Welcome %v", clientName)
	messages <- fmt.Sprintf("%v has arrived", clientName)
	entering <- newClient

	timer := time.NewTimer(timeout)
loop:
	for {
		select {
		case input := <-clientInputCh:
			messages <- fmt.Sprintf("%v: %v", clientName, input)
			timer.Reset(timeout)
		case <-timer.C:
			break loop
		}
	}
	leaving <- newClient
	messages <- fmt.Sprintf("%v has left", clientName)
}

func main() {
	listen, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Starting chat server")
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
