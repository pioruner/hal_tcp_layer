package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"
)

func main() {
	ln, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Listening on port 8000")
	for {
		rand.Seed(time.Now().UnixNano())
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			//log.Warning(err)
		}
		if len(message) > 0 {
			fmt.Print("Message Received:", string(message))
			newmessage := strings.ToUpper(message)
			conn.Write([]byte(newmessage + string(rand.Intn(100)) + "\n"))
			//conn.Close()
		}
	}
}
