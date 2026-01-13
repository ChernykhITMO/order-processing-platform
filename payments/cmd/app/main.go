package main

import (
	"log"
	"sync"

	"github.com/ChernykhITMO/order-processing-platform/payments/internal/handler"
	"github.com/ChernykhITMO/order-processing-platform/payments/internal/kafka_consume"
)

var address = []string{"localhost:9092"}

func main() {
	h := handler.NewHandler()
	c, err := kafka_consume.NewConsumer(h, address, "order-topic", "my-group")
	if err != nil {
		log.Fatal(err)
	}
	c.Start()

	wg := new(sync.WaitGroup)
	go func() {
		defer wg.Done()
		c.Start()
	}()

	wg.Wait()
}
