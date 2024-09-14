package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
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
	// Создаем буфер для хранения байт
	buf := new(bytes.Buffer)

	// Пишем данные структуры в буфер
	err := binary.Write(buf, binary.BigEndian, p)
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
		return Packer{}, err
	}
	return unpackFromBytes(data)
}

func main() {
	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server is listening...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		defer conn.Close()

		p, err := receivePack(conn)
		if err != nil {
			fmt.Println("Error receiving Packer:", err)
			continue
		}

		// Выводим пакет в консоль в hex формате
		fmt.Printf("Received Packer: %+v\n", p)
		buf, err := packData(p)
		if err == nil {
			_, err := conn.Write(buf)
			if err != nil {
				fmt.Println("Error receiving Packer:", err)
				continue
			}
		}

	}
}
