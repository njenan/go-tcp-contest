package main

import (
	"io/ioutil"
	"net"
	"os"
	"testing"
	"time"
)

func TestSetUp(t *testing.T) {
	go main()

	if _, err := os.Create("data.0.log"); err != nil {
		panic(err)
	}
}

func reset() {
	Values = make(map[string][]string)

	if err := os.Truncate("data.0.log", 0); err != nil {
		panic(err)
	}
}

func testTCPConnection(t *testing.T, message string, expected string) {
	if conn, err := net.Dial("tcp", ":3280"); err == nil {
		if _, err := conn.Write([]byte(message)); err == nil {
			conn.Close()

			time.Sleep(100 * time.Millisecond)

			if data, err := ioutil.ReadFile("data.0.log"); err == nil {
				if expected != "" {
					message = expected
				}

				if message != string(data) {
					t.Errorf("log did not match what was sent, expected '%v' but was actually '%v'", message, string(data))
				}
			} else {
				panic(err)
			}
		} else {
			panic(err)
		}
	} else {
		panic(err)
	}
}

func TestUniqueEntriesAreLogged(t *testing.T) {
	reset()
	testTCPConnection(t, "0001000000\n0001000001\n", "")
}

func TestMoreUniqueEntriesAreLogged(t *testing.T) {
	reset()
	testTCPConnection(t, "0001000001\n0001000000\n", "")
}

func TestMultipleLoggingEvents(t *testing.T) {
	reset()
	testTCPConnection(t, "0001000001\n0001000000\n", "")
	testTCPConnection(t, "0001000000\n0001000001\n", "0001000001\n0001000000\n")
}

func TestUniqueEntriesAreCounted(t *testing.T) {
	reset()
	testTCPConnection(t, "0100000000\n0100000000\n0100000001", "0100000000\n0100000001\n")
}
