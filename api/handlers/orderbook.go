package handlers

import (
	"crypto/ecdsa"
	"encoding/json"
	"net/http"

	"github.com/EggsyOnCode/velho-exchange/auth"
	"github.com/EggsyOnCode/velho-exchange/core"
	"github.com/EggsyOnCode/velho-exchange/internals"
	"github.com/google/uuid"
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

type ExOrdersResponse struct {
	Asks []*core.ExOrder
	Bids []*core.ExOrder
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
		OrderType: core.OrderType(placeOrder.OrderType),
		Market:    placeOrder.Market,
	}

	e.AddOrder(o)

	if placeOrder.OrderType == LimitOrder {
		ob.PlaceLimitOrder(placeOrder.Price, order)
		return ctx.JSON(http.StatusOK, map[string]string{"status": "success", "id": order.ID.String()})
	} else if placeOrder.OrderType == MarketOrder {
		currentBidVol := ob.TotalBidVolume()
		currentAskVol := ob.TotalAskVolume()

		if o.Bid && float64(o.Size) > currentAskVol {
			return ctx.JSON(http.StatusExpectationFailed, map[string]string{"status": "false", "error": "insufficient volume"})
		} else if !o.Bid && float64(o.Size) > currentBidVol {
			return ctx.JSON(http.StatusExpectationFailed, map[string]string{"status": "false", "error": "insufficient volume"})
		}

		matches := ob.PlaceMarketOrder(order)
		if len(matches) == 0 {
			return ctx.JSON(http.StatusExpectationFailed, map[string]string{"status": "false", "matches": "no matches"})
		}
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
				Market:    core.Market(market),
				OrderType: core.LimitOrder,
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
				Market:    core.Market(market),
				OrderType: core.LimitOrder,
			}
			bids = append(bids, order)
		})
	})

	return ctx.JSON(http.StatusOK, OrderBookResponse{Asks: asks, Bids: bids, TotalAskVolume: ob.TotalAskVolume(), TotalBidVolume: ob.TotalBidVolume()})
}

func HandleDeleteOrder(ctx echo.Context, e *core.Exchange) error {
	idStr := ctx.QueryParam("id")
	market := ctx.QueryParam("market")
	ob := e.OrderBook[core.Market(market)]

	id, err := uuid.Parse(idStr) // Parse the ID here
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid order ID"})
	}

	ob.CancelOrderById(id.String())
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
	marketStr := ctx.QueryParam("market")
	market := core.Market(marketStr)

	ob, ok := e.OrderBook[market] // Check if the orderbook exists
	if !ok {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid market"})
	}

	if ob.Bids.Size() == 0 {
		return ctx.JSON(http.StatusOK, map[string]float64{"price": 0}) // Or return a specific "no bids" value
	}

	price := ob.GetBestBidPrice()
	return ctx.JSON(http.StatusOK, map[string]float64{"price": price})
}

func HandleGetBestAskPrice(ctx echo.Context, e *core.Exchange) error {
	marketStr := ctx.QueryParam("market")
	market := core.Market(marketStr)

	ob, ok := e.OrderBook[market] // Check if the orderbook exists
	if !ok {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid market"})
	}

	if ob.Asks.Size() == 0 {
		return ctx.JSON(http.StatusOK, map[string]float64{"price": 0}) // Or return a specific "no asks" value
	}

	price := ob.GetBestAskPrice()
	return ctx.JSON(http.StatusOK, map[string]float64{"price": price})
}

func HandleGetTrades(ctx echo.Context, e *core.Exchange) error {
	market := ctx.QueryParam("market")
	ob := e.OrderBook[core.Market(market)]
	trades := ob.GetTrades()

	return ctx.JSON(http.StatusOK, map[string]any{"status": "success", "trades": trades})
}

func HandleGetMarketPrice(ctx echo.Context, e *core.Exchange) error {
	market := ctx.QueryParam("market")
	ob := e.OrderBook[core.Market(market)]
	price := ob.GetMarketPrice()

	return ctx.JSON(http.StatusOK, map[string]any{"status": "success", "price": price})
}

func HandleGetOrders(ctx echo.Context, e *core.Exchange) error {
	id := ctx.QueryParam("userID")

	orders, exists := e.GetOrders(id)
	if !exists {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	if len(orders) == 0 {
		return ctx.JSON(http.StatusExpectationFailed, map[string]string{"error": "Orders not found; they are either filled or non-existant"})
	}

	var ordersRes ExOrdersResponse

	for _, order := range orders {
		if order.Bid {
			ordersRes.Asks = append(ordersRes.Asks, order)
		} else {
			ordersRes.Bids = append(ordersRes.Bids, order)
		}
	}

	return ctx.JSON(http.StatusOK, map[string]any{"status": "success", "orders": ordersRes})
}
