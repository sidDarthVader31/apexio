package config
import (
	"fmt"
	"github.com/spf13/viper"
)

type IConfig struct{
  PORT string
  KAFKA_HOST string
  KAFKA_USER string
  KAFKA_PASSWORD string
  KAFKA_GROUP_ID string
  KAFKA_OFFSET_RESET string
  KAFKA_MAX_RETRIES int
  KAFKA_WORKERS int
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
  Config.KAFKA_GROUP_ID = viper.GetString("KAFKA_GROUP_ID")
  Config.KAFKA_OFFSET_RESET = viper.GetString("KAFKA_OFFSET_RESET")
  Config.KAFKA_WORKERS = viper.GetInt("KAFKA_WORKERS")
  Config.KAFKA_MAX_RETRIES = viper.GetInt("KAFKA_MAX_RETRIES")

  Config.PORT = viper.GetString("PORT")

  Config.ELASTIC_HOST = viper.GetString("ELASTIC_HOST")
  Config.ELASTIC_PORT = viper.GetString("ELASTIC_PORT")
  Config.ELASTIC_USER = viper.GetString("ELASTIC_USER")
  Config.ELASTIC_PASSWORD = viper.GetString("ELASTIC_PASSWORD")

}
