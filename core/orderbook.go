package core

import (
	"fmt"
	"time"

	"github.com/EggsyOnCode/velho-exchange/internals"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	g "github.com/zyedidia/generic"
	"github.com/zyedidia/generic/avl"
)

// these are to be used on the Front end for displaying recent trades
// every match order is a trade
// these trades are getting aggregated later on for analysis
type Trade struct {
	Price     float64
	Size      float64
	Bid       bool
	Timestamp int64
}

type Match struct {
	Ask        *Order
	Bid        *Order
	SizeFilled float64
	Price      float64
}

type Order struct {
	ID        uuid.UUID
	UserID    string
	Size      int64
	Timestamp int64
	Price     float64
	// if the order is for sell then its false, otherwise its true (for buy)
	Bid   bool
	Limit *Limit
}

func NewOrder(size int64, bid bool, price float64, userId string) *Order {
	return &Order{
		ID:        uuid.New(),
		Size:      size,
		Timestamp: time.Now().UnixNano(),
		Bid:       bid,
		Price:     price,
		UserID:    userId,
	}
}

func NewMarketOrder(size int64, bid bool, userID string) *Order {
	return &Order{
		Size:      size,
		Timestamp: time.Now().UnixNano(),
		Bid:       bid,
		UserID:    userID,
	}
}

func (o *Order) TotalPrice() float64 {
	return float64(o.Size * int64(o.Price))
}

func (o *Order) String() string {
	t := time.Unix(o.Timestamp, 0)
	format := t.Format("2006-01-02 15:04:05")
	return fmt.Sprintf("Order{ID: %s, UserID: %s, Size: %d, Timestamp: %s, Price: %f, Bid: %v}", o.ID, o.UserID, o.Size, format, o.Price, o.Bid)
}

func (o *Order) Type() string {
	if o.Bid {
		return "BID"
	}
	return "ASK"
}

func (o *Order) IsFilled() bool {
	return o.Size == 0
}

type Limit struct {
	Price float64
	// sorted by timestamps
	Orders *avl.Tree[int64, *Order]
	// total volume of tokens available for trade (not tokenAmt * Price)
	TotalVolume float64
}

func NewLimit(price float64) *Limit {
	return &Limit{
		Price:       price,
		Orders:      avl.New[int64, *Order](g.Greater[int64]),
		TotalVolume: 0,
	}
}

type OrderBook struct {
	// they'll be sorted by price levels
	Asks *avl.Tree[float64, *Limit]
	Bids *avl.Tree[float64, *Limit]

	// price levels to limit
	AsksMap map[float64]*Limit
	BidsMap map[float64]*Limit

	Trades *avl.Tree[int64, *Trade]

	OrdersMap map[uuid.UUID]*Order

	totalBidVolume float64
	totalAskVolume float64
	Exchange       *Exchange
	TokenId        Market
	CurrentPrice   float64
}

func NewOrderBook(tokenID Market) *OrderBook {
	return &OrderBook{
		// asks are sorted in ascending order : lowest ask first
		Asks: avl.New[float64, *Limit](g.Less[float64]),
		// bids are sorted in descending order : highest bid first
		Bids:         avl.New[float64, *Limit](g.Greater[float64]),
		AsksMap:      make(map[float64]*Limit),
		BidsMap:      make(map[float64]*Limit),
		OrdersMap:    make(map[uuid.UUID]*Order),
		Trades:       avl.New[int64, *Trade](g.Greater[int64]),
		TokenId:      tokenID,
		CurrentPrice: 0,
	}
}

func (ob *OrderBook) SetExchange(e *Exchange) {
	ob.Exchange = e
}

func (l *Limit) AddOrder(o *Order) {
	l.Orders.Put(o.Timestamp, o)
	l.TotalVolume += float64(o.Size)
}

// cancel / clear order
func (l *Limit) RemoveOrders(orders []*Order) bool {
	for _, o := range orders {
		l.Orders.Remove(o.Timestamp)
		l.TotalVolume -= float64(o.Size)
	}

	return l.Orders.Size() == 0
}

