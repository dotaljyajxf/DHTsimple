package dht

import (
	"bytes"
	"container/list"
	"fmt"
	"net"

	"github.com/marksamman/bencode"
)

var seed = []string{
	"router.utorrent.com:6881",
	"router.bittorrent.com:6881",
	"dht.transmissionbt.com:6881",
}

type DHT struct {
	Host        string
	NodeList    *list.List
	Conn        *net.UDPConn
	Id          string
	RequestList chan string
	DataList    chan map[string]interface{}
}

type handleFunc func(d *DHT)

func NewDHT(host string) *DHT {
	return &DHT{
		Host:        host,
		NodeList:    list.New(),
		Id:          RandString(20),
		RequestList: make(chan string, 2048),
		DataList:    make(chan map[string]interface{}, 2048),
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

	d.rung("sendRequest", d.sendRequest)
	d.rung("handleData", d.handleData)
	d.rung("readResponse", d.readResponse)
	d.addSend()
	return nil
}

func (d *DHT) addSend() {
	for _, addr := range seed {
		d.RequestList <- addr
		//udpAddr, err := net.ResolveUDPAddr("udp", addr)
		//if err != nil {
		//	fmt.Printf("resoveSeed error : %s\n", err.Error())
		//	continue
		//}
		//req := make(map[string]interface{})
		//req["t"] = RandString(2)
		//req["y"] = "q"
		//req["q"] = "find_node"
		//req["a"] = map[string]interface{}{"id": d.Id, "target": RandString(20)}
		//
		//_, err = d.Conn.WriteToUDP(bencode.Encode(req), udpAddr)
		//if err != nil {
		//	fmt.Printf("send seed err:%s", err.Error())
		//}

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

func (d *DHT) sendRequest() {
	for {
		select {
		case oneAddr := <-d.RequestList:
			udpAddr, err := net.ResolveUDPAddr("udp", oneAddr)
			if err != nil {
				fmt.Printf("resoveAddr error : %s\n", err.Error())
				continue
			}
			req := MakeRequest("find_node", d.Id, RandString(20))

			_, err = d.Conn.WriteToUDP(bencode.Encode(req), udpAddr)
			if err != nil {
				fmt.Printf("send seed err:%s", err.Error())
			}
			if len(d.RequestList) == 0 {
				d.addSend()
			}
		}
	}
}

func (d *DHT) readResponse() {
	readBuf := NewBufferByte()
	for {
		//fmt.Println("Begin_Read")
		n, _, err := d.Conn.ReadFromUDP(readBuf)
		if err != nil {
			fmt.Printf("read err:%s", err.Error())
			continue
		}
		msg, err := bencode.Decode(bytes.NewBuffer(readBuf[:n]))
		if err != nil {
			fmt.Printf("decode buf error:%s\n", err.Error())
			continue
		}
		d.DataList <- msg
	}
}

func (d *DHT) handleData() {
	for {
		select {
		case data := <-d.DataList:
			{
				//msg, err := bencode.Decode(bytes.NewBuffer(data))
				//if err != nil {
				//	fmt.Printf("decode buf error:%s\n", err.Error())
				//	continue
				//}
				y, ok := data["y"].(string)
				if !ok {
					fmt.Printf("msg y is not string\n")
					continue
				}

				if y == "q" {
					fmt.Println(data)
				} else if y == "r" {
					fmt.Println(data)
				} else if y == "e" {
					e, _ := data["e"].(string)
					fmt.Printf("msg get a err :%s\n", e)
				}
			}
		}
	}

}
