package main

import (
	"math/rand/v2"
	"time"

	"github.com/EggsyOnCode/velho-exchange/api"
	"github.com/EggsyOnCode/velho-exchange/client"
	"github.com/EggsyOnCode/velho-exchange/core"
	mm "github.com/EggsyOnCode/velho-exchange/market_maker"
)

const (
	ethPrice = 1000.0
)

func startServer() {
	exchange := core.NewExchange()
	server := api.NewServer(exchange)
	server.Start(":3000")
}

func initMMs(c *client.Client) []string {

	pvkeys := []string{
		"2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6",
		"4bbbf85ce3377467afe5d46f804f221813b2bb87f24d81f60f1fcdbf7cbf4356",
		"dbda1821b80551c9d65939329250298aa3472ba22feea921c0cf5d620ea67b97",
	}

	usd := 100_000_000.0
	users := make([]string, 0)
	for i := 0; i < len(pvkeys); i++ {
		userId := c.RegisterUser(pvkeys[i], usd)
		users = append(users, userId)
	}

	return users
}

func main() {

	go startServer()
	time.Sleep(1 * time.Second)
	client := client.NewClient()
	mmUsers := initMMs(client)

	time.Sleep(1 * time.Second)

	cfg := mm.Config{
		UserID:         mmUsers[0],
		OrderSize:      100,
		MarketInterval: 1 * time.Second,
		SeedOffset:     40,
		ExClient:       client,
		MinSpread:      20,
		PriceOffset:    10,
	}

	mm := mm.NewMarketMaker(cfg)
	go mm.Start()

	time.Sleep(2 * time.Second)

	go marketPlacer(client)

	select {}
}

func marketPlacer(client *client.Client) {
	ticker := time.NewTicker(3 * time.Second)
	userPk := "5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a"
	user := client.RegisterUser(userPk, 100_000.0)

	for {
		randInt := rand.IntN(10)
		var bid bool
		// more chances of placing a bid
		// 30 % chance of placing an ask
		// 70 % chance of placing a bid
		if randInt > 7 {
			bid = false
		} else {
			bid = true
		}
		// price doesn't matter in market orders
		client.PlaceOrder("MARKET", 0, 3, bid, "ETH", user)
		client.PlaceOrder("MARKET", 0, 4, bid, "ETH", user)

		<-ticker.C
	}
}