// we fill a bid / buyOrder
func (l *Limit) Fill(o *Order) ([]Match, []*Order, bool) {
	var (
		matches      []Match
		filledOrders []*Order
	)
	stop := false

	l.Orders.Each(func(key int64, order *Order) {
		if stop {
			return
		}

		match := l.fillOrder(o, order)
		matches = append(matches, match)

		l.TotalVolume -= match.SizeFilled

		if order.IsFilled() {
			filledOrders = append(filledOrders, order)
		}

		if o.IsFilled() {
			stop = true
		}
	})

	flag := l.RemoveOrders(filledOrders)
	if flag {
		return matches, filledOrders, true
	}

	// if the limit is empty, then we'll remove it from the orderbook
	return matches, filledOrders, false
}

func (l *Limit) fillOrder(o, order *Order) Match {
	var (
		sizeFilled float64
		ask        *Order
		bid        *Order
	)

	if o.Bid {
		// if the order is a bid, then the order is a bid and the limit order is an ask
		ask = order
		bid = o
	} else {
		// if the order is an ask, then the order is an ask and the limit order is a bid
		ask = o
		bid = order
	}

	if o.Size >= order.Size {
		o.Size -= order.Size
		sizeFilled = float64(order.Size)
		order.Size = 0
	} else {
		order.Size -= o.Size
		sizeFilled = float64(o.Size)
		o.Size = 0
	}

	updated_ask := *ask
	updated_bid := *bid

	return Match{
		Ask:        &updated_ask,
		Bid:        &updated_bid,
		SizeFilled: sizeFilled,
		Price:      order.Price * sizeFilled,
	}
}

func (ob *OrderBook) DeleteLimit(price float64, bid bool) {
	if bid {
		if l, ok := ob.BidsMap[price]; ok {
			// Delete the limit from the BidsMap
			delete(ob.BidsMap, price)
			// Remove the limit from the Bids tree
			ob.Bids.Remove(price)

			// Update the total bid volume
			ob.totalBidVolume -= l.TotalVolume
		}
	} else {
		if l, ok := ob.AsksMap[price]; ok {
			// Delete the limit from the AsksMap
			delete(ob.AsksMap, price)
			// Remove the limit from the Asks tree
			ob.Asks.Remove(price)

			// Update the total ask volume
			ob.totalAskVolume -= l.TotalVolume
		}
	}
}

func (ob *OrderBook) TotalAskVolume() float64 {
	return ob.totalAskVolume
}

func (ob *OrderBook) TotalBidVolume() float64 {
	return ob.totalBidVolume
}

// This is for placing limit orders only
// IMP : price level of an order could be different from o.size * o.price
func (ob *OrderBook) PlaceLimitOrder(price float64, o *Order) {
	var limit *Limit

	if o.Bid {
		if l, ok := ob.BidsMap[price]; ok {
			limit = l
			l.AddOrder(o)
			ob.OrdersMap[o.ID] = o
			o.Limit = limit
			ob.totalBidVolume += float64(o.Size)
			// tranferring usd to the exchange
			ob.TransferUSD(o.UserID, o.TotalPrice(), true)

		} else {
			limit = NewLimit(price)
			ob.BidsMap[price] = limit
			limit.AddOrder(o)
			ob.OrdersMap[o.ID] = o
			o.Limit = limit
			ob.Bids.Put(price, limit)
			ob.totalBidVolume += float64(o.Size)

			// transfering usd to the exchange
			ob.TransferUSD(o.UserID, o.TotalPrice(), true)

		}

	} else {
		if l, ok := ob.AsksMap[price]; ok {
			limit = l
			l.AddOrder(o)
			ob.OrdersMap[o.ID] = o
			o.Limit = limit
			ob.totalAskVolume += float64(o.Size)

			// transfer tokens to the exchange
			ob.TransferTokens(o.UserID, ob.TokenId, float64(o.Size), true)

		} else {
			limit = NewLimit(price)
			ob.AsksMap[price] = limit
			limit.AddOrder(o)
			ob.OrdersMap[o.ID] = o
			o.Limit = limit
			ob.Asks.Put(price, limit)
			ob.totalAskVolume += float64(o.Size)

			// transfer tokens to the exchange
			ob.TransferTokens(o.UserID, ob.TokenId, float64(o.Size), true)
		}
	}

	logrus.WithFields(
		logrus.Fields{
			"price":     price,
			"size":      o.Size,
			"type":      o.Type(),
			"userId":    o.UserID,
			"timestamp": o.Timestamp,
		},
	).Info("new limit Order")

}

