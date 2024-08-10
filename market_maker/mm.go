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
	// 2x of price offset
	MinSpread      int64
	SeedOffset     float64
	ExClient       *client.Client
	PriceOffset    float64
}

type MarketMaker struct {
	userID         string
	marketInterval time.Duration
	orderSize      int64
	minSpread      int64
	seedOffset     float64
	priceOffset    float64
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
		priceOffset:    cfg.PriceOffset,
	}
}

func (mm *MarketMaker) Start() {

	logrus.WithFields(logrus.Fields{
		"userId":         mm.userID,
		"orderSize":      mm.orderSize,
		"seedOffset":     mm.seedOffset,
		"minSpread":      mm.minSpread,
		"marketInterval": mm.marketInterval,
	}).Info("market maker starting ")

	go mm.makerLoop()
}

func (mm *MarketMaker) makerLoop() {

	ticker := time.NewTicker(mm.marketInterval)

	for {
		bestBid := mm.exClient.GetBestBidPrice("ETH")
		bestAsk := mm.exClient.GetBestAskPrice("ETH")

		if bestBid == 0 && bestAsk == 0 {
			mm.seedMarket()
			continue
		}

		spread := bestAsk - bestBid
		if spread <= float64(mm.minSpread) {
			continue
		}

		// market making strategy : Tightening the Spread
		mm.exClient.PlaceOrder("LIMIT", bestBid+mm.priceOffset, mm.orderSize, true, "ETH", mm.userID)
		mm.exClient.PlaceOrder("LIMIT", bestAsk-mm.priceOffset, mm.orderSize, false, "ETH", mm.userID)

		<-ticker.C
	}
}

func (mm *MarketMaker) seedMarket() {
	price := simulateFetchCurrentEthPrice()

	//bid
	mm.exClient.PlaceOrder("LIMIT", price-(mm.seedOffset), mm.orderSize, true, "ETH", mm.userID)

	// ask
	mm.exClient.PlaceOrder("LIMIT", price+(mm.seedOffset), mm.orderSize, false, "ETH", mm.userID)

}

// this function is used to simulate fetching the current ETH price
// from a 3rd party exchagne
func simulateFetchCurrentEthPrice() float64 {
	return 1000.0
}
