package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	POST  = "POST"
	GET   = "GET" 
)
type ElasticsearchDatasource struct {
    Name                            string `json:"name"`
    Type                            string `json:"type"`
    URL                             string `json:"url"`
    Access                          string `json:"access"`
    JsonData  struct {
        ESVersion                   int    `json:"esVersion"`
        TimeField                   string `json:"timeField"`
        Index                       string `json:"index"`
        LogMessageField             string `json:"logMessageField"`
        LogLevelField               string `json:"logLevelField"`
    } `json:"jsonData"`
}

type Datasource struct {
    ID            int    `json:"id"`
    UID           string `json:"uid"`
    OrgID         int    `json:"orgId"`
    Name          string `json:"name"`
    Type          string `json:"type"`
    TypeLogoURL   string `json:"typeLogoUrl"`
    Access        string `json:"access"`
    URL           string `json:"url"`
    User          string `json:"user"`
    Database      string `json:"database"`
    BasicAuth     bool   `json:"basicAuth"`
    BasicAuthUser string `json:"basicAuthUser"`
    WithCredentials bool   `json:"withCredentials"`
    IsDefault     bool   `json:"isDefault"`
    JsonData      struct {
        ESVersion    int    `json:"esVersion"`
        Index        string `json:"index"`
        LogLevelField string `json:"logLevelField"`
        LogMessageField string `json:"logMessageField"`
        TimeField     string `json:"timeField"`
    } `json:"jsonData"`
    SecureJsonFields map[string]interface{} `json:"secureJsonFields"`
    Version          int    `json:"version"`
    ReadOnly         bool   `json:"readOnly"`
    APIVersion       string `json:"apiVersion"`
}

type DashboardPayload struct {
    Dashboard json.RawMessage `json:"dashboard"`
    Overwrite bool            `json:"overwrite"`
}

type envVars struct{
  grafanaBaseUrl string
  apiToken string
}
func main(){
  env := getEnv()
  env.grafanaBaseUrl = "http://localhost:3000"

  //create a data source 
  esConfig := ElasticsearchDatasource{
        URL: "http://elasticsearch:9200",
        Name: "elastic_api",
        Access: "proxy",
        Type: "elasticsearch",
    }
  esConfig.JsonData.Index ="raw_logs"
  esConfig.JsonData.TimeField="timestamp"
  esConfig.JsonData.LogLevelField="logLevel"
  esConfig.JsonData.LogMessageField="message"

  payload, _ := json.Marshal(esConfig)
  fmt.Println("JSON payload:", string(payload[:]))
  responseDataSource, err := apicall(POST, "/api/datasources", env.grafanaBaseUrl, payload)
  if err != nil{
    fmt.Println("error creating data source , exiting:", err)
    os.Exit(1)
  }
  var datasource  Datasource
  json.Unmarshal(responseDataSource, &datasource)

  //create application dashboard

  //get json for dashboard 
  jsonDashboard := getDashboardJson()
  newJson := updateDataSourceInJSON(jsonDashboard, datasource.UID)
  fmt.Println("new json:", string(newJson[:]))

  dashboardPayload := DashboardPayload{
    Dashboard: newJson,
    Overwrite: true,
  }
  payloadBytes, _:= json.Marshal(dashboardPayload)
  dashboardresponse, dasherr:= apicall(POST, "/api/dashboards/db", env.grafanaBaseUrl, payloadBytes)

  if dasherr!=nil{
    fmt.Println("error while cratin dashboard:", dasherr)
  }else{
    fmt.Println("dashboard response:", string(dashboardresponse[]))
  }
}


func apicall(method string,endpoint string,  baseUrl string, options []byte) ([]byte, error) {
  switch method {
    case "POST":
      return makePostCall(baseUrl, endpoint, options)
    case "GET":
      return makeGetRequest(baseUrl, endpoint)
    default:
      return  nil,errors.New("Invalid method")
  }
}


func makePostCall(baseUrl string, endpoint string, postBody []byte ) ([]byte, error){
  fmt.Println("endpoiont", fmt.Sprintf("%s%s", baseUrl, endpoint))
  resp, errr := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", baseUrl, endpoint), bytes.NewBuffer(postBody))
  resp.Header.Add("Authorization",  "Bearer <token>")
  resp.Header.Add("Content-Type", "application/json")

  if errr != nil{
    fmt.Println("error during post call:",errr)
  }
  defer resp.Body.Close()
  response, err1 := http.DefaultClient.Do(resp)

  if err1 != nil {
    fmt.Println("error while making a new data source:", err1)
    return nil, err1
  }else{
    fmt.Println("response:", response)
  }
  responseBody, _ := io.ReadAll(response.Body) 
  fmt.Println("response body:", string(responseBody[:]))
  return responseBody, nil
}

func makeGetRequest(baseUrl string, endpoint string) ([]byte, error){

  resp, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", baseUrl, endpoint), nil)
  resp.Header.Add("Authorization", "Bearer <token>")

  response, err := http.DefaultClient.Do(resp)
  defer response.Body.Close()
  if(err != nil){
    fmt.Println("error returned with api call:", err)
    return nil, err
  }
  body, err := io.ReadAll(response.Body)
  return body, nil
}

func getEnv() envVars{
  return envVars{
    grafanaBaseUrl: os.Getenv("GRAFANA_BASE_URL"),
    apiToken: os.Getenv("GRAFANA_API_TOKEN"),
  }
}

func getDashboardJson() []byte{
  dashboardJson, err := os.ReadFile("./dashboard.json")
  if err != nil{
    panic(err)
  }
  return dashboardJson
}

func updateDataSourceInJSON(dashboardJson []byte, datasourceUid string) []byte{
  str := string(dashboardJson[:])

   newString :=strings.ReplaceAll(str, "<replace>", datasourceUid)
  return []byte(newString)
} 
