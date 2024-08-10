package client

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/EggsyOnCode/velho-exchange/api/handlers"
	"github.com/EggsyOnCode/velho-exchange/core"
	"github.com/sirupsen/logrus"
)

const (
	Endpoint = "http://localhost:3000"
)

type Client struct {
	client *http.Client
}

func NewClient() *Client {
	c := http.DefaultClient
	return &Client{
		client: c,
	}
}

func (c *Client) PlaceOrder(orderType string, price float64, size int64, bid bool, market string, user string) string {
	var t handlers.OrderType
	if orderType == "LIMIT" {
		t = handlers.LimitOrder
	} else {
		t = handlers.MarketOrder
	}
	order := &handlers.PlaceOrderRequest{
		OrderType: t,
		Price:     price,
		Size:      size,
		Bid:       bid,
		Market:    core.Market(market),
	}
	body, err := json.Marshal(order)
	if err != nil {
		log.Fatalf("client: error marshaling request body: %s\n", err)
	}

	endpoint := Endpoint + "/order"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))

	// Add the user ID as a query parameter
	q := req.URL.Query()
	q.Add("user", user)
	req.URL.RawQuery = q.Encode()

	if err != nil {
		log.Fatalf("client: error creating http request: %s\n", err)
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		log.Fatalf("client: error making http request: %s\n", err)
	}

	if res.StatusCode == http.StatusOK {

		// Decode response body
		var response map[string]interface{}
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("client: error reading response body: %s\n", err)
		}

		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			log.Fatalf("client: error unmarshaling response body: %s\n", err)
		}

		if orderType == "LIMIT" {
			if orderId, ok := response["id"].(string); ok {
				return orderId
			}
		} else {
			if matches, ok := response["matches"].([]interface{}); ok {
				logrus.WithFields(logrus.Fields{
					"matches": strconv.Itoa(len(matches)),
				}).Info("market order SUCCESSFUL")

				return strconv.Itoa(len(matches))
			}
		}

	} else {
		logrus.WithFields(logrus.Fields{
			"error": "not enough volume for market order ",
		}).Info("market order failed")
	}

	return ""
}

func (c *Client) RegisterUser(privKey string, usd float64) string {
	user := &handlers.User{
		PrivateKey: privKey,
		Usd:        usd,
	}

	body, err := json.Marshal(user)
	if err != nil {
		log.Fatalf("client: error marshaling request body: %s\n", err)
	}

	endpoint := Endpoint + "/user"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))

	if err != nil {
		log.Fatalf("client: error creating http request: %s\n", err)
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		log.Fatalf("client: error making http request: %s\n", err)
	}

	if res.StatusCode == http.StatusOK {
		// Decode response body
		var response map[string]interface{}
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("client: error reading response body: %s\n", err)
		}

		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			log.Fatalf("client: error unmarshaling response body: %s\n", err)
		}

		// Extract user ID from response
		if userID, ok := response["user"].(string); ok {
			return userID
		}

		log.Println("client: user ID not found in response")
	} else {
		log.Printf("client: failed to register user, status code: %d\n", res.StatusCode)
	}

	return ""
}

func (c *Client) GetBestAskPrice(market string) float64 {
	endpoint := Endpoint + "/book/ask?market=" + market
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		log.Fatalf("client: error creating http request: %s\n", err)
	}

	res, err := c.client.Do(req)
	if err != nil {
		log.Fatalf("client: error making http request: %s\n", err)
	}

	if res.StatusCode == http.StatusOK {
		// Decode response body
		var response map[string]float64
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("client: error reading response body: %s\n", err)
		}

		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			log.Fatalf("client: error unmarshaling response body: %s\n", err)
		}

		if price, ok := response["price"]; ok {
			log.Printf("client: best ask price for market %s: %f\n", market, price)
			return price
		}

		log.Println("client: best ask price not found in response")
	} else {
		log.Printf("client: failed to get best ask price, status code: %d\n", res.StatusCode)
	}

	return 0
}

func (c *Client) GetBestBidPrice(market string) float64 {
	endpoint := Endpoint + "/book/bid?market=" + market
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		log.Fatalf("client: error creating http request: %s\n", err)
	}

	res, err := c.client.Do(req)
	if err != nil {
		log.Fatalf("client: error making http request: %s\n", err)
	}

	if res.StatusCode == http.StatusOK {
		// Decode response body
		var response map[string]float64
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("client: error reading response body: %s\n", err)
		}

		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			log.Fatalf("client: error unmarshaling response body: %s\n", err)
		}

		if price, ok := response["price"]; ok {
			log.Printf("client: best ask price for market %s: %f\n", market, price)
			return price
		}

		log.Println("client: best ask price not found in response")
	} else {
		log.Printf("client: failed to get best ask price, status code: %d\n", res.StatusCode)
	}

	return 0
}

type Orders struct {
	Asks []*core.ExOrder `json:"Asks"`
	Bids []*core.ExOrder `json:"Bids"`
}

// Define the Response struct to match the entire JSON response
type Response struct {
	Orders Orders `json:"orders"`
	Status string `json:"status"`
}

func (c *Client) GetOrders(userId string) Response {
	endpoint := Endpoint + "/order?userID=" + userId
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		log.Fatalf("client: error creating http request: %s\n", err)
	}

	res, err := c.client.Do(req)
	if err != nil {
		log.Fatalf("client: error making http request: %s\n", err)
	}

	if res.StatusCode == http.StatusOK {
		// Decode response body
		var response Response
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("client: error reading response body: %s\n", err)
		}

		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			log.Fatalf("client: error unmarshaling response body: %s\n", err)
		}

		log.Printf("client: asks orders are %v and bids orders are %v\n", len(response.Orders.Asks), len(response.Orders.Bids))

		return response

	} else {
		log.Printf("client: failed to get orders, status code: %d\n", res.StatusCode)
	}

	return Response{}
}

func (c *Client) GetTrades(market string) []*core.Trade {
	endpoint := Endpoint + "/trade?market=" + market
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		log.Fatalf("client: error creating http request: %s\n", err)
	}

	res, err := c.client.Do(req)
	if err != nil {
		log.Fatalf("client: error making http request: %s\n", err)
	}

	if res.StatusCode == http.StatusOK {
		// Decode response body
		var response []*core.Trade
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("client: error reading response body: %s\n", err)
		}

		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			log.Fatalf("client: error unmarshaling response body: %s\n", err)
		}

		log.Printf("client: trades are %v\n", len(response))

		return response

	} else {
		log.Printf("client: failed to get trades, status code: %d\n", res.StatusCode)
	}

	return nil
}

func calculateAvgMarketOrderPrice(matches []interface{}) float64 {
	var total float64
	for _, m := range matches {
		match := m.(map[string]interface{})
		total += match["Price"].(float64) / match["SizeFilled"].(float64)
	}
	return total / float64(len(matches))
}
