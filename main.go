package main

import (
	"fmt"
	"time"

	"github.com/EggsyOnCode/velho-exchange/api"
	"github.com/EggsyOnCode/velho-exchange/core"
)

var tick = time.Second * 3

func startServer() {
	exchange := core.NewExchange()
	server := api.NewServer(exchange)
	server.Start(":3000")
}

func marketMakerPlacer(client *Client) {
	ticker := time.NewTicker(time.Second * 3)
	// Register a user
	user := client.RegisterUser("dbda1821b80551c9d65939329250298aa3472ba22feea921c0cf5d620ea67b97", 1_000.0)

	for {

		client.PlaceOrder("MARKET", 10000.0, 10, false, "ETH", user)
		client.PlaceOrder("MARKET", 8000.0, 6, true, "ETH", user)

		<-ticker.C
	}
}

func seedMarketMaker(client *Client) {

	strings := []string{
		"7c852118294e51e653712a81e05800f419141751be58f605c371e15141b007a6",
		"59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d",
		"5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a",
	}

	// Register users
	var userResponses []string
	for _, s := range strings {
		res := client.RegisterUser(s, 1000.0)
		userResponses = append(userResponses, res)
	}

	// Create two limit orders
	client.PlaceOrder("LIMIT", 10000.0, 50, true, "ETH", userResponses[0])
	client.PlaceOrder("LIMIT", 12000.0, 10, false, "ETH", userResponses[1])
	client.PlaceOrder("LIMIT", 8000.0, 7, true, "ETH", userResponses[2])
	client.PlaceOrder("LIMIT", 11500.0, 5, false, "ETH", userResponses[0])

	client.GetOrders(userResponses[0])

}

func createMarketMaker(client *Client) {
	// Register a user
	ticker := time.NewTicker(tick)
	// , _ := internals.GetPrivKeyFromHexString("2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6")
	mmUserID := client.RegisterUser("2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6", 1_000_000.0)
	straddle := 100.0

	for {

		bestBid := client.GetBestBidPrice("ETH")
		bestAsk := client.GetBestAskPrice("ETH")
		spread := bestAsk - bestBid

		fmt.Printf("Best bid: %f\n", bestBid)
		fmt.Printf("Best Ask: %f\n", bestAsk)
		fmt.Printf("Spread: %f\n", spread)

		// tightenting the spread
		client.PlaceOrder("LIMIT", bestBid+straddle, 10, true, "ETH", mmUserID)

		client.PlaceOrder("LIMIT", bestAsk-straddle, 10, false, "ETH", mmUserID)

		bestBid1 := client.GetBestBidPrice("ETH")
		bestAsk1 := client.GetBestAskPrice("ETH")
		spread1 := bestAsk1 - bestBid1

		fmt.Printf("Best bid: %f\n", bestBid1)
		fmt.Printf("Best Ask: %f\n", bestAsk1)
		fmt.Printf("Spread: %f\n", spread1)

		<-ticker.C
	}

}

func main() {

	go startServer()
	time.Sleep(1 * time.Second)

	client := NewClient()

	seedMarketMaker(client)



	// go createMarketMaker(client)

	// time.Sleep(1 * time.Second)

	// marketMakerPlacer(client)

	// // Create a market order to buy a million ETH
	// client.PlaceOrder("MARKET", 0, 50, false, "ETH", userResponses[2])

	select {}
}
