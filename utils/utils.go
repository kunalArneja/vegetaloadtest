package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/perf/parser"

	"github.com/spf13/cast"
	vegeta "github.com/tsenart/vegeta/lib"
)

func GetHttpHeaders(headers map[string]string) http.Header {
	header := http.Header{}
	for headerKey, headerValue := range headers {
		header.Add(headerKey, headerValue)
	}
	return header
}

func GetTargeter(url string, httpmethod string, headers http.Header, body string, dynamicFields map[string]string, requestDumpFileWriter *os.File, dynamicHeaders map[string]map[string]string) vegeta.Targeter {
	return func() vegeta.Targeter {
		return func(t *vegeta.Target) (err error) {
			transformedURL := ""
			jsonStringForRequest := ""

			if body != "" && strings.ToUpper(httpmethod) != "GET" {
				transformedURL, jsonStringForRequest = parser.GetUrlAndBody(url, body, dynamicFields)
				t.Body = []byte(jsonStringForRequest)

				dHeaders := GetDynamicHeaders(dynamicHeaders, t.Body)

				for key, value := range dHeaders {
					headers.Del(key)
					headers.Add(key, value)
				}
				//hack determine a better way to do this. Have to dig into vegeta docs, if there is a handle to requests object in targetter
				//spin off a go routine now
				if requestDumpFileWriter != nil {
					go requestDumpFileWriter.WriteString(jsonStringForRequest + " \n")
				}
			}
			t.Method = httpmethod
			t.URL = transformedURL
			t.Header = headers
			return err
		}
	}()
}

func DumpReportToFile(reporter vegeta.Reporter, writer io.Writer) {
	if writer != nil {
		err := reporter.Report(writer)
		if err != nil {
			fmt.Errorf("Error %s", err)
		}
	}
}

func OpenFileCreateIfNotFound(filePath string) (*os.File, error) {
	if filePath == "" {
		fmt.Errorf("Invalid file path %s ", filePath)
		return nil, errors.New("Invalid file path")
	}
	return os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0777)
}

func ProcessReport(reporter vegeta.Reporter, filePath string) {
	//debug
	fmt.Printf("report %s", reporter.Report(os.Stdout))
	file, err := OpenFileCreateIfNotFound(filePath)

	if err != nil {
		fmt.Errorf("Error dumping into results file %s ", err)
		return

		DumpReportToFile(reporter, file)
	}
}

func CreateMACofBodyandKey(message, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func GetDynamicHeaders(in map[string]map[string]string, body []byte) map[string]string {
	dHeaders := make(map[string]string)
	for headerKey, headerValue := range in {
		key := headerKey
		value := CreateMACofBodyandKey(body, []byte(headerValue["key"]))

		dHeaders[key] = value
	}

	return dHeaders
}

func ConvertToMapStringMapStringString(source map[string]interface{}) map[string]map[string]string {
	mapStringMapStringString := make(map[string]map[string]string)
	for key, value := range source {
		mapStringMapStringString[key] = ConvertToMapStringString(value)
	}
	return mapStringMapStringString
}

func ConvertToMapStringString(in interface{}) map[string]string {
	return cast.ToStringMapString(in)
}
