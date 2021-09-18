package main

import (
	"fmt"
	"log"
	"net"
	"bufio"
	"github.com/yifu/libEmpaquette"
)

func main() {
	fmt.Println("Hello world2.")
	
	conn, err := net.Dial("tcp", "192.168.0.15:1883")
	if err != nil {
		log.Fatal(err)
	}

	w := bufio.NewWriter(conn)
	err = libEmpaquette.SendConnect(w, "clienttesttoto")
	if err != nil {
		log.Fatal(err)
	}
	w.Flush()

	r := bufio.NewReader(conn)
	libEmpaquette.ProcessPkt(r)

	fmt.Println("End of publisher")
}
