package main

import (
	"DHTsimple/dht"
	"DHTsimple/load"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	d := dht.NewDHT()
	err := d.Start()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	go load.LoadTorrent(2)

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-s
	fmt.Println("over")
}
