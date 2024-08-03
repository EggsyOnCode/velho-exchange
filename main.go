package main

import (
	"fmt"
	"time"

	"github.com/EggsyOnCode/velho-exchange/api"
	"github.com/EggsyOnCode/velho-exchange/core"
)

var tick = time.Second * 3
var ethPrice = 3000.0

func startServer() {
	exchange := core.NewExchange()
	server := api.NewServer(exchange)
	server.Start(":3000")
}

func main() {

	go startServer()
	time.Sleep(1 * time.Second)

	client := NewClient()

	go createMarketMaker(client)

	time.Sleep(2 * time.Second)

	go marketPlacer(client)

	select {}
}

func createMarketMaker(client *Client) {
	// Register a user
	ticker := time.NewTicker(tick)
	// mmUserID := client.RegisterUser("2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6", 1_000_000.0)
	// fmt.Println("--------------------------h")
	// fmt.Printf("Market Maker User ID: %s\n", mmUserID)
	// fmt.Println("--------------------------h")

	for {

		bestBid := client.GetBestBidPrice("ETH")
		bestAsk := client.GetBestAskPrice("ETH")

		fmt.Println("Best Bid: ", bestBid)
		fmt.Println("Best Ask: ", bestAsk)

		if (bestAsk == 0) && bestBid == 0 {
			seedMarketMaker(client)
		}

		<-ticker.C
	}

}

func seedMarketMaker(client *Client) []string {
	currentPrice := ethPrice // call to a an oracle / price exc
	priceOffset := 100.0

	strings := []string{
		"7c852118294e51e653712a81e05800f419141751be58f605c371e15141b007a6",
		"59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d",
		// "5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a",
	}

	// Register users
	var userResponses []string
	for _, s := range strings {
		res := client.RegisterUser(s, 1000.0)
		userResponses = append(userResponses, res)
	}

	// Create two limit orders
	client.PlaceOrder("LIMIT", currentPrice-priceOffset, 10, true, "ETH", userResponses[0])
	client.PlaceOrder("LIMIT", currentPrice+priceOffset, 9, false, "ETH", userResponses[1])
	// client.PlaceOrder("LIMIT", 4001, 70, true, "ETH", userResponses[2])
	// client.PlaceOrder("LIMIT", 4999.0, 50, false, "ETH", userResponses[0])

	return userResponses

}

func marketPlacer(client *Client) {
	ticker := time.NewTicker(5 * time.Second)
	userPk := "5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a"
	user := client.RegisterUser(userPk, 1000.0)

	for {

		client.PlaceOrder("MARKET", ethPrice-1000, 3, true, "ETH", user)
		client.PlaceOrder("MARKET", ethPrice+3, 4, false, "ETH", user)

		<-ticker.C
	}
}
