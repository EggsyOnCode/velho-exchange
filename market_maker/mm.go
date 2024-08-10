package mm

import (
	"time"

	"github.com/EggsyOnCode/velho-exchange/client"
	"github.com/sirupsen/logrus"
)

type Config struct {
	UserID         string
	MarketInterval time.Duration
	OrderSize      int64
	MinSpread      int64
	SeedOffset     float64
	ExClient       *client.Client
}

type MarketMaker struct {
	userID         string
	marketInterval time.Duration
	orderSize      int64
	minSpread      int64
	seedOffset     float64
	exClient       *client.Client
}

func NewMarketMaker(cfg Config) *MarketMaker {
	return &MarketMaker{
		userID:         cfg.UserID,
		marketInterval: cfg.MarketInterval,
		orderSize:      cfg.OrderSize,
		minSpread:      cfg.MinSpread,
		seedOffset:     cfg.SeedOffset,
		exClient:       cfg.ExClient,
	}
}

func (mm *MarketMaker) Start() {
	go mm.seedMarket()
}

func (mm *MarketMaker) seedMarket() {
	ticker := time.NewTicker(mm.marketInterval)
	price := simulateFetchCurrentEthPrice()

	for {
		logrus.WithFields(logrus.Fields{
			"price":          price,
			"userId":         mm.userID,
			"orderSize":      mm.orderSize,
			"seedOffset":     mm.seedOffset,
			"minSpread":      mm.minSpread,
			"marketInterval": mm.marketInterval,
		}).Info("market maker => seeding market ")
		//bid
		mm.exClient.PlaceOrder("LIMIT", price-(mm.seedOffset), mm.orderSize, true, "ETH", mm.userID)

		// ask
		mm.exClient.PlaceOrder("LIMIT", price+(mm.seedOffset), mm.orderSize, false, "ETH", mm.userID)

		<-ticker.C
	}
}

// this function is used to simulate fetching the current ETH price
// from a 3rd party exchagne
func simulateFetchCurrentEthPrice() float64 {
	return 1000.0
}
