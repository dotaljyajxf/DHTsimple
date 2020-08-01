package main

import (
	"dhtTest/dht"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("need arg limit")
		return
	}
	limit, _ := strconv.Atoi(os.Args[1])
	d := dht.NewDHT("0.0.0.0:12121", limit)
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
