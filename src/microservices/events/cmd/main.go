package main

import (
	"context"
	"log"
	"os"
	"oyevents/internal/consumer"
	"oyevents/internal/producer"
	"oyevents/internal/types"
	"oyevents/internal/webapi"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

func main() {
	time.Sleep(10 * time.Second) // wait kafka
	data, err := os.ReadFile("config.yml")
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var cfg types.AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers != "" {
		cfg.ProducerConfig.BootstrapServers = brokers
		cfg.ConsumerConfig.BootstrapServers = brokers
	}
	webapiPort := os.Getenv("PORT")
	if webapiPort != "" {
		cfg.WebapiConfig.Listen = webapiPort
	}

	events := make(chan types.EventMessage, cfg.CommonConfig.EventsQueueSize)
	api := webapi.NewEventsWebapi(cfg.WebapiConfig, events)
	p := producer.NewEventsProducer(cfg.ProducerConfig)
	c, err := consumer.NewEventsConsumer(cfg.Topics, cfg.ConsumerConfig)
	if err != nil {
		log.Fatal(err.Error())
	}

	var wg sync.WaitGroup
	ctx := context.Background()
	go p.Run(ctx, &wg, events)
	go c.Run(ctx, &wg)
	go api.Run(ctx, &wg)
	wg.Add(3)
	wg.Wait()
}
