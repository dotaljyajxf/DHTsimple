package main

import (
	"dhtTest/dht"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	d := dht.NewDHT("0.0.0.0:12121")
	err := d.Start()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-s
	fmt.Println("over")
}
