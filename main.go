package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx := context.Background()
	SetupConsumerGroupLagInTopic(ctx)
	time.Sleep(time.Second * 10)

	fmt.Printf("Deleting Kafka topic")
	// DeleteTopic(ctx, SKEWED_TOPIC_NAME)
}
