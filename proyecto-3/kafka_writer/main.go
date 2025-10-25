package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	pb "kafka-writer/proto"

	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
)

type kafkaServer struct {
	pb.UnimplementedWeatherTweetServiceServer
	KafkaBroker string
}

func (s *kafkaServer) SendWeatherTweet(ctx context.Context, req *pb.WeatherTweetRequest) (*pb.WeatherTweetResponse, error) {
	fmt.Printf("📦 [Kafka Writer] Received tweet: %+v\n", req)
	data, _ := json.Marshal(req)

	writer := &kafka.Writer{
		Addr:     kafka.TCP(s.KafkaBroker),
		Topic:    "weathertweets",
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	err := writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(fmt.Sprintf("%d", req.Municipality)),
		Value: data,
	})
	if err != nil {
		log.Printf("Error sending to Kafka: %v", err)
		return &pb.WeatherTweetResponse{Status: "Failed to send tweet to Kafka"}, err
	}

	log.Println("✅ Tweet sent to Kafka successfully")
	return &pb.WeatherTweetResponse{Status: "Sent to Kafka successfully"}, nil
}

func main() {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "kafka-service:9092"
	}

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Error listening: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterWeatherTweetServiceServer(s, &kafkaServer{KafkaBroker: broker})

	log.Println("🚀 Kafka gRPC Writer running on port 50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
