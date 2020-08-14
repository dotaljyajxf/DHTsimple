package load

import (
	"DHTsimple/config"
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/marksamman/bencode"
)

type Tfile struct {
	Name   string `json:"file_name"`
	Length int64  `json:"file_len"`
}

type Torrent struct {
	HashHex string   `json:"hash"`
	Name    string   `json:"name"`
	Length  int64    `json:"len"`
	Files   []*Tfile `json:"files"`
}

type HashPair struct {
	Hash []byte
	Addr string
}

var HashChan chan HashPair

func init() {
	HashChan = make(chan HashPair, config.Conf.LoadBufLen)
}

func work() {
	for {
		select {
		case info := <-HashChan:
			d := NewMeta(info.Addr, info.Hash)
			metaData := d.Load()
			if metaData == nil {
				continue
			}
			t, err := parseTorrent(metaData, hex.EncodeToString(info.Hash))
			if err != nil {
				continue
			}
			InsertToEs(t)
		}
	}
}

func LoadTorrent(n int) {
	for i := 0; i < n; i++ {
		go work()
	}
}

func parseTorrent(meta []byte, hashHex string) (*Torrent, error) {
	dict, err := bencode.Decode(bytes.NewBuffer(meta))
	if err != nil {
		return nil, err
	}

	t := &Torrent{HashHex: hashHex}
	if name, ok := dict["name.utf-8"].(string); ok {
		t.Name = name
	} else if name, ok := dict["name"].(string); ok {
		t.Name = name
	}
	if length, ok := dict["length"].(int64); ok {
		t.Length = length
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
		t.Files = append(t.Files, &Tfile{Name: filename, Length: filelength})
	}

	if files, ok := dict["files"].([]interface{}); ok {
		for _, file := range files {
			if f, ok := file.(map[string]interface{}); ok {
				extractFiles(f)
			}
		}
	}

	if t.Length == 0 {
		t.Length = totalSize
	}
	if len(t.Files) == 0 {
		t.Files = append(t.Files, &Tfile{Name: t.Name, Length: t.Length})
	}

	return t, nil
}
