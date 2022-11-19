package main

import (
	"bufio"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Сетевой адрес.
//
// Служба будет слушать запросы на всех IP-адресах
// компьютера на порту 12345.
// Например, 127.0.0.1:12345
const addr = "0.0.0.0:12345"

// Протокол сетевой службы.
const proto = "tcp4"

func main() {
	var wg sync.WaitGroup
	stopAppChan := make(chan os.Signal, 1)
	signal.Notify(stopAppChan, os.Interrupt)
	rand.Seed(time.Now().UnixNano())
	// Запуск сетевой службы по протоколу TCP
	// на порту 12345.

	// re for Go Proverbs
	reg := `(">)(.*)(<\/a><)`
	re := regexp.MustCompile(reg)
	listener, err := net.Listen(proto, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(listener)

	goProverbsSlice, err := getGoProverbs(re)

	if err != nil {
		log.Println(err)
	}

	wg.Add(1)

	go func() {
		for {
			// Принимаем подключение.
			conn, err := listener.Accept()
			if err != nil {
				log.Fatal(err)
			}

			select {
			default:
				// Вызов обработчика подключения.
				go handleConn(conn, goProverbsSlice, stopAppChan, &wg)
			}
		}
	}()
	wg.Wait()
}

// Обработчик. Вызывается для каждого соединения.
func handleConn(conn net.Conn, goProverbsSlice []string,
	stopAppChan chan os.Signal, wg *sync.WaitGroup) {
	// Закрытие соединения.
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(conn)
	ticker := time.NewTicker(3 * time.Second)
	// Чтение сообщения от клиента.
	reader := bufio.NewReader(conn)
	b, err := reader.ReadBytes('\n')
	if err != nil {
		log.Println(err)
		return
	}

	// Удаление символов конца строки.
	msg := strings.TrimSuffix(string(b), "\n")
	msg = strings.TrimSuffix(msg, "\r")
	// Если получили "go" - пишем поговорку в соединение.
	if msg == "go" {

		for {
			select {
			case <-stopAppChan:
				wg.Done()
			case <-ticker.C:
				// Вызов обработчика подключения.
				randomProverbIndex := getRandGoProverbIndex(len(goProverbsSlice))
				_, err = conn.Write([]byte(goProverbsSlice[randomProverbIndex] + "\n"))
				if err != nil {
					log.Println(err)
					return
				}
			}
		}
	}
}

func getGoProverbs(re *regexp.Regexp) ([]string, error) {
	goProverbs := make([]string, 0)
	resp, err := http.Get("https://go-proverbs.github.io/")
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(err)
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	goProverbsSubMatch := re.FindAllStringSubmatch(string(body), -1)

	for _, subMatchSlice := range goProverbsSubMatch {
		goProverbs = append(goProverbs, subMatchSlice[2])
	}

	return goProverbs, nil
}

func getRandGoProverbIndex(lastIndex int) int {
	return rand.Intn(lastIndex)
}
