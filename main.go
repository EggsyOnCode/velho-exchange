package main

import (
	"github.com/EggsyOnCode/velho-exchange/api"
)

func main() {
	server := api.NewServer()

	server.Start(":3000")
}
