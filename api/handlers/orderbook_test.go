package handlers

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EggsyOnCode/velho-exchange/auth"
	"github.com/EggsyOnCode/velho-exchange/core"
	"github.com/EggsyOnCode/velho-exchange/internals"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlePlaceOrderLimitOrder(t *testing.T) {
	e := core.NewExchange()
	user := auth.NewUser(nil, 100)
	e.AddUser(user)
	userId := user.ID.String()

	req := PlaceOrderRequest{
		OrderType: LimitOrder,
		Price:     10000.0,
		Size:      1,
		Bid:       false,
		Market:    core.BTC,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/orders?user="+userId, bytes.NewReader(toJson(req)))
	r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	ctx := echo.New().NewContext(r, w)

	err := HandlePlaceOrder(ctx, e)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotEmpty(t, response["id"])

	orders, _ := e.GetOrders(userId)
	assert.Len(t, orders, 1)
	assert.Equal(t, orders[0].Price, req.Price)

	ob := e.OrderBook[core.BTC]
	assert.Equal(t, ob.TotalAskVolume(), float64(req.Size))

}

func TestHandlePlaceOrderMarketOrderInsufficientVolume(t *testing.T) {
	e := core.NewExchange()
	user := auth.NewUser(nil, 100)
	e.AddUser(user)
	userId := user.ID.String()

	ob := e.OrderBook[core.BTC]
	ob.PlaceLimitOrder(10000.0, core.NewOrder(1, false, 10000.0, "otherUser")) // Add an ask

	req := PlaceOrderRequest{
		OrderType: MarketOrder,
		Size:      10,
		Bid:       true,
		Market:    core.BTC,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/orders?user="+userId, bytes.NewReader(toJson(req)))
	r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	ctx := echo.New().NewContext(r, w)

	err := HandlePlaceOrder(ctx, e)

	require.NoError(t, err)
	assert.Equal(t, http.StatusExpectationFailed, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "insufficient volume", response["error"])
}

func TestHandleGetOrderBook(t *testing.T) {
	e := core.NewExchange()

	// Create a user with sufficient USD (important!)
	pk := internals.GenerateNewPrivateKey()
	user := auth.NewUser(pk, 10000.0) // High initial balance
	e.AddUser(user)
	userId := user.ID.String()

	// Add orders to the order book, using the user ID
	ob := e.OrderBook[core.BTC]
	ob.PlaceLimitOrder(10000.0, core.NewOrder(1, true, 10000.0, userId))  // Bid order
	ob.PlaceLimitOrder(10001.0, core.NewOrder(2, false, 10001.0, userId)) // Ask order

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/orderbook?market=BTC", nil)
	ctx := echo.New().NewContext(r, w)

	err := HandleGetOrderBook(ctx, e)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)

	var response OrderBookResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	require.Len(t, response.Bids, 1)
	require.Len(t, response.Asks, 1)

	assert.Equal(t, float64(1), response.TotalBidVolume)
	assert.Equal(t, float64(2), response.TotalAskVolume)

	// Verify the order data in the response
	assert.Equal(t, float64(10000), response.Bids[0].Price)
	assert.Equal(t, float64(10001), response.Asks[0].Price)
	assert.Equal(t, int64(1), response.Bids[0].Size)
	assert.Equal(t, int64(2), response.Asks[0].Size)
	assert.Equal(t, userId, response.Bids[0].UserID)
	assert.Equal(t, userId, response.Asks[0].UserID)

}

func toJson(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func TestHandleUserRegistration(t *testing.T) {
	e := core.NewExchange()

	pk := internals.GenerateNewPrivateKey()
	privateKeyHex := internals.EncodeHexString(pk) // Correct conversion

	req := UserRegistrationRequest{
		PrivateKey: privateKeyHex, // Use the hex string
		Usd:        100.0,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(toJson(req)))
	r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON) // Important

	ctx := echo.New().NewContext(r, w)

	err := HandleUserRegistration(ctx, e)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response["user"])
}

// Ensure you have this function in your internals package
func EncodeHexString(pk *ecdsa.PrivateKey) string {
	return hex.EncodeToString(crypto.FromECDSA(pk))
}

// func TestHandleDeleteOrder(t *testing.T) {
// 	e := core.NewExchange()

// 	pk := internals.GenerateNewPrivateKey()
// 	user := auth.NewUser(pk, 100)
// 	e.AddUser(user)
// 	userId := user.ID.String()

// 	ob := e.OrderBook[core.BTC]
// 	order, err := ob.PlaceMarketOrder(10000.0, core.NewOrder(1, true, 10000.0, userId))
// 	require.NoError(t, err, "Failed to place order")

// 	orderId := order.ID.String()

// 	w := httptest.NewRecorder()
// 	r := httptest.NewRequest(http.MethodDelete, "/orders?id="+orderId+"&market=BTC", nil)
// 	ctx := echo.New().NewContext(r, w)

// 	err = HandleDeleteOrder(ctx, e)
// 	require.NoError(t, err, "HandleDeleteOrder returned an error")
// 	assert.Equal(t, http.StatusOK, w.Code)

// 	orders, _ := e.GetOrders(userId)
// 	assert.Empty(t, orders, "Orders should be empty after deletion")

