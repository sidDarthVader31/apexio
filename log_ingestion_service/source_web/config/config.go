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
  MESSAGE_BROKER_CLIENTID string
  MESSSAGE_BROKER_ACKS string
  MESSAGE_BROKER_RETRIES int
  MESSAGE_BROKER_MAX_RETRIES int
  MESSAGE_BROKER_TIMEOUT int
  MESSAGE_BROKER_LINGER_MS int
  MESSAGE_BROKER_LOG_LEVEL int
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
  Config.MESSAGE_BROKER_CLIENTID = viper.GetString("MESSAGE_BROKER_CLIENTID")
  Config.MESSSAGE_BROKER_ACKS = viper.GetString("MESSSAGE_BROKER_ACKS")
  Config.MESSAGE_BROKER_RETRIES = viper.GetInt("MESSAGE_BROKER_RETRIES")
  Config.MESSAGE_BROKER_MAX_RETRIES = viper.GetInt("MESSAGE_BROKER_MAX_RETRIES")
  Config.MESSAGE_BROKER_TIMEOUT = viper.GetInt("MESSAGE_BROKER_TIMEOUT")
  Config.MESSAGE_BROKER_LINGER_MS = viper.GetInt("MESSAGE_BROKER_LINGER_MS")
  Config.MESSAGE_BROKER_LOG_LEVEL = viper.GetInt("MESSAGE_BROKER_LOG_LEVEL")
}
