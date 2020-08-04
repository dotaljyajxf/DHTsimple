package dht

import (
	"bytes"
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

func CreateTorrent(meta []byte, infohashHex string) error {
	dict, err := bencode.Decode(bytes.NewBuffer(meta))
	if err != nil {
		return err
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

	fmt.Printf("name:%s,HASH:%s,length:%d\n", t.name, t.infohashHex, t.length)
	for _, nfile := range t.files {
		fmt.Printf("%s-->%d\n", nfile.name, nfile.length)
	}
	return nil
}
