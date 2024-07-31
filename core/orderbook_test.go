package core

// import (
// 	"fmt"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func TestLimit(t *testing.T) {
// 	limit := NewLimit(10_000)
// 	order := NewOrder(2, true, 3_000)
// 	limit.AddOrder(order)

// 	limit.Orders.Each(func(key int64, val *Order) {
// 		fmt.Println(key, val)
// 	})

// 	fetchedOrder, _ := limit.Orders.Get(order.Timestamp)
// 	fmt.Printf("Fetched order: %v\n", fetchedOrder)
// 	assert.Equal(t, fetchedOrder, order)

// 	// removing orders
// 	limit.RemoveOrders([]*Order{order})
// 	assert.Equal(t, limit.TotalVolume, float64(0))

// 	fetchedOrder1, _ := limit.Orders.Get(order.Timestamp)
// 	assert.Nil(t, fetchedOrder1)

// 	// adding multiple orders
// 	orders := []*Order{
// 		NewOrder(1, true, 1_000),
// 		NewOrder(2, true, 2_000),
// 		NewOrder(3, true, 3_000),
// 	}

// 	for _, o := range orders {
// 		limit.AddOrder(o)
// 	}

// 	assert.Equal(t, limit.TotalVolume, float64(14_000))

// 	// removing 2nd Order
// 	limit.RemoveOrders([]*Order{orders[1]})

// 	assert.Equal(t, limit.TotalVolume, float64(10_000))
// }

// func TestOrderBook(t *testing.T) {
// 	ob := NewOrderBook()
// 	order := NewOrder(3, true, 400)
// 	ob.PlaceLimitOrder(1300, order)

// 	expectedOrder, _ := ob.BidsMap[1300].Orders.Get(order.Timestamp)
// 	assert.Equal(t, expectedOrder, order)
// 	assert.Equal(t, ob.BidsMap[1300].TotalVolume, float64(1200))
// }

// func TestSortingOfLimitsAndOrders(t *testing.T) {
// 	ob := NewOrderBook()
// 	orders := []*Order{
// 		NewOrder(3, false, 400),
// 		NewOrder(3, false, 400),
// 		NewOrder(2, false, 300),
// 		NewOrder(1, false, 200),
// 	}

// 	for _, o := range orders {
// 		ob.PlaceLimitOrder(float64(o.Size)*o.Price, o)
// 	}

// 	ob.Asks.Each(func(key float64, val *Limit) {
// 		fmt.Println(key, val)
// 	})

// 	ob.AsksMap[1200].Orders.Each(func(key int64, val *Order) {
// 		fmt.Println(key, val)
// 	})

// 	// assert.NotNil(t, nil)

// 	assert.Equal(t, ob.Asks.Size(), 3)
// }

// func TestPlaceBuyMarketOrder(t *testing.T) {
// 	ob := NewOrderBook()

// 	// selling 3 BTC
// 	sellOrder := NewOrder(3, false, 400)
// 	sellOrder1 := NewOrder(5, false, 300)
// 	ob.PlaceLimitOrder(sellOrder.TotalPrice(), sellOrder)
// 	ob.PlaceLimitOrder(sellOrder1.TotalPrice(), sellOrder1)

// 	// buying 5 BTC
// 	buyOrder := NewMarketOrder(5, true)
// 	matches := ob.PlaceMarketOrder(buyOrder)
// 	fmt.Printf("Matches: %v\n", matches)

// 	assert.Equal(t, len(matches), 2)
// 	assert.Equal(t, ob.Asks.Size(), 1)
// 	assert.Equal(t, ob.totalAskVolume, float64(900))
// 	assert.Equal(t, matches[0].Price, float64(1200))

// 	// assert.NotNil(t, nil)
// }

// func TestPlaceSellMarketOrder(t *testing.T) {
// 	ob := NewOrderBook()

// 	// buying 3 BTC
// 	buyOrder := NewOrder(3, true, 400)
// 	ob.PlaceLimitOrder(buyOrder.Price, buyOrder)

// 	buyOrder1 := NewOrder(3, true, 800)
// 	ob.PlaceLimitOrder(buyOrder1.Price, buyOrder1)

// 	buyOrder2 := NewOrder(1, true, 200)
// 	ob.PlaceLimitOrder(buyOrder2.Price, buyOrder2)

// 	// selling 5 BTC
// 	sellOrder := NewMarketOrder(5, false)
// 	matches := ob.PlaceMarketOrder(sellOrder)
// 	fmt.Printf("Matches: %v\n", matches)

// 	assert.Equal(t, len(matches), 2)
// 	assert.Equal(t, ob.Bids.Size(), 2)
// 	assert.Equal(t, ob.totalBidVolume, float64(600))
// 	// testing order of bids being matches
// 	assert.Equal(t, matches[0].Bid, buyOrder1)
// 	assert.Equal(t, matches[1].Bid, buyOrder)

// 	// assert.NotNil(t, nil)
// }

// func TestPlaceSellMarketOrderWithDuplicates(t *testing.T) {
// 	ob := NewOrderBook()

// 	// buying 3 BTC
// 	buyOrder := NewOrder(3, true, 400)
// 	ob.PlaceLimitOrder(buyOrder.Price, buyOrder)

// 	buyOrder3 := NewOrder(3, true, 400)
// 	ob.PlaceLimitOrder(buyOrder3.Price, buyOrder3)

// 	buyOrder1 := NewOrder(3, true, 800)
// 	ob.PlaceLimitOrder(buyOrder1.Price, buyOrder1)

// 	buyOrder2 := NewOrder(1, true, 200)
// 	ob.PlaceLimitOrder(buyOrder2.Price, buyOrder2)

// 	// selling 5 BTC
// 	sellOrder := NewMarketOrder(5, false)
// 	matches := ob.PlaceMarketOrder(sellOrder)
// 	fmt.Printf("Matches: %v\n", matches)

// 	assert.Equal(t, len(matches), 2)
// 	assert.Equal(t, ob.Bids.Size(), 2)
// 	assert.Equal(t, ob.totalBidVolume, float64(1800))
// 	// testing order of bids being matches
// 	assert.Equal(t, matches[0].Bid, buyOrder1)
// 	assert.Equal(t, matches[1].Bid, buyOrder3)

// 	// assert.NotNil(t, nil)
// }

// func TestCancelOrderFromOb(t *testing.T) {
// 	ob := NewOrderBook()
// 	o := NewOrder(3, true, 400)
// 	ob.PlaceLimitOrder(o.Price, o)
// 	assert.Equal(t, o, ob.GetOrderById(o.ID.String()))
// 	ob.CancelOrderById(o.ID.String())
// 	assert.Equal(t, o.Limit.Orders.Size(), 0)
// 	assert.Equal(t, ob.Bids.Size(), 0)
// 	assert.Nil(t, ob.GetOrderById(o.ID.String()))

// }
