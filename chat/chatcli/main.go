package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter a nickname: ")
	nickname, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.Dial("tcp", "localhost:8001")
	if err != nil {
		log.Fatal(err)
	}

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(conn)

	// Send nickname to server
	_, err = fmt.Fprintf(conn, nickname)
	if err != nil {
		log.Fatal(err)
	}

	// Read messages from server
	go func() {
		_, err := io.Copy(os.Stdout, conn)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Send messages to server
	_, err = io.Copy(conn, os.Stdin)
	if err != nil {
		log.Fatal(err)
	} // until you send ^Z
	fmt.Printf("%s: exit", conn.LocalAddr())
}
