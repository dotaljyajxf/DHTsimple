package dht

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/marksamman/bencode"
)

const (
	perBlock        = 16384
	maxMetadataSize = perBlock * 1024
	extended        = 20
	extHandshake    = 0
)

type Meta struct {
	addr         string
	infoHash     []byte
	timeout      time.Duration
	conn         net.Conn
	peerId       string
	preHeader    []byte
	metadataSize int64
	utMetadata   int64
	pieceCount   int64
	pieces       [][]byte
}

func NewMeta(addr string, hash []byte) *Meta {
	return &Meta{
		addr:      addr,
		infoHash:  hash,
		timeout:   20 * time.Second,
		peerId:    RandString(20),
		preHeader: MakePreHeader(),
	}
}

func (mw *Meta) checkDone() bool {
	for _, b := range mw.pieces {
		if b == nil {
			return false
		}
	}
	return true
}

func (m *Meta) readOnePiece(payload []byte) error {
	trailerIndex := bytes.Index(payload, []byte("ee")) + 2
	if trailerIndex == 1 {
		return errors.New("ee == 1")
	}

	dict, err := bencode.Decode(bytes.NewBuffer(payload[:trailerIndex]))
	if err != nil {
		return err
	}

	pieceIndex, ok := dict["piece"].(int64)
	if !ok || pieceIndex >= m.pieceCount {
		return errors.New("piece num error")
	}

	msgType, ok := dict["msg_type"].(int64)
	if !ok || msgType != 1 {
		return errors.New("piece type error")
	}
	m.pieces[pieceIndex] = payload[trailerIndex:]
	return nil
}

func (m *Meta) sendRequestPiece() {
	for i := 0; i < int(m.pieceCount); i++ {
		m.requestPiece(i)
	}
}

func (m *Meta) Begin() ([]byte, error) {
	m.SetDeadLine(300 * 2)

	for {
		data, err := m.ReadN()
		if err != nil {
			return nil, err
		}

		if len(data) < 2 {
			continue
		}

		if data[0] != extended {
			continue
		}

		if data[1] == 0 {
			err := m.onExtHandshake(data[2:])
			if err != nil {
				fmt.Println("ext Hand err: ", err.Error())
			}
			continue
		}
		err = m.readOnePiece(data[2:])
		if err != nil {
			return nil, err
		}

		if !m.checkDone() {
			continue
		}

		pie := bytes.Join(m.pieces, []byte(""))
		sum := sha1.Sum(pie)
		if bytes.Equal(sum[:], m.infoHash) {
			return pie, nil
		}

		return nil, errors.New("metadata checksum mismatch")
	}
}

func (m *Meta) Start() []byte {
	err := m.Connect()
	if err != nil {
		fmt.Printf("connect err:%s\n", err.Error())
		return nil
	}
	defer m.conn.Close()
	fmt.Println("connect finish")
	ret, err := m.Begin()
	if err != nil {
		fmt.Printf("read  err:%s\n", err.Error())
		return nil
	}
	return ret
}

func (m *Meta) SetDeadLine(second time.Duration) {
	deadLine := time.Now().Add(second * time.Second)
	m.conn.SetReadDeadline(deadLine)
	m.conn.SetWriteDeadline(deadLine)
}

func (m *Meta) Connect() error {
	var err error
	m.conn, err = net.DialTimeout("tcp", m.addr, m.timeout)
	if err != nil {
		return err
	}
	m.SetDeadLine(300)
	err = m.HandShake()
	if err != nil {
		return err
	}
	return m.extHandShake()
}

func (m *Meta) WriteTo(data []byte) error {

	length := uint32(len(data))

	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.BigEndian, length)

	sendMsg := append(buf.Bytes(), data...)
	_, err := m.conn.Write(sendMsg)
	if err != nil {
		return fmt.Errorf("write message failed: %v", err)
	}
	return nil
}

func (m *Meta) ReadN() ([]byte, error) {
	length := make([]byte, 4)
	_, err := io.ReadFull(m.conn, length)
	if err != nil {
		return nil, err
	}

	size := binary.BigEndian.Uint32(length)

	data := make([]byte, size)
	_, err = io.ReadFull(m.conn, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (m *Meta) extHandShake() error {
	//etxHandShark
	data := append([]byte{extended, extHandshake}, bencode.Encode(map[string]interface{}{
		"m": map[string]interface{}{
			"ut_metadata": 1,
		},
	})...)

	return m.WriteTo(data)
}

func (m *Meta) HandShake() error {
	buf := bytes.NewBuffer(nil)
	buf.Write(m.preHeader)
	buf.Write(m.infoHash)
	buf.WriteString(m.peerId)
	_, err := m.conn.Write(buf.Bytes())

	res := make([]byte, 68)
	n, err := io.ReadFull(m.conn, res)
	if err != nil {
		return err
	}
	if n != 68 {
		return errors.New("hand read len err")
	}

	if !bytes.Equal(res[:20], m.preHeader[:20]) {
		return errors.New("remote peer not supporting bittorrent protocol")
	}

	if res[25]&0x10 != 0x10 {
		return errors.New("remote peer not supporting extension protocol")
	}

	if !bytes.Equal(res[28:48], m.infoHash) {
		return errors.New("invalid bittorrent header response")
	}
	return nil
}

func (this *Meta) onExtHandshake(payload []byte) error {

	dict, err := bencode.Decode(bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	metadataSize, ok := dict["metadata_size"].(int64)
	if !ok {
		return errors.New("invalid extension header response")
	}

	if metadataSize > maxMetadataSize {
		return errors.New("metadata_size too long")
	}

	if metadataSize < 0 {
		return errors.New("negative metadata_size")
	}

	m, ok := dict["m"].(map[string]interface{})
	if !ok {
		return errors.New("negative metadata m")
	}

	utMetadata, ok := m["ut_metadata"].(int64)
	if !ok {
		return errors.New("negative metadata ut_metadata")
	}
	this.metadataSize = metadataSize
	this.utMetadata = utMetadata
	this.pieceCount = metadataSize / perBlock
	if this.metadataSize%perBlock != 0 {
		this.pieceCount++
	}
	this.pieces = make([][]byte, this.pieceCount)
	this.sendRequestPiece()
	return nil
}

func (mw *Meta) requestPiece(i int) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(extended)
	buf.WriteByte(byte(mw.utMetadata))
	//buf.WriteByte(2)
	buf.Write(bencode.Encode(map[string]interface{}{
		"msg_type": 0,
		"piece":    i,
	}))
	//fmt.Println(buf.String())
	err := mw.WriteTo(buf.Bytes())
	if err != nil {
		fmt.Println("write err :", err.Error())
	}
}
