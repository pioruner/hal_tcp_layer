package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Packer struct {
	F1 float64
	F2 float64
}

// Функция для разборки структуры из байт
func unpackFromBytes(data []byte) (Packer, error) {
	var p Packer
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.BigEndian, &p)
	if err != nil {
		return p, err
	}
	return p, nil
}

func packData(p Packer) ([]byte, error) {
	tmp_p := p
	tmp_p.F1 += 1
	tmp_p.F2 += 2.22222
	// Создаем буфер для хранения байт
	buf := new(bytes.Buffer)

	// Пишем данные структуры в буфер
	err := binary.Write(buf, binary.BigEndian, tmp_p)
	if err != nil {
		return nil, err
	}

	// Возвращаем срез байт
	return buf.Bytes(), nil
}

// Функция для получения и распаковки пакета через TCP
func receivePack(conn net.Conn) (Packer, error) {
	data := make([]byte, 16) // float64 занимает 8 байт, 2 float64 = 16 байт
	_, err := conn.Read(data)
	if err != nil {
		if err == io.EOF {
			fmt.Println("Client closed connection")
		}
		return Packer{}, err
	}
	return unpackFromBytes(data)
}

// Обработка каждого подключения в отдельной горутине
func handleConnection(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	// Устанавливаем таймаут на чтение и запись
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	for {
		p, err := receivePack(conn)
		if err != nil {
			// Если клиент закрыл соединение
			if err == io.EOF {
				break
			}
			fmt.Println("Error receiving Packer:", err)
			break
		}

		// Выводим полученные данные от клиента
		fmt.Printf("Received Packer from client: %+v\n", p)

		// Пакуем данные и отправляем обратно
		buf, err := packData(p)
		if err != nil {
			fmt.Println("Error packing data:", err)
			break
		}

		_, err = conn.Write(buf)
		if err != nil {
			fmt.Println("Error writing response:", err)
			break
		}

		// Выводим данные, которые были отправлены обратно клиенту
		fmt.Printf("Sent back to client: %+v\n", p)

		// Сбрасываем таймауты после каждой успешной операции
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	}

	// Сообщаем о закрытии соединения только один раз
	fmt.Println("Closing connection with client...")
	conn.Close()
}

func main() {
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
			handleConnection(conn, &wg)
		}(conn)
	}

	// Ожидаем завершения всех горутин
	wg.Wait()
}
