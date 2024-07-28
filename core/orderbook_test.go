package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLimit(t *testing.T) {
	limit := NewLimit(10_000)
	order := NewOrder(2, true, 3_000)
	limit.AddOrder(order)

	limit.Orders.Each(func(key int64, val *Order) {
		fmt.Println(key, val)
	})

	fetchedOrder, _ := limit.Orders.Get(order.Timestamp)
	fmt.Printf("Fetched order: %v\n", fetchedOrder)
	assert.Equal(t, fetchedOrder, order)

	// removing orders
	limit.RemoveOrders([]*Order{order})
	assert.Equal(t, limit.TotalVolume, float64(0))

	fetchedOrder1, _ := limit.Orders.Get(order.Timestamp)
	assert.Nil(t, fetchedOrder1)

	// adding multiple orders
	orders := []*Order{
		NewOrder(1, true, 1_000),
		NewOrder(2, true, 2_000),
		NewOrder(3, true, 3_000),
	}

	for _, o := range orders {
		limit.AddOrder(o)
	}

	assert.Equal(t, limit.TotalVolume, float64(14_000))

	// removing 2nd Order
	limit.RemoveOrders([]*Order{orders[1]})

	assert.Equal(t, limit.TotalVolume, float64(10_000))
}

func TestOrderBook(t *testing.T) {
	ob := NewOrderBook()
	order := NewOrder(3, true, 400)
	ob.PlaceLimitOrder(1300, order)

	expectedOrder, _ := ob.BidsMap[1300].Orders.Get(order.Timestamp)
	assert.Equal(t, expectedOrder, order)
	assert.Equal(t, ob.BidsMap[1300].TotalVolume, float64(1200))
}

func TestSortingOfLimitsAndOrders(t *testing.T) {
	ob := NewOrderBook()
	orders := []*Order{
		NewOrder(3, false, 400),
		NewOrder(3, false, 400),
		NewOrder(2, false, 300),
		NewOrder(1, false, 200),
	}

	for _, o := range orders {
		ob.PlaceLimitOrder(float64(o.Size)*o.Price, o)
	}

	ob.Asks.Each(func(key float64, val *Limit) {
		fmt.Println(key, val)
	})

	ob.AsksMap[1200].Orders.Each(func(key int64, val *Order) {
		fmt.Println(key, val)
	})

	// assert.NotNil(t, nil)

	assert.Equal(t, ob.Asks.Size(), 3)
}

func TestPlaceMarketOrder(t *testing.T) {
	ob := NewOrderBook()

	// selling 3 BTC
	sellOrder := NewOrder(3, false, 400)
	sellOrder1 := NewOrder(5, false, 300)
	ob.PlaceLimitOrder(sellOrder.TotalPrice(), sellOrder)
	ob.PlaceLimitOrder(sellOrder1.TotalPrice(), sellOrder1)

	// buying 5 BTC
	buyOrder := NewOrder(5, true, 400)
	matches := ob.PlaceMarketOrder(buyOrder.TotalPrice(), buyOrder)
	fmt.Printf("Matches: %v\n", matches)

	assert.Equal(t, len(matches), 2)
	assert.Equal(t, ob.Asks.Size(), 2)
	assert.Equal(t, ob.totalAskVolume, float64(700))
	assert.Equal(t, matches[0].Price, float64(1200))

	// assert.NotNil(t, nil)
}
