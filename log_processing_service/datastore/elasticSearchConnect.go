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
    Username: "elastic",
    Password: "v9e95jhUulIPWCKMsmBn",
  }
  Es, err = elasticsearch.NewClient(cfg);
  if err != nil {
    fmt.Println("connection to elastic search failed:", err)
  }else{
    fmt.Println("successfully connected to elastic search :", err)
  }
  fmt.Println("elastic client:", &Es)
  return Es, err
}
