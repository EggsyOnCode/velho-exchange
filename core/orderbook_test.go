package core

import (
	"testing"

	"github.com/EggsyOnCode/velho-exchange/auth"
	"github.com/EggsyOnCode/velho-exchange/internals"
)

// import (
// 	"fmt"
// 	"testing"

//	"github.com/EggsyOnCode/velho-exchange/auth"
//	"github.com/stretchr/testify/assert"
//
// )
func BenchmarkPlaceLimitOrder(b *testing.B) {
	ob := NewOrderBook(Market("ETH"))
	ex := NewExchange()
	ob.SetExchange(ex)
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		// Create a user for each iteration (important for accurate benchmarking)
		pk := internals.GenerateNewPrivateKey()
		user := auth.NewUser(pk, 10000)
		ex.AddUser(user)
		order := NewOrder(100, true, 1000.0, user.ID.String())
		ob.PlaceLimitOrder(1000.0, order)

		// Clean up the order and user for the next iteration
		ob.CancelOrder(order)
		delete(ex.Users, user.ID.String())
	}
}

func BenchmarkPlaceLimitOrderWithPreExistingUser(b *testing.B) {
	ob := NewOrderBook(Market("ETH"))
	ex := NewExchange()
	ob.SetExchange(ex)

	pk := internals.GenerateNewPrivateKey()
	user := auth.NewUser(pk, 10000)
	ex.AddUser(user)
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		order := NewOrder(100, true, 1000.0, user.ID.String())
		ob.PlaceLimitOrder(1000.0, order)

		// Clean up the order for the next iteration
		ob.CancelOrder(order)
	}
}

func FuzzGetBestAskPrice(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		var asks []float64
		for _, v := range data {
			asks = append(asks, float64(v))
		}

		ob := NewOrderBook(Market("ETH"))
		for _, price := range asks {
			ob.Asks.Put(price, NewLimit(price))
		}

		_ = ob.GetBestAskPrice()
	})
}

func TestPlaceMarketOrder_InsufficientAskVolume(t *testing.T) {
  // Create order book
  ob := NewOrderBook(Market("ETH"))

  // Create ask order
  askOrder := NewOrder(10, false, 1000, "user1")
  ob.PlaceLimitOrder(1000, askOrder)

  // Create market buy order for a larger size
  marketOrder := NewMarketOrder(20, true, "user2")

  // Place market order
  matches := ob.PlaceMarketOrder(marketOrder)

  // Assertions
  if len(matches) != 1 {
    t.Errorf("Expected 1 match, got %d", len(matches))
  }

  totalFilled := 0.0
  for _, m := range matches {
    totalFilled += m.SizeFilled
  }

  if totalFilled != 10 {
    t.Errorf("Expected order to be partially filled, filled %f", totalFilled)
  }
}