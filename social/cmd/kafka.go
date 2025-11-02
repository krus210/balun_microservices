package main

import (
	"os"
)

var (
	kafkaBrokers                     = os.Getenv("KAFKA_BROKERS")
	kafkaFriendRequestEventTopicName = os.Getenv("KAFKA_FRIEND_REQUEST_EVENTS_TOPIC_NAME")
)
