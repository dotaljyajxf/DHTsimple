package config

import (
	"fmt"
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Host               string `yaml:"host"`
	PerSecondSendLimit int    `yaml:"per_second_send_limit"`
	RequestBufLen      int    `yaml:"request_buf_len"`
	ResponseBufLen     int    `yaml:"response_buf_len"`
	DataBufLen         int    `yaml:"data_buf_len"`
	ConnectTimeout     int    `yaml:"connect_timeout"`
	HandTimeout        int    `yaml:"hand_timeout"`
	ReadTimeout        int    `yaml:"read_timeout"`
	WriteTimeout       int    `yaml:"write_timeout"`
	LoadBufLen         int    `yaml:"load_buf_len"`
	MongoUri           string `yaml:"mongo_uri"`
}

var Conf Config

func init() {
	dir, _ := os.Getwd()
	fp, err := os.OpenFile(path.Join(dir, "config.yaml"), os.O_RDONLY, 0664)
	if err != nil {
		fmt.Println("open config file err:", err.Error())
		return
	}
	d := yaml.NewDecoder(fp)

	err = d.Decode(&Conf)
	if err != nil {
		fmt.Println("decode config err:", err.Error())
	}
}
