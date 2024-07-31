package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/EggsyOnCode/velho-exchange/core"
	"github.com/labstack/echo"
)

type OrderType string

const (
	LimitOrder  OrderType = "LIMIT"
	MarketOrder OrderType = "MARKET"
)

type PlaceOrderRequest struct {
	OrderType OrderType   `json:"order_type"`
	Price     float64     `json:"price"`
	Size      int64       `json:"size"`
	Bid       bool        `json:"bid"`
	Market    core.Market `json:"market"`
}

type Order struct {
	ID        string
	Size      int64
	Timestamp int64
	Price     float64
	Bid       bool
	UserId    string
}

type OrderBookResponse struct {
	TotalAskVolume float64  `json:"total_ask_volume"`
	TotalBidVolume float64  `json:"total_bid_volume"`
	Asks           []*Order `json:"asks"`
	Bids           []*Order `json:"bids"`
}

func HandlePlaceOrder(ctx echo.Context, e *core.Exchange) error {
	var placeOrder PlaceOrderRequest
	userId := ctx.QueryParam("user")

	if err := json.NewDecoder(ctx.Request().Body).Decode(&placeOrder); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	ob := e.OrderBook[placeOrder.Market]
	order := core.NewOrder(placeOrder.Size, placeOrder.Bid, placeOrder.Price, userId)
	if placeOrder.OrderType == LimitOrder {
		ob.PlaceLimitOrder(placeOrder.Price, order)
		return ctx.JSON(http.StatusOK, map[string]string{"status": "success", "id": order.ID.String()})
	} else if placeOrder.OrderType == MarketOrder {
		matches := ob.PlaceMarketOrder(order)
		return ctx.JSON(http.StatusOK, map[string]any{"status": "success", "matches": matches})
	}

	return nil
}

func HandleGetOrderBook(ctx echo.Context, e *core.Exchange) error {
	market := ctx.QueryParam("market")
	ob := e.OrderBook[core.Market(market)]

	asks := make([]*Order, 0)
	bids := make([]*Order, 0)

	ob.Asks.Each(func(key float64, val *core.Limit) {
		val.Orders.Each(func(key int64, val *core.Order) {
			order := &Order{
				Size:      val.Size,
				Timestamp: val.Timestamp,
				Price:     val.Price,
				Bid:       val.Bid,
				ID:        val.ID.String(),
			}
			asks = append(asks, order)
		})
	})

	ob.Bids.Each(func(key float64, val *core.Limit) {
		val.Orders.Each(func(key int64, val *core.Order) {
			order := &Order{
				Size:      val.Size,
				Timestamp: val.Timestamp,
				Price:     val.Price,
				Bid:       val.Bid,
				ID:        val.ID.String(),
			}
			bids = append(bids, order)
		})
	})

	return ctx.JSON(http.StatusOK, OrderBookResponse{Asks: asks, Bids: bids, TotalAskVolume: ob.TotalAskVolume(), TotalBidVolume: ob.TotalBidVolume()})
}

func HandleDeleteOrder(ctx echo.Context, e *core.Exchange) error {
	id := ctx.QueryParam("id")
	market := ctx.QueryParam("market")
	ob := e.OrderBook[core.Market(market)]

	ob.CancelOrderById(id)
	return ctx.JSON(http.StatusOK, map[string]string{"status": "success"})
}
