package dht

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/marksamman/bencode"
)

var seed = []string{
	"router.utorrent.com:6881",
	"router.bittorrent.com:6881",
	"dht.transmissionbt.com:6881",
}

type FindNodeReq struct {
	Addr string
	Req  map[string]interface{}
}

type Response struct {
	Addr *net.UDPAddr
	R    map[string]interface{}
	T    string
}

type DHT struct {
	Host         string
	NodeList     *list.List
	Conn         *net.UDPConn
	Id           string
	RequestList  chan *FindNodeReq
	ResponseList chan *Response
	DataList     chan map[string]interface{}
}

func NewDHT(host string) *DHT {
	return &DHT{
		Host:         host,
		NodeList:     list.New(),
		Id:           RandString(20),
		RequestList:  make(chan *FindNodeReq, 2048),
		ResponseList: make(chan *Response, 2048),
		DataList:     make(chan map[string]interface{}, 2048),
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
	d.rung("sendResponse", d.sendResponse)
	d.rung("handleData", d.handleData)
	d.rung("readResponse", d.readResponse)
	d.addSend()
	go d.seedLoop()
	return nil
}

func (d *DHT) seedLoop() {
	timer := time.NewTicker(15 * time.Second)
	for {
		select {
		case <-timer.C:
			if len(d.RequestList) == 0 {
				d.addSend()
			}
		}
	}
}

func (d *DHT) addSend() {
	for _, addr := range seed {
		//d.RequestList <- addr
		//udpAddr, err := net.ResolveUDPAddr("udp", addr)
		//if err != nil {
		//	fmt.Printf("resoveSeed error : %s\n", err.Error())
		//	continue
		//}

		findNodeReq := new(FindNodeReq)
		req := make(map[string]interface{})
		req["t"] = RandString(2)
		req["y"] = "q"
		req["q"] = "find_node"
		req["a"] = map[string]interface{}{"id": d.Id, "target": RandString(20)}

		findNodeReq.Addr = addr
		findNodeReq.Req = req

		d.RequestList <- findNodeReq
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
		case req := <-d.RequestList:
			udpAddr, err := net.ResolveUDPAddr("udp", req.Addr)
			if err != nil {
				fmt.Printf("resoveAddr error : %s\n", err.Error())
				continue
			}
			//req := MakeRequest("find_node", d.Id, RandString(20))
			fmt.Println("send  find_node")
			_, err = d.Conn.WriteToUDP(bencode.Encode(req.Req), udpAddr)
			if err != nil {
				fmt.Printf("send seed err:%s", err.Error())
			}
		}
	}
}

func (d *DHT) sendResponse() {
	for {
		select {
		case resp := <-d.ResponseList:

			r := MakeResponse(resp.T, resp.R)
			_, err := d.Conn.WriteToUDP(bencode.Encode(r), resp.Addr)
			if err != nil {
				fmt.Printf("send seed err:%s", err.Error())
			}
		}
	}
}

func (d *DHT) readResponse() {
	readBuf := NewBufferByte()
	for {
		//fmt.Println("Begin_Read")
		n, addr, err := d.Conn.ReadFromUDP(readBuf)
		if err != nil {
			fmt.Printf("read err:%s", err.Error())
			continue
		}
		msg, err := bencode.Decode(bytes.NewBuffer(readBuf[:n]))
		if err != nil {
			fmt.Printf("decode buf error:%s\n", err.Error())
			continue
		}
		msg["remote_addr"] = addr
		d.DataList <- msg
	}
}

func (d *DHT) handleData() {
	for {
		select {
		case data := <-d.DataList:
			{
				fmt.Println("read  data")
				y, ok := data["y"].(string)
				if !ok {
					fmt.Printf("msg y is not string\n")
					continue
				}
				t, ok := data["t"].(string)
				if !ok {
					fmt.Printf("msg t is not string\n")
					continue
				}
				remoteAddr, _ := data["remote_addr"].(*net.UDPAddr)

				if y == "q" {
					q := data["y"].(string)
					switch q {
					case "ping":
						d.doPing(remoteAddr, t)
					case "find_node":
						d.doFindNode(remoteAddr, t)
					case "get_peers":
						rId, _ := data["id"].(string)
						d.doGetPeer(rId, remoteAddr, t)
					case "announce_peer":
						a, _ := data["a"].(map[string]interface{})
						d.doAnnouncePeer(remoteAddr, t, a)
					}
				} else if y == "r" {
					r, ok := data["r"].(map[string]interface{})
					if !ok {
						break
					}
					d.decodeNodes(r)
				} else if y == "e" {
					e, _ := data["e"].(string)
					fmt.Printf("msg get a err :%s\n", e)
				}
			}
		}
	}

}

func (d *DHT) doPing(addr *net.UDPAddr, t string) {
	fmt.Println("doPing")
	resp := new(Response)
	resp.R = map[string]interface{}{"id": d.Id}
	resp.T = t
	resp.Addr = addr
	d.ResponseList <- resp

}

func (d *DHT) doFindNode(addr *net.UDPAddr, t string) {
	fmt.Println("doFindNode")
	r := make(map[string]interface{})
	r["nodes"] = ""
	r["id"] = d.Id
	resp := &Response{Addr: addr, T: t, R: r}
	d.ResponseList <- resp
}

func (d *DHT) doGetPeer(id string, addr *net.UDPAddr, t string) {
	fmt.Println("doGetPeer")
	r := make(map[string]interface{})
	r["nodes"] = ""
	r["token"] = MakeToken(addr.String())
	r["id"] = neighborId(id, d.Id)
	resp := &Response{Addr: addr, T: t, R: r}

	d.ResponseList <- resp
}

func (d *DHT) doAnnouncePeer(addr *net.UDPAddr, t string, arg map[string]interface{}) {
	fmt.Println("doAnnouncePeer")
	token, ok := arg["token"].(string)
	if !ok {
		fmt.Println("doAnnouncePeer no token")
		return
	}

	if !ValidateToken(token, addr.String()) {
		fmt.Println("doAnnouncePeer token un match")
		return
	}

	infoHash, ok := arg["info_hash"].(string)
	if !ok {
		fmt.Println("doAnnouncePeer no info_hash")
		return
	}

	GetHash(infoHash)

	r := make(map[string]interface{})
	r["id"] = d.Id
	resp := &Response{Addr: addr, T: t, R: r}

	d.ResponseList <- resp
}

func (d DHT) decodeNodes(r map[string]interface{}) {
	fmt.Println("decodeNodes")
	nodes, ok := r["nodes"].(string)
	if !ok {
		fmt.Println("r not have nodes")
		return
	}

	length := len(nodes)
	if length%26 != 0 {
		fmt.Println("node can not mod 26")
		return
	}

	for i := 0; i < length; i += 26 {
		id := nodes[i : i+20]
		ip := net.IP(nodes[i+20 : i+24]).String()
		port := binary.BigEndian.Uint16([]byte(nodes[i+24 : i+26]))
		addr := ip + ":" + strconv.Itoa(int(port))
		r := MakeRequest("find_node", d.Id, id)
		req := &FindNodeReq{Addr: addr, Req: r}
		d.RequestList <- req
	}

	return
}
