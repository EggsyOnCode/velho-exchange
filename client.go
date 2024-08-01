package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/EggsyOnCode/velho-exchange/api/handlers"
	"github.com/EggsyOnCode/velho-exchange/core"
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
				log.Printf("new LIMIT order : bid %v , price %v , size %v , market %v \n", bid, price, size, market)
				return orderId
			}
		} else {
			if matches, ok := response["matches"].([]interface{}); ok {
				avgPrice := calculateAvgMarketOrderPrice(matches)
				log.Printf("new Market order : bid %v , size %v , matches %v , avg Price %v \n", bid, size, len(matches), avgPrice)
				return strconv.Itoa(len(matches))
			}
		}

	} else {
		log.Printf("client: failed to place order, status code: %d\n", res.StatusCode)
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
			log.Printf("client: user registered successfully, ID: %s\n", userID)
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




func calculateAvgMarketOrderPrice(matches []interface{}) float64 {
	var total float64
	for _, m := range matches {
		match := m.(map[string]interface{})
		total += match["Price"].(float64) / match["SizeFilled"].(float64)
	}
	return total / float64(len(matches))
}


