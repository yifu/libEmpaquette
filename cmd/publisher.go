package main

import (
	"fmt"
	"log"
	"github.com/yifu/libEmpaquette"
)

func main() {
	fmt.Println("Hello world2.")
	
	ctx, err := libEmpaquette.Connect("192.168.0.15:1883")
	if err != nil {
		log.Fatal(err)
	}

	err = ctx.SendConnect("clienttesttoto")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("calling process pkt.")
	ctx.ProcessPkt()

	fmt.Println("End of publisher")
}
