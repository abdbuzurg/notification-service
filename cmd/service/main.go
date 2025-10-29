package main

import (
	pb "asr_leasing_notification/asr_leasing_notification/protos"
	"asr_leasing_notification/internal/platform/database"
	"asr_leasing_notification/internal/platform/message_broker"
	publisher_handler "asr_leasing_notification/internal/publisher/handler"
	"asr_leasing_notification/internal/server"
	subscriber_handler "asr_leasing_notification/internal/subscriber/handler"
	"asr_leasing_notification/internal/subscriber/notifiers"
	subscriber_repository "asr_leasing_notification/internal/subscriber/repository"
	subscriber_usecase "asr_leasing_notification/internal/subscriber/usecase"
	"log"
	"net"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}
}

func main() {
	dbPool, err := database.NewPostgresConnection()
	if err != nil {
		log.Fatalf("DB connection failed: %s", err)
	}
	defer dbPool.Close()

	rabbitMQConn, err := message_broker.NewRabbitMQConn()
	if err != nil {
		log.Fatalf("RabbitMQ connection failed: %s", err)
	}
	defer rabbitMQConn.Close()

	// Publisher
	amqpPublisher := publisher_handler.NewAMQPPublisher(rabbitMQConn)

	//Subscriber
	subscriberRepo := subscriber_repository.New(dbPool)
	notifiers := map[subscriber_repository.NotificationChannel]notifiers.Notifier{
		subscriber_repository.NotificationChannelEMAIL:    notifiers.NewEmailNotifier(),
		subscriber_repository.NotificationChannelSMS:      notifiers.NewSmsNotifier(),
		subscriber_repository.NotificationChannelPUSH:     notifiers.NewPushNotifier(),
		subscriber_repository.NotificationChannelTELEGRAM: notifiers.NewTelegramNotifier(),
	}
	subscriberUC := subscriber_usecase.NewNotificationService(subscriberRepo, notifiers, amqpPublisher)
	amqpSubscriber := subscriber_handler.NewAMQPSubscriber(rabbitMQConn, subscriberUC)

	go func() {
		queueNames := viper.GetStringSlice("rabbitmq_queue_names")
		if len(queueNames) == 0 {
			log.Fatal("No queue names found for subscriber")
		}
		log.Println("Starting background subscriber...")
		if err := amqpSubscriber.Start(queueNames); err != nil {
			log.Fatalf("Subscriber fatal error: %s", err)
		}
	}()

	// gRPC
	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("Failed to listen to port 9090: %v", err)
	}

	grpcServer := grpc.NewServer()
	notificationServer := server.NewGRPCServer(subscriberRepo, notifiers)
	pb.RegisterNotificationServiceServer(grpcServer, notificationServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC server failed: %v", err)
	}
}
