package testdata

import (
	"fmt"
	"time"

	"order-service/internal/models"

	"github.com/brianvoe/gofakeit/v6"
)

func Init() {
	gofakeit.Seed(0)
}

func GenerateOrder() *models.Order {
	itemsCount := gofakeit.Number(1, 5)
	items := make([]models.Item, itemsCount)
	
	goodsTotal := 0
	for i := 0; i < itemsCount; i++ {
		price := gofakeit.Number(100, 5000)
		totalPrice := price
		
		items[i] = models.Item{
			ChrtID:      gofakeit.Number(1000000, 9999999),
			TrackNumber: fmt.Sprintf("WBILMTEST%s", gofakeit.Regex("[A-Z0-9]{6}")),
			Price:       price,
			Rid:         gofakeit.UUID(),
			Name:        gofakeit.ProductName(),
			Sale:        gofakeit.Number(0, 50),
			Size:        "0",
			TotalPrice:  totalPrice,
			NmID:        gofakeit.Number(1000000, 9999999),
			Brand:       gofakeit.Company(),
			Status:      202,
		}
		goodsTotal += totalPrice
	}

	deliveryCost := gofakeit.Number(300, 2000)
	customFee := gofakeit.Number(0, 100)
	amount := deliveryCost + goodsTotal + customFee

	dateCreated := time.Now().Add(-time.Duration(gofakeit.Number(0, 30)) * 24 * time.Hour)

	return &models.Order{
		OrderUID:    gofakeit.UUID(),
		TrackNumber: fmt.Sprintf("WBILMTEST%s", gofakeit.Regex("[A-Z0-9]{6}")),
		Entry:       "WBIL",
		Delivery: models.Delivery{
			Name:    gofakeit.Name(),
			Phone:   "+1" + fmt.Sprint(gofakeit.Number(1000000000, 9999999999)),
			Zip:     gofakeit.Zip(),
			City:    gofakeit.City(),
			Address: gofakeit.Street(),
			Region:  gofakeit.State(),
			Email:   gofakeit.Email(),
		},
		Payment: models.Payment{
			Transaction:  gofakeit.UUID(),
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       amount,
			PaymentDt:    gofakeit.DateRange(time.Now().Add(-365*24*time.Hour), time.Now()).Unix(),
			Bank:         "alpha",
			DeliveryCost: deliveryCost,
			GoodsTotal:   goodsTotal,
			CustomFee:    customFee,
		},
		Items:              items,
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       dateCreated,
		OofShard:          "1",
	}
}

func GenerateOrders(count int) []*models.Order {
	orders := make([]*models.Order, count)
	for i := 0; i < count; i++ {
		orders[i] = GenerateOrder()
	}
	return orders
}

func GenerateOrderWithID(orderUID string) *models.Order {
	order := GenerateOrder()
	order.OrderUID = orderUID
	return order
}

func GenerateValidatedOrder() *models.Order {
	order := GenerateOrder()
	
	order.TrackNumber = "WBILMTESTTRACK"
	order.Entry = "WBIL" 
	order.CustomerID = "test"
	order.DeliveryService = "meest"
	order.Locale = "en"
	
	return order
}