package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/paudarco/orders-db-task/internal/models"
	"github.com/segmentio/kafka-go"
)

func main() {
	// Инициализация генератора случайных чисел
	rand.Seed(time.Now().UnixNano())

	// Создание писателя Kafka
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"host.docker.internal:29092"},
		Topic:   "orders",
	})
	defer w.Close()

	// Создание и публикация трех заказов
	for i := 1; i <= 3; i++ {
		order := generateRandomOrder(i)
		publishOrder(w, order)
	}
}

func generateRandomOrder(index int) models.Order {
	return models.Order{
		OrderUID:    fmt.Sprintf("test_order_%d", index),
		TrackNumber: fmt.Sprintf("TRACK%d", rand.Intn(1000000)),
		Entry:       "WBIL",
		Delivery: models.Delivery{
			Name:    fmt.Sprintf("Customer %d", index),
			Phone:   fmt.Sprintf("+1234567890%d", index),
			Zip:     fmt.Sprintf("1000%d", index),
			City:    "New York",
			Address: fmt.Sprintf("%d Broadway St", rand.Intn(1000)),
			Region:  "NY",
			Email:   fmt.Sprintf("customer%d@example.com", index),
		},
		Payment: models.Payment{
			Transaction:  fmt.Sprintf("tr%d", rand.Intn(1000000)),
			RequestID:    fmt.Sprint(rand.Intn(1000000)),
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       rand.Intn(10000) + 1000,
			PaymentDt:    time.Now().Unix(),
			Bank:         "JPMorgan",
			DeliveryCost: rand.Intn(1000) + 500,
			GoodsTotal:   rand.Intn(5000) + 500,
			CustomFee:    index,
		},
		Items: []models.Item{
			{
				ChrtID:      rand.Intn(10000000),
				TrackNumber: fmt.Sprintf("TRACK%d", rand.Intn(1000000)),
				Price:       rand.Intn(1000) + 100,
				Rid:         fmt.Sprintf("rid%d", rand.Intn(1000000)),
				Name:        fmt.Sprintf("Product %d", index),
				Sale:        rand.Intn(50),
				Size:        "M",
				TotalPrice:  rand.Intn(1000) + 200,
				NmID:        rand.Intn(1000000),
				Brand:       "BrandName",
				Status:      202,
			},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        fmt.Sprintf("cust%d", rand.Intn(1000)),
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       time.Now(),
		OofShard:          "1",
	}
}

func publishOrder(w *kafka.Writer, order models.Order) {
	value, err := json.Marshal(order)
	if err != nil {
		log.Printf("Error marshaling order: %v", err)
		return
	}

	err = w.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(order.OrderUID),
			Value: value,
		},
	)
	if err != nil {
		log.Printf("Error publishing message: %v", err)
	} else {
		log.Printf("Successfully published order: %s", order.OrderUID)
	}
}
