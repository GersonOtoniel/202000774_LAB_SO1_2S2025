package main

import (
	"context"
	"fmt"
	pb "go-deploy/proto"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	//"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type server struct {
	pb.UnimplementedWeatherTweetServiceServer
	KafkaAddr  string
	RabbitAddr string
}

func (s *server) SendWeatherTweet(ctx context.Context, req *pb.WeatherTweetRequest) (*pb.WeatherTweetResponse, error) {
	fmt.Printf("Received tweet request: %+v\n", req)
	rand.New(rand.NewSource(time.Now().UnixNano()))
	useKafka := rand.Intn(2) == 0
	var (
		resp *pb.WeatherTweetResponse
		err  error
	)
	if useKafka {
		fmt.Println("Enviando mensaje a Kafka...")
		resp, err = forwardToKafka(s.KafkaAddr, req)
		if err != nil {
			return &pb.WeatherTweetResponse{Status: "Error forwarding to Kafka"}, err
		}
	} else {
		fmt.Println("Enviando mensaje a RabbitMQ...")
		resp, err = forwardToRabbitMQ(s.RabbitAddr, req)
		if err != nil {
			return &pb.WeatherTweetResponse{Status: "Error forwarding to RabbitMQ"}, err
		}
	}
	//log.Printf("Received deployment request for service: %s, version: %s", req.ServiceName, req.Version)
	// Here you would add your deployment logic
	//return &pb.WeatherTweetResponse{Status: "Tweet received! "}, nil
	return resp, nil
}

func forwardToKafka(addr string, request *pb.WeatherTweetRequest) (*pb.WeatherTweetResponse, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewWeatherTweetServiceClient(conn)
	resp, err := client.SendWeatherTweet(context.Background(), request)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func forwardToRabbitMQ(addr string, request *pb.WeatherTweetRequest) (*pb.WeatherTweetResponse, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := pb.NewWeatherTweetServiceClient(conn)
	resp, err := client.SendWeatherTweet(context.Background(), request)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	kafka := os.Getenv("KAFKA_GRPC_ADDR")
	if kafka == "" {
		kafka = "server-go-kafka:50052"
	}
	rabbit := os.Getenv("RABBITMQ_GRPC_ADDR")
	if rabbit == "" {
		rabbit = "server-go-rabbit:50053"
	}
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterWeatherTweetServiceServer(s, &server{
		KafkaAddr:  kafka,
		RabbitAddr: rabbit,
	})
	fmt.Println("gRPC server listening on port 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
