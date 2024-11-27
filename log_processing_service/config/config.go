package config
import (
	"fmt"
	"github.com/spf13/viper"
)

type IConfig struct{
  KAFKA_HOST string
  KAFKA_USER string
  KAFKA_PASSWORD string
  PORT string
  MESSAGE_BROKER string
  ELASTIC_HOST string
  ELASTIC_PORT string
  ELASTIC_PASSWORD string
  ELASTIC_USER string
}

var Config IConfig

func InitEnv() error{
  viper.SetConfigFile(".env")
  err := viper.ReadInConfig()
  if err != nil{
    fmt.Println("Error while reading env file:", err)
    return err
  }
  initEnvVariables()
  return nil
}

func initEnvVariables(){
  Config.KAFKA_HOST = viper.GetString("KAFKA_HOST")
  Config.MESSAGE_BROKER = viper.GetString("MESSAGE_BROKER")
  Config.KAFKA_USER = viper.GetString("KAFKA_USER")
  Config.KAFKA_PASSWORD = viper.GetString("KAFKA_PASSWORD")
  Config.PORT = viper.GetString("PORT")
  Config.ELASTIC_HOST = viper.GetString("ELASTIC_HOST")
  Config.ELASTIC_PORT = viper.GetString("ELASTIC_PORT")
  Config.ELASTIC_USER = viper.GetString("ELASTIC_USER")
  Config.ELASTIC_PASSWORD = viper.GetString("ELASTIC_PASSWORD")
}
