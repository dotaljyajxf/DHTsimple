package load

import (
	"DHTsimple/config"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/olivere/elastic/v7"
)

var esClient *elastic.Client

const INDEX = "torrent"

func init() {
	var err error
	esClient, err = elastic.NewClient(
		elastic.SetURL(config.Conf.ElasticUrl),
		elastic.SetBasicAuth(config.Conf.ElasticName, config.Conf.ElasticPwd),
		//elastic.SetErrorLog(os.Stdout),
		elastic.SetSniff(false),
	)
	if err != nil {
		fmt.Println("open elastic err:", err.Error())
		return
	}
}

func InsertToEs(t *Torrent) {
	codeT, err := json.Marshal(t)
	if err != nil {
		fmt.Println("encode json err ", err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	index := fmt.Sprintf("torrent-%s", time.Now().Add(-24*time.Hour).Format("20060102"))

	_, err = esClient.Index().Index(index).Id(t.HashHex).BodyString(string(codeT)).Do(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func GetHashInfo(name string) {
	//短语搜索 搜索about字段中有 name
	matchQuery := elastic.NewMatchQuery("name", name)

	nestedQuery := elastic.NewMatchQuery("files.file_name", name)
	nestedQ := elastic.NewNestedQuery("files", nestedQuery)

	searchQuery := elastic.NewBoolQuery().Should(matchQuery, nestedQ)
	res, err := esClient.Search(INDEX).Query(searchQuery).Do(context.Background())
	if err != nil {
		fmt.Println("search err: ", err.Error())
		return
	}
	var typ Torrent
	for _, item := range res.Each(reflect.TypeOf(typ)) { //从搜索结果中取数据的方法
		t := item.(Torrent)
		fmt.Printf("%#v\n", t)
	}
}
