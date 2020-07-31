package dht

import (
	"bytes"
	"container/list"
	"fmt"

	"net"

	"github.com/marksamman/bencode"
	//"github.com/jackpal/bencode-go"
)

var seed = []string{
	"router.utorrent.com:6881",
	"router.bittorrent.com:6881",
	"dht.transmissionbt.com:6881",
}

type DHT struct {
	Host     string
	NodeList *list.List
	Conn     *net.UDPConn
	Id       string
}

type handleFunc func(d *DHT)

func NewDHT(host string) *DHT {
	return &DHT{
		Host:     host,
		NodeList: list.New(),
		Id:       RandString(20),
	}
}

func (d *DHT) Start() error {
	addr, err := net.ResolveUDPAddr("udp", d.Host)
	if err != nil {
		return err
	}
	d.Conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	d.rung("readRequest", d.readRequest)
	d.rung("sendResponse", d.sendResponse)
	d.initSend()
	return nil
}

func (d *DHT) initSend() {
	fmt.Println(d.Id)
	for _, addr := range seed {
		udpAddr, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			fmt.Printf("resoveSeed error : %s\n", err.Error())
			continue
		}
		req := make(map[string]interface{})
		req["t"] = RandString(2)
		req["y"] = "q"
		req["q"] = "find_node"
		req["a"] = map[string]interface{}{"id": d.Id, "target": RandString(20)}

		_, err = d.Conn.WriteToUDP(bencode.Encode(req), udpAddr)
		if err != nil {
			fmt.Printf("send seed err:%s", err.Error())
		}

	}
}

func (d *DHT) rung(name string, localFunc func()) {
	f := func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("name:%s get a error:%s\n", name, err)
			}
		}()
		localFunc()
	}

	go f()
}

func (d *DHT) sendResponse() {

}

func (d *DHT) readRequest() {
	byteBuf := make([]byte, 526)
	readBuf := bytes.NewBuffer(byteBuf)
	for {
		fmt.Println("Begin_Read")
		n, addr, err := d.Conn.ReadFromUDP(readBuf.Bytes())
		if err != nil {
			fmt.Println("read err:%s", err.Error())
			continue
		}
		fmt.Println(addr.String())
		fmt.Println(n)
		fmt.Println(readBuf.String())
		fmt.Println("finish read")
	}

}
