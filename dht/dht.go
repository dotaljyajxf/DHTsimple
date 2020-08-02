package dht

import (
	"bytes"
	"container/list"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/marksamman/bencode"
	"golang.org/x/time/rate"
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
	Limiter      *rate.Limiter
}

func NewDHT(host string, limit int) *DHT {
	return &DHT{
		Host:         host,
		NodeList:     list.New(),
		Id:           RandString(20),
		RequestList:  make(chan *FindNodeReq, 2048),
		ResponseList: make(chan *Response, 2048),
		DataList:     make(chan map[string]interface{}, 2048),
		Limiter:      rate.NewLimiter(rate.Every(time.Second/time.Duration(limit)), limit),
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
		d.Limiter.Wait(context.Background())
		select {
		case req := <-d.RequestList:

			udpAddr, err := net.ResolveUDPAddr("udp", req.Addr)
			if err != nil {
				fmt.Printf("resoveAddr error : %s\n", err.Error())
				continue
			}
			//fmt.Println(len(bencode.Encode(req.Req)))
			//req := MakeRequest("find_node", d.Id, RandString(20))
			//fmt.Println("send  find_node to ", udpAddr.String())
			_, err = d.Conn.WriteToUDP(bencode.Encode(req.Req), udpAddr)
			if err != nil {
				fmt.Printf("send seed err:%s", err.Error())
			}
		}
	}
}

func (d *DHT) sendResponse() {
	for {
		d.Limiter.Wait(context.Background())
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
	readBuf := make([]byte, 2048)
	for {
		//fmt.Println("Begin_Read")
		//readBuf := NewBufferByte()
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
				y, ok := data["y"].(string)
				if !ok {
					//fmt.Printf("msg y is not string: %v\n", data)
					continue
				}
				t, ok := data["t"].(string)
				if !ok {
					fmt.Printf("msg t is not string\n")
					continue
				}
				remoteAddr, _ := data["remote_addr"].(*net.UDPAddr)

				if y == "q" {
					q, ok := data["q"].(string)
					if !ok {
						fmt.Printf("msg q is not string\n")
						continue
					}
					switch q {
					case "ping":
						d.doPing(remoteAddr, t)
					case "find_node":
						d.doFindNode(remoteAddr, t)
					case "get_peers":
						a, ok := data["a"].(map[string]interface{})
						if !ok {
							fmt.Printf("get peer no arg\n")
							break
						}
						d.doGetPeer(a, remoteAddr, t)
					case "announce_peer":
						a, ok := data["a"].(map[string]interface{})
						if !ok {
							fmt.Printf("announce_peer no arg\n")
							break
						}
						d.doAnnouncePeer(remoteAddr, t, a)
					}
				} else if y == "r" {
					r, ok := data["r"].(map[string]interface{})
					if !ok {
						break
					}
					d.decodeNodes(r)
				} else if y == "e" {
					//e, _ := data["e"]
					//fmt.Printf("msg get a err :%v\n", e)
				} else {
					fmt.Printf("Unknow msg :%v\n", data)
				}
			}
		}
	}

}

func (d *DHT) doPing(addr *net.UDPAddr, t string) {
	resp := new(Response)
	resp.R = map[string]interface{}{"id": d.Id}
	resp.T = t
	resp.Addr = addr
	d.ResponseList <- resp

}

func (d *DHT) doFindNode(addr *net.UDPAddr, t string) {
	r := make(map[string]interface{})
	r["nodes"] = ""
	r["id"] = d.Id
	resp := &Response{Addr: addr, T: t, R: r}
	d.ResponseList <- resp
}

func (d *DHT) doGetPeer(arg map[string]interface{}, addr *net.UDPAddr, t string) {

	id, ok := arg["id"].(string)
	if !ok {
		fmt.Println("doGetPeer no id")
		return
	}

	infoHash, ok := arg["info_hash"].(string)
	if !ok {
		fmt.Println("doGetPeer no info_hash")
		return
	}

	//InsertHash(infoHash, addr.String(), id)

	r := make(map[string]interface{})
	r["nodes"] = ""
	r["token"] = MakeToken(addr.String())
	r["id"] = neighborId(infoHash, d.Id)
	resp := &Response{Addr: addr, T: t, R: r}

	d.ResponseList <- resp
}

func (d *DHT) doAnnouncePeer(addr *net.UDPAddr, t string, arg map[string]interface{}) {
	//token, ok := arg["token"].(string)
	//if !ok {
	//	fmt.Println("doAnnouncePeer no token")
	//	return
	//}

	//if !ValidateToken(token, addr.String()) {
	//	fmt.Println("doAnnouncePeer token un match")
	//	return
	//}
	id, ok := arg["id"].(string)
	if !ok {
		fmt.Println("doGetPeer no id")
		return
	}

	infoHash, ok := arg["info_hash"].(string)
	if !ok {
		fmt.Println("doAnnouncePeer no info_hash")
		return
	}

	port := int64(addr.Port)
	if impliedPort, ok := arg["implied_port"].(int64); ok && impliedPort == 0 {
		if p, ok := arg["port"].(int64); ok {
			port = p
		}
	}

	peer := &net.TCPAddr{IP: addr.IP, Port: int(port)}

	InsertHash(infoHash, peer.String(), id)

	r := make(map[string]interface{})
	r["id"] = d.Id
	resp := &Response{Addr: addr, T: t, R: r}

	d.ResponseList <- resp
}

//nodes 0-19为id,20-23为ip,24-25为端口
func (d *DHT) decodeNodes(r map[string]interface{}) {
	//fmt.Println("decodeNodes")
	nodes, ok := r["nodes"].(string)
	if !ok {
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
		if port <= 0 || port > 65535 {
			continue
		}
		addr := ip + ":" + strconv.Itoa(int(port))
		r := MakeRequest("find_node", d.Id, id)
		req := &FindNodeReq{Addr: addr, Req: r}
		d.RequestList <- req
	}

	return
}
