package sqs_listener

import "strings"

func cleanSQSUrl(url string) string {
	// https://sqs.us-west-2.amazonaws.com/490047155434/trademarkia-ElasticImporterFunctionQueue-qG7FSuX8Iyup
	split := strings.Split(url, "/")
	return split[len(split)-1]
}
