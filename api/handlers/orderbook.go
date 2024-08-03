package handlers

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/EggsyOnCode/velho-exchange/auth"
	"github.com/EggsyOnCode/velho-exchange/core"
	"github.com/EggsyOnCode/velho-exchange/internals"
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

type User struct {
	PrivateKey string  `json:"private_key"`
	Usd        float64 `json:"usd"`
}

type OrderBookResponse struct {
	TotalAskVolume float64         `json:"total_ask_volume"`
	TotalBidVolume float64         `json:"total_bid_volume"`
	Asks           []*core.ExOrder `json:"asks"`
	Bids           []*core.ExOrder `json:"bids"`
}

func HandlePlaceOrder(ctx echo.Context, e *core.Exchange) error {
	var placeOrder PlaceOrderRequest
	userId := ctx.QueryParam("user")

	if err := json.NewDecoder(ctx.Request().Body).Decode(&placeOrder); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	// add order to exchange
	ob := e.OrderBook[placeOrder.Market]
	order := core.NewOrder(placeOrder.Size, placeOrder.Bid, placeOrder.Price, userId)

	o := &core.ExOrder{
		Size:      order.Size,
		Price:     order.Price,
		ID:        order.ID.String(),
		UserID:    userId,
		Bid:       order.Bid,
		Timestamp: order.Timestamp,
	}

	e.AddOrder(o)

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

	asks := make([]*core.ExOrder, 0)
	bids := make([]*core.ExOrder, 0)

	ob.Asks.Each(func(key float64, val *core.Limit) {
		val.Orders.Each(func(key int64, val *core.Order) {
			order := &core.ExOrder{
				Size:      val.Size,
				Timestamp: val.Timestamp,
				Price:     val.Price,
				Bid:       val.Bid,
				ID:        val.ID.String(),
				UserID:    val.UserID,
			}
			asks = append(asks, order)
		})
	})

	ob.Bids.Each(func(key float64, val *core.Limit) {
		val.Orders.Each(func(key int64, val *core.Order) {
			order := &core.ExOrder{
				Size:      val.Size,
				Timestamp: val.Timestamp,
				Price:     val.Price,
				Bid:       val.Bid,
				ID:        val.ID.String(),
				UserID:    val.UserID,
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

type UserRegistrationRequest struct {
	PrivateKey string  `json:"private_key"`
	Usd        float64 `json:"usd"`
}

func HandleUserRegistration(ctx echo.Context, e *core.Exchange) error {
	var req UserRegistrationRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	balance := req.Usd

	// Note : DB will handle if teh private key is already registered
	var pk *ecdsa.PrivateKey
	var err error

	if req.PrivateKey == "" {
		pk = nil
	} else {
		pk, err = internals.GetPrivKeyFromHexString(req.PrivateKey)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid private key"})
		}
	}

	user := auth.NewUser(pk, balance)

	e.AddUser(user)

	fmt.Printf("Registering user with balance: %f\n", e.Users[user.ID.String()].USD)

	return ctx.JSON(http.StatusOK, map[string]any{"status": "success", "user": user.ID.String()})
}

func HandleGetUser(ctx echo.Context, e *core.Exchange) error {
	userPk := ctx.Param("id")

	if userPk == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user id"})
	}

	user := e.Users[userPk]

	return ctx.JSON(http.StatusOK, map[string]any{"status": "success", "user": user})
}

func HandleGetBestBidPrice(ctx echo.Context, e *core.Exchange) error {
	market := ctx.QueryParam("market")
	ob := e.OrderBook[core.Market(market)]
	price := ob.GetBestBidPrice()

	return ctx.JSON(http.StatusOK, map[string]float64{"price": price})
}

func HandleGetBestAskPrice(ctx echo.Context, e *core.Exchange) error {
	market := ctx.QueryParam("market")
	ob := e.OrderBook[core.Market(market)]
	price := ob.GetBestAskPrice()

	return ctx.JSON(http.StatusOK, map[string]float64{"price": price})
}

func HandleGetOrders(ctx echo.Context, e *core.Exchange) error {
	id := ctx.QueryParam("userID")

	orders, exists := e.GetOrders(id)
	if !exists {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	return ctx.JSON(http.StatusOK, map[string]any{"status": "success", "orders": orders})
}
