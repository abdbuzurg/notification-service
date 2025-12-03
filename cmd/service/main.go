package main

import (
	"asr_leasing_notification/internal/platform/database"
	"asr_leasing_notification/internal/platform/message_broker"
	publisher_handler "asr_leasing_notification/internal/publisher/handler"
	subscriber_handler "asr_leasing_notification/internal/subscriber/handler"
	"asr_leasing_notification/internal/subscriber/notifiers"
	subscriber_repository "asr_leasing_notification/internal/subscriber/repository"
	subscriber_usecase "asr_leasing_notification/internal/subscriber/usecase"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
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
		subscriber_repository.NotificationChannelEMAIL: notifiers.NewEmailNotifier(),
		subscriber_repository.NotificationChannelSMS: notifiers.NewOsonSMSNotifier(
			subscriberRepo,
			viper.GetString("oson_sms.login"),
			viper.GetString("oson_sms.from"),
			viper.GetString("oson_sms.hash"),
			viper.GetString("oson_sms.url"),
		),
		subscriber_repository.NotificationChannelPUSH:     notifiers.NewPushNotifier(),
		subscriber_repository.NotificationChannelTELEGRAM: notifiers.NewTelegramNotifier(),
	}
	subscriberUC := subscriber_usecase.NewNotificationService(subscriberRepo, notifiers, amqpPublisher)
	amqpSubscriber := subscriber_handler.NewAMQPSubscriber(rabbitMQConn, subscriberUC)

	go func() {
		queueNames := viper.GetStringSlice("rabbitmq.queue_names")
		if len(queueNames) == 0 {
			log.Fatal("No queue names found for subscriber")
		}
		log.Println("Starting background subscriber...")
		if err := amqpSubscriber.Start(queueNames); err != nil {
			log.Fatalf("Subscriber fatal error: %s", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	log.Println("Shutting down gracefully...")
}
