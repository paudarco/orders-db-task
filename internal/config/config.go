package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	PostgresURL  string   `json:"postgres_url"`
	KafkaBrokers []string `json:"kafka_brokers"`
	KafkaTopic   string   `json:"kafka_topic"`
	ServerPort   string   `json:"server_port"`
}

func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	// Переопределяем значения из переменных окружения, если они установлены
	if url := os.Getenv("POSTGRES_URL"); url != "" {
		config.PostgresURL = url
	}
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		config.KafkaBrokers = []string{brokers}
	}

	return &config, nil
}
