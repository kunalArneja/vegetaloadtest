package main

import (
	"github.com/perf/parser"
  "time"
  "github.com/perf/config"
  "github.com/perf/utils"
  vegeta "github.com/tsenart/vegeta/lib"
)

func main() {
  var metrics vegeta.Metrics

  conf:= config.InitConfig("../config.yaml")
  
  //refactor this into constants file
  URL := conf.GetString("url")
  HTTP_METHOD := conf.GetString("httpmethod")
  VEGETA_RATE := conf.GetInt("rate")
  DURATION := conf.GetInt("duration")
  HEADERS := conf.GetStringMapString("headers")
  JSON_FILE_PATH := conf.GetString("post-request-json-file-path")
  DYNAMIC_FIELDS := conf.GetStringMapString("post-request-json-dynamic-fields")
  RESULTS_FILE_PATH := conf.GetString("attack-results-file-path")

  http_headers := utils.GetHttpHeaders(HEADERS)
  test_rate := vegeta.Rate{Freq: VEGETA_RATE, Per: time.Second}
  test_duration := time.Duration(DURATION) * (time.Second)

  jsonString := parser.GetJsonString(JSON_FILE_PATH)
  targeter := utils.GetTargeter(URL, HTTP_METHOD, http_headers, jsonString, DYNAMIC_FIELDS)
  attacker := vegeta.NewAttacker()

  for res := range attacker.Attack(targeter, test_rate, test_duration, "Bang!") {
    metrics.Add(res)
  }

  reporter := vegeta.NewJSONReporter(&metrics)
  metrics.Close()
  utils.ProcessReport(reporter, RESULTS_FILE_PATH)
}