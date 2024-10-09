package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/paudarco/orders-db-task/internal/cache"
	"github.com/paudarco/orders-db-task/internal/database"
	"github.com/paudarco/orders-db-task/internal/models"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	db     *database.Database
	cache  *cache.Cache
}

func NewConsumer(brokers []string, topic string, db *database.Database, cache *cache.Cache) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: "order-service",
	})

	return &Consumer{
		reader: reader,
		db:     db,
		cache:  cache,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading message: %v", err)
				continue
			}

			var order models.Order
			err = json.Unmarshal(msg.Value, &order)
			if err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}

			if !order.IsValidOrder() {
				log.Printf("Invalid order data received: %v", order)
				continue
			}

			err = c.db.SaveOrder(&order)
			if err != nil {
				log.Printf("Error saving order to database: %v", err)
				continue
			}

			c.cache.Set(&order)
			log.Printf("Processed order: %s", order.OrderUID)
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
