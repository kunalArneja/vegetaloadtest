package parser

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tidwall/sjson"
)

var Counter = 0
var CounterIncrement = 0

func GetJsonString(filePath string) string {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	jsonString := string(byteValue[:])

	return jsonString
}

func GetUrlAndBody(url string, jsonString string, dynamicfields map[string]string) (urlRet string, body string) {
	newJSON := jsonString
	attrValue := ""
	for key, value := range dynamicfields {
		switch value {
		case "timestamp":
			attrValue = time.Now().UTC().Format(time.RFC3339)
			newJSON, _ = sjson.Set(newJSON, key, attrValue)
		case "uuid":
			attrValue = uuid.New().String()
			newJSON, _ = sjson.Set(newJSON, key, attrValue)
		case "epoch":
			attrValue = strconv.FormatInt(time.Now().Unix(), 10)
			newJSON, _ = sjson.Set(newJSON, key, attrValue)
		case "epochnano":
			attrValue = strconv.FormatInt(time.Now().UnixNano(), 10)
			newJSON, _ = sjson.Set(newJSON, key, attrValue)
		case "identity":
			attrValue = strconv.Itoa(Counter)
			newJSON, _ = sjson.Set(newJSON, key, attrValue)
			Counter = Counter + CounterIncrement
		}
		searchString := "{" + key + "}"
		if strings.Contains(url, searchString) {
			url = strings.Replace(url, searchString, attrValue, 1)
		}
	}
	return url, newJSON
}
