package datastore

import (
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
)



var Es *elasticsearch.Client
var err error
func ConnectToElastic() (*elasticsearch.Client, error){
  cfg := elasticsearch.Config{
    Addresses: []string{
      "http://localhost:9200",
    },
  }
  Es, err = elasticsearch.NewClient(cfg);
  if err != nil {
    fmt.Println("connection to elastic search failed:", err)
  }else{
    fmt.Println("successfully connected to elastic search ")
  }
  return Es, err
}
