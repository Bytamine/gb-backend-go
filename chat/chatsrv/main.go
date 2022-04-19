package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

type client chan<- string

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	cfg := net.ListenConfig{
		KeepAlive: time.Minute,
	}
	listener, err := cfg.Listen(ctx, "tcp", "localhost:8001")
	if err != nil {
		log.Fatal(err)
	}

	go broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	ch := make(chan string)
	go clientWriter(conn, ch)

	// Get the client's nickname
	input := bufio.NewScanner(conn)
	input.Scan()
	who := input.Text()

	ch <- "You are " + who
	messages <- who + " has arrived"
	entering <- ch

	log.Println(who + " has arrived")

	for input.Scan() {
		messages <- who + ": " + input.Text()
	}
	leaving <- ch
	messages <- who + " has left"
	err := conn.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		_, err := fmt.Fprintln(conn, msg)
		if err != nil {
			log.Print(err)
		}
	}
}

func broadcaster() {
	clients := make(map[client]bool)
	for {
		select {
		case msg := <-messages:
			for cli := range clients {
				cli <- msg
			}
		case cli := <-entering:
			clients[cli] = true

		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}