// 	ob = e.OrderBook[core.BTC]
// 	limit := ob.Bids.GetLimit(10000)
// 	if limit != nil {
// 		assert.Empty(t, limit.Orders, "Order should be removed from orderbook")
// 	}
// }

func TestHandleDeleteOrder_InvalidID(t *testing.T) {
	e := core.NewExchange()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/orders?id=invalidID&market=BTC", nil)
	ctx := echo.New().NewContext(r, w)

	err := HandleDeleteOrder(ctx, e)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Invalid order ID", response["error"])
}

// func TestHandleDeleteOrder_InvalidMarket(t *testing.T) {
// 	e := core.NewExchange()

// 	w := httptest.NewRecorder()
// 	r := httptest.NewRequest(http.MethodDelete, "/orders?id=f47ac10b-58cc-4372-a567-0e02b2c3d479&market=INVALID", nil)
// 	ctx := echo.New().NewContext(r, w)

// 	err := HandleDeleteOrder(ctx, e)
// 	require.NoError(t, err)
// 	assert.Equal(t, http.StatusBadRequest, w.Code)

//		var response map[string]string
//		err = json.Unmarshal(w.Body.Bytes(), &response)
//		require.NoError(t, err)
//		assert.Equal(t, "Invalid market", response["error"])
//	}
func TestHandleGetUser(t *testing.T) {
	e := core.NewExchange()

	// Create a user
	pk := internals.GenerateNewPrivateKey()
	user := auth.NewUser(pk, 100.0)
	e.AddUser(user)
	userId := user.ID.String()

	// Test successful retrieval
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/users/:id", nil) // Using a placeholder in the request
	ctx := echo.New().NewContext(r, w)                          // Use echo.New().NewContext
	ctx.SetPath("/users/:id")                                   // Set the correct path
	ctx.SetParamNames("id")                                     // Set the parameter names
	ctx.SetParamValues(userId)                                  // Set the parameter values

	err := HandleGetUser(ctx, e)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleGetBestBidPrice(t *testing.T) {
	e := core.NewExchange()
	market := core.BTC

	ob := e.OrderBook[market]

	// Create users with USD
	pk1 := internals.GenerateNewPrivateKey()
	user1 := auth.NewUser(pk1, 10000)
	e.AddUser(user1)

	pk2 := internals.GenerateNewPrivateKey()
	user2 := auth.NewUser(pk2, 9500)
	e.AddUser(user2)

	// Place some orders for testing, using user IDs
	ob.PlaceLimitOrder(10000.0, core.NewOrder(1, true, 10000.0, user1.ID.String()))
	ob.PlaceLimitOrder(9500.0, core.NewOrder(2, true, 9500.0, user2.ID.String()))

	// Test successful retrieval
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/markets/BTC/bestBid", nil)
	ctx := echo.New().NewContext(r, w)
	ctx.SetParamNames("market")
	ctx.SetParamValues(string(market))

	err := HandleGetBestBidPrice(ctx, e)
	require.NoError(t, err)
	log.Println(w.Body.String())
	// assert.Equal(t, http.StatusOK, w.Code)

	// var priceResponse map[string]float64
	// err = json.Unmarshal(w.Body.Bytes(), &priceResponse)
	// require.NoError(t, err)
	// assert.Equal(t, 10000.0, priceResponse["price"])

	// Test invalid market
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/markets/INVALID/bestBid", nil)
	ctx = echo.New().NewContext(r, w)
	ctx.SetParamNames("market")
	ctx.SetParamValues("INVALID")

	err = HandleGetBestBidPrice(ctx, e)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResponse map[string]string // Use map[string]string for error response
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "Invalid market", errorResponse["error"])
}

func TestHandleGetOrders(t *testing.T) {
	// Setup
	e := core.NewExchange()

	// Simulate user creation and orders
	user := auth.NewUser(internals.GenerateNewPrivateKey(), 10000)
	userID := user.ID.String()
	e.AddUser(user)
	e.OrderBook["BTC"].PlaceLimitOrder(10000, core.NewOrder(1, true, 10000, userID))
	e.OrderBook["BTC"].PlaceLimitOrder(9500, core.NewOrder(2, false, 9500, userID))

	// Test case 1: Existing user with orders
	r := httptest.NewRequest(http.MethodGet, "/users/"+userID+"/orders", nil)
	w := httptest.NewRecorder()
	ctx := echo.New().NewContext(r, w)
	ctx.SetParamNames("userID")
	ctx.SetParamValues(userID)

	err := HandleGetOrders(ctx, e)
	require.NoError(t, err)
	// assert.Equal(t, http.StatusOK, w.Code)

	// ... (assertions for successful response with orders)

	// Test case 2: Non-existent user
	userID = "non-existent-user"
	r = httptest.NewRequest(http.MethodGet, "/users/"+userID+"/orders", nil)
	w = httptest.NewRecorder()
	ctx = echo.New().NewContext(r, w)
	ctx.SetParamNames("userID")
	ctx.SetParamValues(userID)

	err = HandleGetOrders(ctx, e)
	require.NoError(t, err)
	log.Printf("Response: %v", w.Code)
	assert.Equal(t, http.StatusNotFound, w.Code)

}
