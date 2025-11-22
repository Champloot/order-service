package main

import (
    "context"
    "encoding/json"
    "log"
    "time"

    "order-service/internal/testdata"
    "order-service/internal/models"


    "github.com/segmentio/kafka-go"
)

func main() {
    testdata.Init()

    writer := &kafka.Writer{
        Addr:         kafka.TCP("localhost:9092"),
        Topic:        "orders",
        Balancer:     &kafka.LeastBytes{},
        BatchTimeout: 10 * time.Millisecond,
    }
    defer writer.Close()

    orders := make([]*models.Order, 1)
    for i := 0; i < 1; i++ {
        orders[i] = testdata.GenerateValidatedOrder()
    }

    for _, order := range orders {
        orderJSON, err := json.Marshal(order)
        if err != nil {
            log.Fatalf("Failed to marshal order: %v", err)
        }

        const maxRetries = 3
        for i := 0; i < maxRetries; i++ {
            err = writer.WriteMessages(context.Background(),
                kafka.Message{
                    Key:   []byte(order.OrderUID),
                    Value: orderJSON,
                },
            )
            if err == nil {
                break
            }
            log.Printf("Attempt %d failed to write message: %v", i+1, err)
            time.Sleep(time.Duration(i+1) * time.Second)
        }
        if err != nil {
            log.Fatalf("Failed to write message after %d attempts: %v", maxRetries, err)
        }

        log.Printf("Message sent successfully: %s", order.OrderUID)
        time.Sleep(100 * time.Millisecond)
    }

    log.Println("All test messages sent successfully")
}