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

func main() {
	fmt.Println("Hello world2.")
	
	conn, err := net.Dial("tcp", "192.168.0.15:1883")
	if err != nil {
		log.Fatal(err)
	}

	w := bufio.NewWriter(conn)
	err = libEmpaquette.CreateConnect(w, "clienttes-toto")
	if err != nil {
		log.Fatal(err)
	}
	w.Flush()

	r := bufio.NewReader(conn)
	b, err := r.ReadByte()
	if err != nil {
		log.Fatal(err)
	}
	libEmpaquette.ProcessPkt(b, r)

	fmt.Println("End of publisher")
}
