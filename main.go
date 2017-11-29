package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var Values = make(map[string][]string)
var count = 0
var logFileNumber = 0

var lock sync.Mutex

func handleConnection(logFile *os.File, connection net.Conn) {
	defer connection.Close()

	data := make([]byte, 0, 4096)
	temp := make([]byte, 256)

	for {
		length, err := connection.Read(temp)

		data = append(data, temp[:length]...)
		if err != nil {
			break
		}

		split := strings.Split(string(data), "\n")
		for _, value := range split {
			lock.Lock()
			length = len(Values[value])
			lock.Unlock()
			if length == 0 {
				if value == "" {
					continue
				}

				count++

				_, err := logFile.WriteString(value + "\n")

				if err != nil {
					panic(err)
				}
			}

			lock.Lock()
			Values[value] = append(Values[value], value)
			lock.Unlock()
		}
	}
}

func printLoop() {
	uniqueEntries := 0

	for {
		time.Sleep(5 * time.Second)
		lock.Lock()
		temp := len(Values)
		lock.Unlock()
		periodUniqueEntries := temp - uniqueEntries
		uniqueEntries = temp

		fmt.Printf("**** START REPORT ****\n")
		fmt.Printf("Total unique entries: %v\n", uniqueEntries)
		fmt.Printf("Period unique entries: %v\n", periodUniqueEntries)
		fmt.Printf("**** END REPORT ****\n\n")

		periodUniqueEntries = 0
	}
}

func logIncrementer() {
	for {
		time.Sleep(10 * time.Second)
		logFileNumber++
		os.Create("data." + strconv.Itoa(logFileNumber) + ".log")

		var err error
		logFile, err = os.OpenFile("data."+strconv.Itoa(logFileNumber)+".log", os.O_APPEND|os.O_WRONLY, 0600)

		if err != nil {
			panic(err)
		}
	}
}

var startTime time.Time
var logFile *os.File

func main() {
	startTime = time.Now()
	server, err := net.Listen("tcp", ":3280")

	if err != nil {
		panic(err)
	}

	go printLoop()
	go logIncrementer()

	os.Create("data." + strconv.Itoa(count) + ".log")

	logFile, err = os.OpenFile("data."+strconv.Itoa(count)+".log", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}

	defer logFile.Close()

	for {
		connection, err := server.Accept()
		if err != nil {
			panic(err)
		}

		go handleConnection(logFile, connection)
	}
}
