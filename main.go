package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func main() {
	// set up the topic and consumer group
	ctx := context.Background()
	SetupConsumerGroupLagInTopic(ctx)

	// sleep until the dust settles
	time.Sleep(time.Second * 10)

	// check consumer lag endpoints
	skewGroupEndpoint := fmt.Sprintf("http://localhost:8000/v3/kafka/docker/consumer/%v", SKEWED_CONSUMER_GROUP_ID)
	err := MakeRequestToBurrow(skewGroupEndpoint)
	if err != nil {
		panic(err)
	}

	statusEndpoint := fmt.Sprintf("%v/status", skewGroupEndpoint)
	err = MakeRequestToBurrow(statusEndpoint)
	if err != nil {
		panic(err)
	}

	lagEndpoint := fmt.Sprintf("%v/lag", skewGroupEndpoint)
	err = MakeRequestToBurrow(lagEndpoint)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Deleting Kafka topic")
	DeleteTopic(ctx, SKEWED_TOPIC_NAME)
}

func MakeRequestToBurrow(url string) error {
	fmt.Println(fmt.Sprintf("Making request to Burrow endpoint: %v", url))
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close() // don't forget to close this or else suffer a memory leak!

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var prettyJson bytes.Buffer
	err = json.Indent(&prettyJson, body, "", "\t")
	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf("Response Payload JSON: %v", string(prettyJson.Bytes())))
	return nil
}
