package datastore

import (
	"fmt"
	"log-processor/config"

	"github.com/elastic/go-elasticsearch/v8"
)
var Es *elasticsearch.Client
var err error
func ConnectToElastic() (*elasticsearch.Client, error){
  fmt.Println("elastic host:", config.Config.ELASTIC_HOST)
  cfg := elasticsearch.Config{
    Addresses: []string{
      config.Config.ELASTIC_HOST,
    },
    Username: config.Config.ELASTIC_USER,
    Password: config.Config.ELASTIC_PASSWORD,
  }
  Es, err = elasticsearch.NewClient(cfg);

  if err != nil {
    fmt.Println("connection to elastic search failed:", err)
  }else{
    fmt.Println("successfully connected to elastic search, now creating index :")
    ie :=InitIndex()
    if ie != nil {
      fmt.Println("error while creating index::", ie)
      return nil, ie
    }else{
      fmt.Println("index created successfully")
    }
  }
  return Es, err
}
