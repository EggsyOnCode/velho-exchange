package core

import (
	"fmt"
	"testing"

	"github.com/EggsyOnCode/velho-exchange/auth"
	"github.com/EggsyOnCode/velho-exchange/internals"
	"github.com/stretchr/testify/assert"
)

func TestExchange(t *testing.T) {
	ex := NewExchange()
	ob := ex.OrderBook[ETH]

	users := auth.GenerateUsers()

	for _, user := range users {
		user.USD = 100_000
	}

	ex.AddUser(users[0])
	ex.AddUser(users[1])
	ex.AddUser(users[2])
	fmt.Println("user id of buyer 0", users[0].ID.String())
	fmt.Println("user id of 1", users[1].ID.String())
	fmt.Println("user id of seller 2", users[2].ID.String())

	// buying ETH and selling USD
	buyOrder := NewOrder(3, true, 400, users[0].ID.String())
	ob.PlaceLimitOrder(buyOrder.Price, buyOrder)

	buyOrder1 := NewOrder(3, true, 800, users[1].ID.String())
	ob.PlaceLimitOrder(buyOrder1.Price, buyOrder1)

	assert.NotNil(t, 1)

	// selling ETH and buying USD
	sellOrder := NewMarketOrder(5, false, users[2].ID.String())
	matches := ob.PlaceMarketOrder(sellOrder)
	fmt.Printf("Matches: %v\n", matches)

	assert.Equal(t, len(matches), 2)
	assert.Equal(t, users[2].USD, float64(100_000+400*2+800*3))
	assert.Equal(t, users[0].USD, float64(100_000-400*3))
	assert.Equal(t, users[1].USD, float64(100_000-800*3))

	user0Bal := internals.GetBalance(internals.GetAddress(users[0].PrivateKey))
	assert.Equal(t, user0Bal, float64(10002))

	user1Bal := internals.GetBalance(internals.GetAddress(users[1].PrivateKey))

	user2Bal := internals.GetBalance(internals.GetAddress(users[2].PrivateKey))
	exBal := internals.GetBalance(internals.GetAddress(ex.PrivateKey))

	assert.Equal(t, user1Bal, float64(10003))

	gasPrice, _ := internals.GetGasPrice()

	tolerance := 0.0005

	assert.InEpsilon(t, 9995.0-gasPrice, user2Bal, tolerance, "User balance should match expected value within tolerance")
	assert.InEpsilon(t, 10000.0-gasPrice, exBal, tolerance, "Exchange balance should match expected value within tolerance")
}

func TestExchangeSellLimitBuyMarket(t *testing.T) {
	ex := NewExchange()
	ob := ex.OrderBook[ETH]

	users := auth.GenerateUsers()

	for _, user := range users {
		user.USD = 100_000 // Initial USD balance
	}

	ex.AddUser(users[0])
	ex.AddUser(users[1])
	ex.AddUser(users[2])
	fmt.Println("user id of seller 0", users[0].ID.String())
	fmt.Println("user id of seller 1", users[1].ID.String())
	fmt.Println("user id of buyer 2", users[2].ID.String())

	// selling ETH and buying USD
	sellOrder := NewOrder(3, false, 400, users[0].ID.String()) // Sell order from user 0
	ob.PlaceLimitOrder(sellOrder.Price, sellOrder)

	sellOrder1 := NewOrder(3, false, 800, users[1].ID.String()) // Sell order from user 1
	ob.PlaceLimitOrder(sellOrder1.Price, sellOrder1)

	// buying ETH and selling USD
	buyOrder := NewMarketOrder(5, true, users[2].ID.String()) // Buy market order from user 2
	matches := ob.PlaceMarketOrder(buyOrder)
	fmt.Printf("Matches: %v\n", matches)

	// Check that we have two matches
	assert.Equal(t, len(matches), 2)


	gasPrice, _ := internals.GetGasPrice()

	// Calculate expected balances after the transactions
	expectedUser0USD := 100_000 + 400*3           
	expectedUser0ETH := float64(10000 - 3)-gasPrice        
	expectedUser1USD := 100_000 + 800*2           
	// it will be 10000 - 3 because when submitting a limit order, the users's entire order size is tranferred to exchange (here 3)
	expectedUser1ETH := float64(10000 - 3)-gasPrice                         
	expectedUser2USD := 100_000 - (400*3 + 800*2) 

	assert.Equal(t, users[0].USD, float64(expectedUser0USD))
	assert.Equal(t, users[1].USD, float64(expectedUser1USD))
	assert.Equal(t, users[2].USD, float64(expectedUser2USD))

	// Assuming internals.GetBalance() retrieves the current balance of the given address
	user0Bal := internals.GetBalance(internals.GetAddress(users[0].PrivateKey))

	user1Bal := internals.GetBalance(internals.GetAddress(users[1].PrivateKey))
	user2Bal := internals.GetBalance(internals.GetAddress(users[2].PrivateKey))


	tolerance := 0.0005

	assert.InEpsilon(t, user0Bal, expectedUser0ETH, tolerance, "User balance should match expected value within tolerance")
	assert.InEpsilon(t, user1Bal, expectedUser1ETH, tolerance, "User balance should match expected value within tolerance")
	assert.InEpsilon(t, 10005-gasPrice, user2Bal, tolerance, "User balance should match expected value within tolerance")
}
