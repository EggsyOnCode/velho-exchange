package main

import (
	"log"

	"github.com/EggsyOnCode/velho-exchange/api"
	"github.com/EggsyOnCode/velho-exchange/core"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatal("Failed to connect to the Ethereum client:", err)
	}
	_ = client

	exchange := core.NewExchange()
	server := api.NewServer(exchange)

	server.Start(":3000")
}
