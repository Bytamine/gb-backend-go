package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	cfg := net.ListenConfig{
		KeepAlive: time.Minute,
	}

	l, err := cfg.Listen(ctx, "tcp", ":9000")
	if err != nil {
		log.Fatal(err)
	}
	wg := &sync.WaitGroup{}
	log.Println("I'm started!")

	serverMessageCh := make(chan string)

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Println(err)
				return
			} else {
				wg.Add(1)
				go handleConn(ctx, conn, wg, serverMessageCh)
			}
		}
	}()

	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			text, err := reader.ReadString('\n')
			if err != nil {
				log.Println(err)
				return
			}
			serverMessageCh <- text
		}
	}()

	<-ctx.Done()

	log.Println("done")
	err = l.Close()
	if err != nil {
		log.Fatal(err)
	}
	wg.Wait()
	log.Println("exit")
}

func handleConn(ctx context.Context, conn net.Conn, wg *sync.WaitGroup, serverMessageCh chan string) {
	defer wg.Done()
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println(err)
		}
	}(conn)
	// Every second send current time to clients
	tck := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-tck.C:
			_, err := fmt.Fprintf(conn, "now: %s\n", t)
			if err != nil {
				log.Println(err)
			}
		case msg := <-serverMessageCh:
			_, err := fmt.Fprintf(conn, "server message: %s\n", msg)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
