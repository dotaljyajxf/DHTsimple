package dht

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/marksamman/bencode"
)

type tfile struct {
	name   string
	length int64
}

type Torrent struct {
	infohashHex string
	name        string
	length      int64
	files       []*tfile
}

type HashPair struct {
	Hash []byte
	Addr string
}

var hashChan chan HashPair

func init() {
	hashChan = make(chan HashPair, 1024)
}

func work() {
	for {
		select {
		case info := <-hashChan:
			d := NewMeta(info.Addr, info.Hash)
			metaData := d.Start()
			if metaData == nil {
				continue
			}
			t, err := parseTorrent(metaData, hex.EncodeToString([]byte(info.Hash)))
			if err != nil {
				continue
			}
			fmt.Println("--------------------------------------------------------------------")
			fmt.Printf("name:%s,HASH:%s,length:%d\n", t.name, t.infohashHex, t.length)
			for _, nfile := range t.files {
				fmt.Printf("\t%s length:%d\n", nfile.name, nfile.length)
			}
			fmt.Println("--------------------------------------------------------------------")
		}
	}
}

func LoadTorrent(n int) {
	for i := 0; i < n; i++ {
		go work()
	}
}

func parseTorrent(meta []byte, infohashHex string) (*Torrent, error) {
	dict, err := bencode.Decode(bytes.NewBuffer(meta))
	if err != nil {
		return nil, err
	}

	t := &Torrent{infohashHex: infohashHex}
	if name, ok := dict["name.utf-8"].(string); ok {
		t.name = name
	} else if name, ok := dict["name"].(string); ok {
		t.name = name
	}
	if length, ok := dict["length"].(int64); ok {
		t.length = length
	}

	var totalSize int64
	var extractFiles = func(file map[string]interface{}) {
		var filename string
		var filelength int64
		if inter, ok := file["path.utf-8"].([]interface{}); ok {
			name := make([]string, len(inter))
			for i, v := range inter {
				name[i] = fmt.Sprint(v)
			}
			filename = strings.Join(name, "/")
		} else if inter, ok := file["path"].([]interface{}); ok {
			name := make([]string, len(inter))
			for i, v := range inter {
				name[i] = fmt.Sprint(v)
			}
			filename = strings.Join(name, "/")
		}
		if length, ok := file["length"].(int64); ok {
			filelength = length
			totalSize += filelength
		}
		t.files = append(t.files, &tfile{name: filename, length: filelength})
	}

	if files, ok := dict["files"].([]interface{}); ok {
		for _, file := range files {
			if f, ok := file.(map[string]interface{}); ok {
				extractFiles(f)
			}
		}
	}

	if t.length == 0 {
		t.length = totalSize
	}
	if len(t.files) == 0 {
		t.files = append(t.files, &tfile{name: t.name, length: t.length})
	}

	return t, nil
}
