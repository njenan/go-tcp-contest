package main

import "net"

func main() {
	if conn, err := net.Dial("tcp", ":3280"); err == nil {
		defer conn.Close()
		if _, err := conn.Write([]byte("shutdown\n")); err == nil {
		} else {
			panic(err)
		}
	} else {
		panic(err)
	}
}
