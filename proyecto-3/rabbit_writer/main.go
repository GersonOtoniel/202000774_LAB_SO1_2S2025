package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	pb "writer-rabbit/proto"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
)

type rabbitServer struct {
	pb.UnimplementedWeatherTweetServiceServer
	RabbitURL string
}

func (s *rabbitServer) SendWeatherTweet(ctx context.Context, req *pb.WeatherTweetRequest) (*pb.WeatherTweetResponse, error) {
	fmt.Printf("🕊️ [RabbitMQ Writer] Received tweet: %+v\n", req)
	data, _ := json.Marshal(req)

	conn, err := amqp.Dial(s.RabbitURL)
	if err != nil {
		log.Printf("Error connecting to RabbitMQ: %v", err)
		return &pb.WeatherTweetResponse{Status: "Failed to connect to RabbitMQ"}, err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("Error creating channel: %v", err)
		return &pb.WeatherTweetResponse{Status: "Failed to create channel"}, err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"weathertweets",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Error declaring queue: %v", err)
		return &pb.WeatherTweetResponse{Status: "Queue declare failed"}, err
	}

	err = ch.Publish("", q.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	})
	if err != nil {
		log.Printf("Error publishing message: %v", err)
		return &pb.WeatherTweetResponse{Status: "Publish failed"}, err
	}

	log.Println("✅ Tweet sent to RabbitMQ successfully")
	return &pb.WeatherTweetResponse{Status: "Sent to RabbitMQ successfully"}, nil
}

func main() {
	rabbit := os.Getenv("RABBITMQ_URL")
	if rabbit == "" {
		rabbit = "amqp://guest:guest@rabbitmq-service:5672/"
	}

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("Error listening: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterWeatherTweetServiceServer(s, &rabbitServer{RabbitURL: rabbit})

	log.Println("🚀 RabbitMQ gRPC Writer running on port 50053")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
