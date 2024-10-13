package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Data struct {
	CMD    uint8
	Keys   []string
	Values []float64
}

func (d Data) print(buf []byte, lend int) {
	err := json.Unmarshal(buf[:lend], &d)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("CMD:", d.CMD)
	fmt.Print("Keys: ")
	for _, v := range d.Keys {
		fmt.Print(string(v) + ",")
	}
	fmt.Println("")
	fmt.Print("Values: ")
	for _, v := range d.Values {
		s := fmt.Sprintf("%f,", v)
		fmt.Print(s)
	}
	fmt.Println("")
}

// Обработка каждого подключения в отдельной горутине
func handleConnections(conn net.Conn, wg *sync.WaitGroup, dd *Data) {
	defer wg.Done()
	// Устанавливаем таймаут на чтение и запись
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	for {
		header := make([]byte, 1024)
		lend, err := conn.Read(header)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error receiving Packet:", err)
			break
		}
		dd.print(header[:lend], lend)
		//printData(header, lend)
		// Пакуем данные и отправляем обратно
		buf := "OK"
		_, err = conn.Write([]byte(buf))
		if err != nil {
			fmt.Println("Error writing response:", err)
			break
		}
		// Сбрасываем таймауты после каждой успешной операции
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	}

	// Сообщаем о закрытии соединения только один раз
	fmt.Println("Closing connection with client...")
	conn.Close()
}

func main() {
	var d Data
	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server is listening...")

	var wg sync.WaitGroup
	sem := make(chan struct{}, 100) // Ограничение на количество одновременных подключений (например, 100)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// Используем семафор для ограничения одновременных подключений
		sem <- struct{}{}
		wg.Add(1)

		go func(conn net.Conn) {
			defer func() { <-sem }()
			handleConnections(conn, &wg, &d)
		}(conn)
	}

	// Ожидаем завершения всех горутин
	wg.Wait()
}
