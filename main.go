package main

import (
	"fmt"
	"time"

	"github.com/EggsyOnCode/velho-exchange/api"
	"github.com/EggsyOnCode/velho-exchange/core"
)

func startServer() {
	exchange := core.NewExchange()
	server := api.NewServer(exchange)
	server.Start(":3000")
}

func main() {
	strings := []string{
		"7c852118294e51e653712a81e05800f419141751be58f605c371e15141b007a6",
		"59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d",
		"5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a",
	}

	go startServer()
	time.Sleep(1 * time.Second)

	client := NewClient()

	// Register users
	var userResponses []string
	for _, s := range strings {
		res := client.RegisterUser(s, 1000.0)
		userResponses = append(userResponses, res)
	}

	// Create two limit orders
	limitOrder1 := client.PlaceOrder("LIMIT", 10000.0, 50, true, "ETH", userResponses[0])
	limitOrder2 := client.PlaceOrder("LIMIT", 50000.0, 10, true, "ETH", userResponses[1])

	fmt.Printf("Limit Order 1 Response: %v\n", limitOrder1)
	fmt.Printf("Limit Order 2 Response: %v\n", limitOrder2)

	// Create a market order to buy a million ETH
	marketOrder := client.PlaceOrder("MARKET", 0, 50, false, "ETH", userResponses[2])
	fmt.Printf("Market Order Response: %v\n", marketOrder)

	select {}
}
