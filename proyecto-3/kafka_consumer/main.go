package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

// Estructura de los mensajes (igual que WeatherTweetRequest del proto)
type WeatherTweet struct {
	Municipality int32 `json:"municipality"`
	Temperature  int32 `json:"temperature"`
	Humidity     int32 `json:"humidity"`
	Weather      int32 `json:"weather"`
}

func main() {
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "kafka-service:9092"
	}

	valkeyAddr := os.Getenv("VALKEY_ADDR")
	if valkeyAddr == "" {
		valkeyAddr = "valkey-service:6379"
	}

	ctx := context.Background()

	// Conexión a Valkey (Redis)
	rdb := redis.NewClient(&redis.Options{
		Addr:     valkeyAddr,
		Password: "",
		DB:       0,
	})
	defer rdb.Close()

	// Conexión al topic de Kafka
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaBroker},
		Topic:    "weathertweets",
		GroupID:  "valkey-consumer",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	log.Println("📥 Kafka Consumer started — listening for weather tweets...")

	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("❌ Error reading from Kafka: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		var tweet WeatherTweet
		if err := json.Unmarshal(m.Value, &tweet); err != nil {
			log.Printf("❌ Error decoding JSON: %v", err)
			continue
		}

		// Generar una clave por municipio y fecha/hora
		key := fmt.Sprintf("tweet:%d:%d", tweet.Municipality, time.Now().Unix())

		// Guardar en Valkey
		err = rdb.HSet(ctx, key, map[string]interface{}{
			"municipality": tweet.Municipality,
			"temperature":  tweet.Temperature,
			"humidity":     tweet.Humidity,
			"weather":      tweet.Weather,
			"timestamp":    time.Now().Format(time.RFC3339),
		}).Err()

		if err != nil {
			log.Printf("⚠️ Error writing to Valkey: %v", err)
		} else {
			log.Printf("✅ Stored in Valkey: %s", key)
		}
	}
}