func (ob *OrderBook) PlaceMarketOrder(o *Order) []Match {
	var matches []Match

	if o.Bid {
		// buying tokens in return for USD (for now)

		logrus.WithFields(
			logrus.Fields{
				"best ask price": ob.GetBestAskPrice(),
				"size":           o.Size,
				"type":           o.Type(),
				"userId":         o.UserID,
				"timestamp":      o.Timestamp,
			},
		).Info("new Market Order")

		if float64(o.Size) > ob.totalAskVolume {
			// market order can't be filled
			fmt.Errorf("market order can't be filled, not enough asks, current totalAskVolume: %f, order.TotalPrice: %f", ob.totalAskVolume, o.TotalPrice())
			return nil
		}

		stop := false
		ob.Asks.Each(func(key float64, l *Limit) {

			if stop {
				return
			}
			// we'll match the order with the asks ; incrementally starting from the lowest ask
			limitMatches, filledOrders, flag := l.Fill(o)
			matches = append(matches, limitMatches...)
			ob.deleteOrders(filledOrders)

			if flag {
				ob.DeleteLimit(key, false)
			}

			for _, m := range limitMatches {
				ob.totalAskVolume -= m.SizeFilled
			}

			if o.IsFilled() {
				stop = true
			}
		})

	} else {

		// user is selling tokens in return for USD from exchange

		logrus.WithFields(
			logrus.Fields{
				"best bid price": ob.GetBestBidPrice(),
				"size":           o.Size,
				"type":           o.Type(),
				"userId":         o.UserID,
				"timestamp":      o.Timestamp,
			},
		).Info("new Market Order")

		if float64(o.Size) > ob.totalBidVolume {
			// market order can't be filled
			fmt.Errorf("market order can't be consumed, not enough bids, current totalBidVolume: %f, order.TotalPrice: %f", ob.totalBidVolume, o.TotalPrice())
			return nil
		}

		ob.TransferTokens(o.UserID, ob.TokenId, float64(o.Size), true)

		stop := false
		ob.Bids.Each(func(key float64, l *Limit) {
			if stop {
				return
			}
			// we'll match the order with the bids ; incrementally starting from the highest ask
			limitMatches, filledOrders, flag := l.Fill(o)
			matches = append(matches, limitMatches...)
			ob.deleteOrders(filledOrders)

			if o.IsFilled() {
				stop = true
			}

			if flag {
				ob.DeleteLimit(key, true)
			}
			for _, m := range limitMatches {
				ob.totalBidVolume -= m.SizeFilled
			}

		})
	}

	for _, m := range matches {
		ob.Trades.Put(m.Ask.Timestamp, &Trade{
			Price:     m.Price,
			Size:      m.SizeFilled,
			Bid:       m.Bid.Bid,
			Timestamp: time.Now().UnixNano(),
		})
	}

	ob.BalanceOrderBookForMarketOrder(o, matches)

	//INFO: the current price of an asset is the price it was latest traded on (doesn't matter buy or sell)
	latestTrade, _ := ob.Trades.Get(matches[len(matches)-1].Ask.Timestamp)
	price := latestTrade.Price / latestTrade.Size

	logrus.WithFields(logrus.Fields{
		"currentPrice": price,
	}).Info("current price of the asset")

	logrus.WithFields(logrus.Fields{
		"matches":   len(matches),
		"avg Price": price,
	}).Info("market order filled")

	ob.CurrentPrice = price

	return matches
}

func (ob *OrderBook) CancelOrder(o *Order) {
	limit := o.Limit
	flag := limit.RemoveOrders([]*Order{o})
	if flag {
		ob.DeleteLimit(limit.Price, o.Bid)
	}
}

func (ob *OrderBook) deleteOrders(o []*Order) {
	for _, order := range o {
		delete(ob.OrdersMap, order.ID)
	}

}

func (ob *OrderBook) GetOrderById(id string) *Order {
	uuid := uuid.MustParse(id)
	o := ob.OrdersMap[uuid]

	return o
}

