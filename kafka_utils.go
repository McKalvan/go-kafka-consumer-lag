package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const (
	SKEWED_TOPIC_NAME        = "skewed-topic"
	SKEWED_CONSUMER_GROUP_ID = "skewed-consumer-group"
)

var (
	config = kafka.ConfigMap{
		"bootstrap.servers": kafka.ConfigValue(os.Getenv("BOOTSTRAP_SERVERS")),
	}
)

/*
Sets up test environment with partition skew in topic
*/
func SetupConsumerGroupLagInTopic(ctx context.Context) {
	fmt.Println(fmt.Sprintf("Creating skewed topic, '%v'", SKEWED_TOPIC_NAME))
	if err := CreateTopic(ctx, SKEWED_TOPIC_NAME, 3); err != nil {
		panic(err)
	}

	fmt.Println("Producing messages to Kafka topic")
	if err := ProduceSkewedMessagesToTopic(SKEWED_TOPIC_NAME, 1000, 3); err != nil {
		panic(err)
	}

	fmt.Println("Creating consumer group for skewed topic")
	if err := CreateConsumerGroup(SKEWED_TOPIC_NAME, SKEWED_CONSUMER_GROUP_ID, 3); err != nil {
		panic(err)
	}
}

/*
Creates topic in our dummy cluster
*/
func CreateTopic(ctx context.Context, topic string, numPartitions int) error {
	adminClient, err := kafka.NewAdminClient(&config)
	if err != nil {
		return err
	}
	defer adminClient.Close()

	topicSpec := kafka.TopicSpecification{
		Topic:             topic,
		NumPartitions:     numPartitions,
		ReplicationFactor: 3,
	}

	_, err = adminClient.CreateTopics(ctx, []kafka.TopicSpecification{topicSpec})
	return err
}

/*
Produces n dummy messages to topic,
attempts to replicate partition skew by assigning more messages to one partition
*/
func ProduceSkewedMessagesToTopic(topic string, numMessages int, numPartitions int) error {
	producer, err := kafka.NewProducer(&config)
	if err != nil {
		return err
	}
	defer producer.Close()

	msgArr := make([]int, numMessages)
	for i := 0; i < numMessages; i++ {
		msgArr[i] = i
	}

	for _, msg := range msgArr {
		var partitionNum int32
		// all even message vals get routed to the first partition to replicate partition skew
		if msg%2 == 0 {
			partitionNum = 1
		} else {
			partitionNum = int32(msg % numPartitions)
		}

		producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: partitionNum,
			},
			Value:     []byte(string(msg)),
			Timestamp: time.Now(),
		}, nil)
	}

	producer.Flush(10 * 1000)
	fmt.Println("Messages produced")
	return nil
}

/*
Reads one message off the topic to initiate the consumer group
*/
func CreateConsumerGroup(topic string, consumerGroupName string, numPartitions int) error {
	config.SetKey("group.id", consumerGroupName)
	config.SetKey("auto.offset.reset", "earliest")

	consumer, err := kafka.NewConsumer(&config)
	if err != nil {
		return err
	}
	defer consumer.Close()
	consumer.Subscribe(topic, nil)

	readPartitionsBitMap := map[int32]bool{}
	for len(readPartitionsBitMap) != numPartitions {
		message, _ := consumer.ReadMessage(10 * time.Second)
		fmt.Printf("Read message from partition: %v, message: %v\n", message.TopicPartition.Partition, message.Value)
		readPartitionsBitMap[message.TopicPartition.Partition] = true
	}
	consumer.Commit()
	return nil
}

func DeleteTopic(ctx context.Context, topic string) error {
	adminClient, err := kafka.NewAdminClient(&config)
	if err != nil {
		return err
	}
	defer adminClient.Close()

	_, err = adminClient.DeleteTopics(ctx, []string{topic}, nil)
	if err != nil {
		return err
	}
	return nil
}
