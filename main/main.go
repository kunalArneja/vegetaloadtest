package main

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	"github.com/perf/config"
	"github.com/perf/parser"
	"github.com/perf/utils"
	vegeta "github.com/tsenart/vegeta/lib"
)

func main() {
	var metrics vegeta.Metrics

	//conf:= config.InitConfig("../shopify/orders-qa.yaml")
	confArg := flag.String("config", "../magento/shipment-config.yaml", "path to config yml")
	flag.Parse()
	fmt.Println(*confArg)
	//panic(0)
	conf := config.InitConfig(*confArg)

	//refactor this into constants file
	URL := conf.GetString("url")
	HTTP_METHOD := conf.GetString("httpmethod")
	VEGETA_RATE := conf.GetInt("rate")
	DURATION := conf.GetInt("duration")
	HEADERS := conf.GetStringMapString("static-headers")
	DYNAMIC_HEADERS := conf.GetStringMap("dynamic-headers")

	dynamic_headers := utils.ConvertToMapStringMapStringString(DYNAMIC_HEADERS)

	JSON_FILE_PATH := conf.GetString("post-request-json-file-path")
	DYNAMIC_FIELDS := conf.GetStringMapString("post-request-json-dynamic-fields")
	RESULTS_FILE_PATH := conf.GetString("dump-attack-results-file-path")
	REQUESTS_FILE_PATH := conf.GetString("dump-request-file-path")
	IDENTITY := conf.GetStringMapString("identity")
	IDENTITY_SHIPMENT := conf.GetStringMapString("identity-shipment")

	identityStart, exists := IDENTITY["start"]
	identityIncr, _ := IDENTITY["increment"]

	if exists {
		parser.Counter, _ = strconv.Atoi(identityStart)
		parser.CounterIncrement, _ = strconv.Atoi(identityIncr)
	}

	identityShipmentStart, exists := IDENTITY_SHIPMENT["start"]
	identityShipmentIncr, _ := IDENTITY_SHIPMENT["increment"]

	if exists {
		parser.ShipmentCounter, _ = strconv.Atoi(identityShipmentStart)
		parser.ShipmentCounterIncrement, _ = strconv.Atoi(identityShipmentIncr)
	}

	http_headers := utils.GetHttpHeaders(HEADERS)
	test_rate := vegeta.Rate{Freq: VEGETA_RATE, Per: time.Second}
	test_duration := time.Duration(DURATION) * (time.Second)

	requestsFileWriter, _ := utils.OpenFileCreateIfNotFound(REQUESTS_FILE_PATH)
	jsonString := parser.GetJsonString(JSON_FILE_PATH)

	targeter := utils.GetTargeter(URL, HTTP_METHOD, http_headers, jsonString, DYNAMIC_FIELDS, requestsFileWriter, dynamic_headers)
	attacker := vegeta.NewAttacker()
	vegeta.Timeout(2 * time.Hour)

	for res := range attacker.Attack(targeter, test_rate, test_duration, "Bang!") {
		metrics.Add(res)
	}

	reporter := vegeta.NewJSONReporter(&metrics)
	metrics.Close()
	utils.ProcessReport(reporter, RESULTS_FILE_PATH)
}
