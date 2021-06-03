package main

import (
	"io"
	"log"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	log.Printf("Server connected.")
	buf := make([]byte, 256) // создаем буфер
	for {
		_, err = conn.Read(buf)
		if err == io.EOF {
			break
		}
		log.Printf("Read from server: %s", string(buf)) // выводим измененное сообщение сервера в консоль
	}
}
