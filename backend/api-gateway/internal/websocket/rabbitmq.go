package websocket

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ConsumeNotifications подписывается на очередь RabbitMQ и отправляет сообщения в Hub
func ConsumeNotifications(hub *Hub, rabbitURL string) {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open channel: %v", err)
	}
	defer ch.Close()

	// Объявляем обменник и очереди (если не созданы)
	err = ch.ExchangeDeclare(
		"notifications", // name
		"topic",         // type
		true,            // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		log.Fatalf("failed to declare exchange: %v", err)
	}

	// Создаём временную очередь, которая получает все уведомления
	q, err := ch.QueueDeclare(
		"",    // name (случайное)
		false, // durable
		false, // auto-delete (удаляем при отключении)
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Fatalf("failed to declare queue: %v", err)
	}

	// Привязываем очередь к обменнику с routing key #
	err = ch.QueueBind(
		q.Name,           // queue name
		"notification.#", // routing key (все уведомления)
		"notifications",  // exchange
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("failed to bind queue: %v", err)
	}

	deliveries, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("failed to consume: %v", err)
	}

	log.Println("Connected to RabbitMQ, waiting for notifications...")

	for d := range deliveries {
		// Парсим сообщение (можно добавить обработку)
		var msg map[string]interface{}
		if err := json.Unmarshal(d.Body, &msg); err != nil {
			log.Printf("failed to parse message: %v", err)
			continue
		}
		// Отправляем в Hub
		hub.Broadcast(d.Body)
		log.Printf("Broadcast notification: %s", d.Body)
	}
}
