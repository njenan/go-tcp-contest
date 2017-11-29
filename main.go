package main

import (
	"bytes"
	"errors"
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

var mapLock sync.RWMutex
var fileLock sync.Mutex

var cont = true

var malformedError = errors.New("Malformed Data")

type tcpReader struct {
	data   []rune
	cursor int
}

type tcpReadNext interface {
	Next() (string, error)
}

func (reader *tcpReader) Next() (string, error) {
	if reader.cursor >= len(reader.data) {
		return "", nil
	}

	startingCursor := reader.cursor
	next := reader.data[reader.cursor]

	for next != '\n' {
		reader.cursor++

		if reader.cursor >= len(reader.data) {
			return "", malformedError
		}

		next = reader.data[reader.cursor]
	}

	return string(reader.data[startingCursor:reader.cursor]), nil
}

func newTCPReader(data []byte) *tcpReader {
	reader := new(tcpReader)
	reader.cursor = 0
	reader.data = bytes.Runes(data)
	return reader
}

func handleConnection(logFile *os.File, connection net.Conn) {
	defer connection.Close()

	//data := make([]byte, 0, 4096)
	temp := make([]byte, 256)

	for cont {
		_, err := connection.Read(temp)

		if err != nil {
			panic(err)
		}

		reader := newTCPReader(temp)

		fmt.Println("Trying to get next entry")
		for {
			entry, err := reader.Next()

			if err != nil {
				fmt.Println(err)
				break
			}

			if entry == "" {
				break
			}

			fmt.Printf("Entry %v is %v\n", count, entry)

			mapLock.RLock()
			length := len(Values[entry])

			if length == 0 {
				count++

				fileLock.Lock()
				_, err := logFile.WriteString(entry + "\n")
				fileLock.Unlock()

				if err != nil {
					panic(err)
				}
			}

			mapLock.Lock()
			Values[entry] = append(Values[entry], entry)
			mapLock.Unlock()
			mapLock.RUnlock()
		}

		/*

			data = append(data, temp[:length]...)
			if err != nil {
				break
			}

			split := strings.Split(string(data), "\n")
			for _, value := range split {
				mapLock.Lock()
				length = len(Values[value])
				if length == 0 {
					if value == "" {
						mapLock.Unlock()
						continue
					}

					if value == "shutdown" {
						fmt.Println("**** SHUTDOWN ****")
						cont = false
					}

					count++

					fileLock.Lock()
					_, err := logFile.WriteString(value + "\n")
					fileLock.Unlock()

					if err != nil {
						panic(err)
					}
				}

				Values[value] = append(Values[value], value)
				mapLock.Unlock()
			}

		*/
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

	defer logFile.Close()

	for cont {
		connection, err := server.Accept()
		if err != nil {
			panic(err)
		}

		go handleConnection(logFile, connection)
	}
}