func (ob *OrderBook) CancelOrderById(orderId string) {
	orderID := uuid.MustParse(orderId)
	order, exists := ob.OrdersMap[orderID]
	if !exists {
		fmt.Printf("Order with ID %s not found\n", orderID)
		return
	}

	var limit *Limit
	if order.Bid {
		limit = ob.BidsMap[order.Price]
	} else {
		limit = ob.AsksMap[order.Price]
	}

	if limit != nil {

		// we only trasnfer tokens if the limit order is an ask
		// in case of a bid, the tokens are already in the user's custody
		// and usd are transferred p2p during matching
		if !order.Bid {
			// if the order is an ask, then even if it has already been matched
			// and has some tokens consumed, the remaining tokens will be left in teh CEX's custody
			// we will trasnfer those tokens
			ob.TransferTokens(order.UserID, ob.TokenId, float64(order.Size), false)
		}

		flag := limit.RemoveOrders([]*Order{order})
		if flag {
			ob.DeleteLimit(limit.Price, order.Bid)
		}
	}

	delete(ob.OrdersMap, orderID)
}

// userID: user who is transferring  the tokens or to whom the tokens are being transferred
func (ob *OrderBook) TransferTokens(userId string, token Market, tokenCount float64, toExchange bool) {
	// transfer tokens to/from the exchange
	switch token {
	case BTC:
		// Add BTC transfer logic here if needed
	case ETH:
		pvUser := ob.Exchange.Users[userId]
		exAddr := internals.GetAddress(ob.Exchange.PrivateKey)
		if toExchange {
			internals.TransferETH(pvUser.PrivateKey, exAddr, tokenCount)
		} else {
			pubKeyUser := internals.GetAddress(pvUser.PrivateKey)
			internals.TransferETH(ob.Exchange.PrivateKey, pubKeyUser, tokenCount)
		}
	}
}

func (ob *OrderBook) TransferUSDBetweenUsers(from, to string, usd float64) {
	fromUser := ob.Exchange.Users[from]
	toUser := ob.Exchange.Users[to]

	fromUser.USD -= usd
	toUser.USD += usd
}

func (ob *OrderBook) TransferUSD(userID string, usd float64, toExchange bool) {
	user := ob.Exchange.Users[userID]
	if toExchange {
		user.USD -= usd
		ob.Exchange.UsdPool += usd
	} else {
		user.USD += usd
		ob.Exchange.UsdPool -= usd
	}
}

// iterates over matches and transfers tokens or/and USD depending on the type of market order
func (ob *OrderBook) BalanceOrderBookForMarketOrder(o *Order, matches []Match) {
	for _, m := range matches {
		if o.Bid {
			// buying tokens in return for USD (for now)
			ob.TransferTokens(o.UserID, ob.TokenId, m.SizeFilled, false)
			// the user who placed the ask (who wanna sell their ETH for USD) will receive the USD
			// transferring USD from the buyer to the seller
			ob.TransferUSDBetweenUsers(o.UserID, m.Ask.UserID, m.Price)
		} else {
			// user is selling tokens in return for USD from exchange
			ob.TransferUSD(o.UserID, m.Price, false)
			// the user who placed the bid will receive the tokens
			userId := m.Bid.UserID

			// transfer tokens to the bid orders (who supplied ETH)
			ob.TransferTokens(userId, ob.TokenId, m.SizeFilled, false)
		}
	}
}

// returns highest bid price
func (ob *OrderBook) GetBestBidPrice() float64 {
	if ob.Bids.Size() == 0 {
		return 0
	}

	counter := 0
	var requiredLimit *Limit

	ob.Bids.Each(func(key float64, val *Limit) {
		if counter > 0 {
			return
		}
		requiredLimit = val
		counter++
	})

	return requiredLimit.Price

}

// returns lowest ask price
func (ob *OrderBook) GetBestAskPrice() float64 {
	if ob.Asks.Size() == 0 {
		return 0
	}

	counter := 0
	var requiredLimit *Limit

	ob.Asks.Each(func(key float64, val *Limit) {
		if counter > 0 {
			return
		}
		requiredLimit = val
		counter++
	})

	return requiredLimit.Price
}

func (ob *OrderBook) GetTrades() []*Trade {
	trades := make([]*Trade, 0)
	ob.Trades.Each(func(key int64, val *Trade) {
		trades = append(trades, val)
	})
	return trades
}

func CalculateAvgMarketOrderPrice(matches []Match) float64 {
	var total float64
	for _, m := range matches {
		total += m.Price / m.SizeFilled
	}
	return total / float64(len(matches))
}
