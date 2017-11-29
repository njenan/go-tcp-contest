package main

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var Values = make(map[string][]string)
var count = 0
var logFileNumber = 0

var mapLock sync.Mutex
var fileLock sync.Mutex

var cont = true

type TCPReader struct {
	data   []rune
	cursor int
}

type TCPReadNext interface {
	Next() ([]rune, error)
}

func (reader *TCPReader) Next() ([]rune, error) {
	startingCursor := reader.cursor
	next := reader.data[reader.cursor]

	for next != '\n' {
		reader.cursor++

		if reader.cursor >= len(reader.data) {
			return nil, errors.New("Malformed data")
		}
	}

	return reader.data[startingCursor:reader.cursor], nil
}

func NewTCPReader(data []byte) *TCPReader {
	reader := new(TCPReader)
	reader.cursor = 0
	reader.data = bytes.Runes(data)
	return reader
}

func handleConnection(logFile *os.File, connection net.Conn) {
	defer connection.Close()

	data := make([]byte, 0, 4096)
	temp := make([]byte, 256)

	for cont {
		length, err := connection.Read(temp)

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
