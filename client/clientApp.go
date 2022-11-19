package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
)

func main() {
	stopAppChan := make(chan os.Signal, 1)
	signal.Notify(stopAppChan, os.Interrupt)

	// Подключение к сетевой службе.
	conn, err := net.Dial("tcp4", "localhost:12345")
	if err != nil {
		log.Fatal(err)
	}
	// Не забываем закрыть ресурс.
	_, err = conn.Write([]byte("go\n"))
	if err != nil {
		log.Fatal(err)
	}

	// Буфер для чтения данных из соединения.
	reader := bufio.NewReader(conn)

	for {
		select {
		case <-stopAppChan:
			return
		default:
			// Считывание массива байт до перевода строки.
			b, err := reader.ReadBytes('\n')
			if err == io.EOF {
				continue
			}
			if err != nil {
				log.Fatal(err)
			}

			// Обработка ответа.
			fmt.Println("Ответ от сервера:", string(b))
		}
	}
}
