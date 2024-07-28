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
	ob.PlaceOrder(1300, order)

	expectedOrder, _ := ob.BidsMap[1300].Orders.Get(order.Timestamp)
	assert.Equal(t, expectedOrder, order)
	assert.Equal(t, ob.BidsMap[1300].TotalVolume, float64(1200))
}

func TestSortingOfLimitsAndOrders(t *testing.T) {
	ob := NewOrderBook()
	orders := []*Order{
		NewOrder(3, true, 400),
		NewOrder(3, true, 400),
		NewOrder(2, true, 300),
		NewOrder(1, true, 200),
	}

	for _, o := range orders {
		ob.PlaceOrder(float64(o.Size)*o.Price, o)
	}

	ob.Bids.Each(func(key float64, val *Limit) {
		fmt.Println(key, val)
	})

	ob.BidsMap[1200].Orders.Each(func(key int64, val *Order) {
		fmt.Println(key, val)
	})

	assert.Equal(t, ob.Bids.Size(), 3)
}
