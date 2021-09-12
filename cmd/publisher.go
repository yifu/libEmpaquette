package main

import (
	"fmt"
	"log"
	"net"
	"bufio"
	"github.com/yifu/libEmpaquette"
)

type myTCPConn struct {
	*net.TCPConn
}

func (conn myTCPConn) WriteByte(b byte) error {
	buf := []byte {b}
	_, err := conn.Write(buf)
	return err
}

func main() {
	fmt.Println("Hello world2.")
	
	conn, err := net.Dial("tcp", "192.168.0.15:1883")
	if err != nil {
		log.Fatal(err)
	}

	w := bufio.NewWriter(conn)

	err = libEmpaquette.CreateConnect(w)
	if err != nil {
		log.Fatal(err)
	}

	w.Flush()

	buf := make([]byte, 100)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("n = ", n)

	fmt.Println("End of publisher")
}