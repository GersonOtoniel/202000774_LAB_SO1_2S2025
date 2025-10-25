package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type WeatherTweet struct {
	Municipality int32 `json:"municipality"`
	Temperature  int32 `json:"temperature"`
	Humidity     int32 `json:"humidity"`
	Weather      int32 `json:"weather"`
}

func main() {
	ctx := context.Background()

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@rabbitmq-service:5672/"
	}

	valkeyAddr := os.Getenv("VALKEY_ADDR")
	if valkeyAddr == "" {
		valkeyAddr = "valkey-service:6379"
	}

	// Conexión a Valkey (Redis)
	rdb := redis.NewClient(&redis.Options{
		Addr:     valkeyAddr,
		Password: "",
		DB:       0,
	})

	// Conexión a RabbitMQ
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("❌ Cannot connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("❌ Cannot open channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"weathertweets",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("❌ Cannot declare queue: %v", err)
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("❌ Cannot consume queue: %v", err)
	}

	log.Println("📥 [RabbitMQ Consumer] Listening for messages...")

	for msg := range msgs {
		var tweet WeatherTweet
		if err := json.Unmarshal(msg.Body, &tweet); err != nil {
			log.Printf("❌ Error decoding message: %v", err)
			continue
		}

		key := fmt.Sprintf("tweet:%d:%d", tweet.Municipality, time.Now().Unix())
		err = rdb.HSet(ctx, key, map[string]interface{}{
			"municipality": tweet.Municipality,
			"temperature":  tweet.Temperature,
			"humidity":     tweet.Humidity,
			"weather":      tweet.Weather,
			"timestamp":    time.Now().Format(time.RFC3339),
		}).Err()
		if err != nil {
			log.Printf("⚠️ Error saving to Valkey: %v", err)
		} else {
			log.Printf("✅ Saved in Valkey → %s", key)
		}
	}
}
