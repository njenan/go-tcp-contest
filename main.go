package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

var Values = make(map[string][]string)
var count = 0
var logFileNumber = 0

var mapLock sync.Mutex
var fileLock sync.Mutex
var listenLock sync.Mutex

func handleConnection(connection net.Conn) {
	defer connection.Close()

	data := make([]byte, 0, 4096)
	temp := make([]byte, 256)

	cursor := 0

	for {
		length, err := connection.Read(temp)

		if err != nil {
			break
		}

		data = append(data, temp[:length]...)

	Loop:
		for {
			nextCursor := cursor + 1

			if nextCursor >= len(data) {
				break Loop
			}

			for rune(data[nextCursor]) != '\n' {
				nextCursor++

				if nextCursor >= len(data) {
					break Loop
				}
			}

			digits := string(data[cursor:nextCursor])

			if digits == "shutdown" {
				fmt.Println("**** SHUTDOWN RECEIVED ****")
				os.Exit(0)
			}

			if len(digits) != 10 {
				connection.Close()
				return
			}

			mapLock.Lock()
			instances := len(Values[digits])

			if instances == 0 {
				count++

				fileLock.Lock()
				_, err := logFile.WriteString(digits + "\n")
				fileLock.Unlock()

				if err != nil {
					panic(err)
				}
			}

			Values[digits] = append(Values[digits], digits)

			cursor = nextCursor + 1
			mapLock.Unlock()
		}
	}
}

func printLoop() {
	uniqueEntries := 0

	for {
		time.Sleep(5 * time.Second)
		mapLock.Lock()
		temp := len(Values)
		mapLock.Unlock()
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
		fileLock.Lock()
		logFileNumber++
		os.Create("data." + strconv.Itoa(logFileNumber) + ".log")

		var err error
		logFile, err = os.OpenFile("data."+strconv.Itoa(logFileNumber)+".log", os.O_APPEND|os.O_WRONLY, 0600)

		if err != nil {
			panic(err)
		}

		fileLock.Unlock()
	}
}

var startTime time.Time
var logFile *os.File

func waitForConnection(wg *sync.WaitGroup, server net.Listener) {
	for {
		listenLock.Lock()
		connection, err := server.Accept()
		listenLock.Unlock()

		if err != nil {
			panic(err)
		}

		handleConnection(connection)
	}
}

func main() {
	matches, err := filepath.Glob("data.*.log")

	for _, match := range matches {
		err = os.Remove(match)

		if err != nil {
			panic(err)
		}
	}

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

	var wg sync.WaitGroup

	for i := 0; i < 6; i++ {
		wg.Add(1)
		go waitForConnection(&wg, server)
	}

	wg.Wait()
}
