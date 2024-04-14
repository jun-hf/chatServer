package main

import (
	"bufio"
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

func handleConnection(conn net.Conn) {
	defer conn.Close()
	ch := make(chan string)
	go clientWriter(conn, ch)
	who := conn.RemoteAddr().String()
	ch <- fmt.Sprintf("You are %v", who)
	messages <- fmt.Sprintf("%v has arrived", who)
	entering <- ch

	input := bufio.NewScanner(conn)
	for input.Scan() {
		messages <- fmt.Sprintf("%v : %v", who, input.Text())
	}
	leaving <- ch
	messages <- fmt.Sprintf("%v has left", who)
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for ms := range ch {
		fmt.Fprintln(conn, ms)
	}
}