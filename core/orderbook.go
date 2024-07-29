package core

import (
	"fmt"
	"time"

	g "github.com/zyedidia/generic"
	"github.com/zyedidia/generic/avl"
)

type Match struct {
	Ask        *Order
	Bid        *Order
	SizeFilled float64
	Price      float64
}

type Order struct {
	Size      int64
	Timestamp int64
	Price     float64
	// if the order is for sell then its false, otherwise its true (for buy)
	Bid bool
}

func NewOrder(size int64, bid bool, price float64) *Order {
	return &Order{
		Size:      size,
		Timestamp: time.Now().UnixNano(),
		Bid:       bid,
		Price:     price,
	}
}

func NewMarketOrder(size int64, bid bool) *Order {
	return &Order{
		Size:      size,
		Timestamp: time.Now().UnixNano(),
		Bid:       bid,
	}
}

func (o *Order) TotalPrice() float64 {
	return float64(o.Size * int64(o.Price))
}

func (o *Order) String() string {
	t := time.Unix(o.Timestamp, 0)
	format := t.Format("2006-01-02 15:04:05")
	return fmt.Sprintf("Order{Size: %d, Timestamp: %v, Bid: %v}", o.Size, format, o.Bid)
}

func (o *Order) IsFilled() bool {
	return o.Size == 0
}

type Limit struct {
	Price float64
	// sorted by timestamps
	Orders      *avl.Tree[int64, *Order]
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

	totalBidVolume float64
	totalAskVolume float64
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		// asks are sorted in ascending order : lowest ask first
		Asks: avl.New[float64, *Limit](g.Less[float64]),
		// bids are sorted in descending order : highest bid first
		Bids:    avl.New[float64, *Limit](g.Greater[float64]),
		AsksMap: make(map[float64]*Limit),
		BidsMap: make(map[float64]*Limit),
	}
}

func (l *Limit) AddOrder(o *Order) {
	l.Orders.Put(o.Timestamp, o)
	l.TotalVolume += float64(o.Size * int64(o.Price))
}

func (l *Limit) RemoveOrders(orders []*Order) {
	for _, o := range orders {
		l.Orders.Remove(o.Timestamp)
		l.TotalVolume -= float64(o.Size * int64(o.Price))
	}
}

// we fill a bid / buyOrder
func (l *Limit) Fill(o *Order) ([]Match, bool) {
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

		l.TotalVolume -= match.Price

		if order.IsFilled() {
			filledOrders = append(filledOrders, order)
		}

		if o.IsFilled() {
			stop = true
		}
	})

	if len(filledOrders) == l.Orders.Size() {
		l.RemoveOrders(filledOrders)
		// true flag indicates that the limit itself needs to be deleted
		return matches, true
	}

	l.RemoveOrders(filledOrders)

	// false flag indicates that the limit itself doesn't need to be deleted
	return matches, false
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
			delete(ob.BidsMap, price)
			ob.Bids.Remove(price)
			ob.totalBidVolume -= l.TotalVolume
		}
	} else {
		if l, ok := ob.AsksMap[price]; ok {
			delete(ob.AsksMap, price)
			ob.Asks.Remove(price)
			ob.totalAskVolume -= l.TotalVolume
		}
	}
}

// This is for placing limit orders only
// IMP : price level of an order could be different from o.size * o.price
func (ob *OrderBook) PlaceLimitOrder(price float64, o *Order) {
	var limit *Limit

	if o.Bid {
		if l, ok := ob.BidsMap[price]; ok {
			limit = l
			l.AddOrder(o)
			ob.totalBidVolume += o.TotalPrice()
		} else {
			limit = NewLimit(price)
			ob.BidsMap[price] = limit
			limit.AddOrder(o)
			ob.Bids.Put(price, limit)
			ob.totalBidVolume += o.TotalPrice()
		}
	} else {
		if l, ok := ob.AsksMap[price]; ok {
			limit = l
			l.AddOrder(o)
			ob.totalAskVolume += o.TotalPrice()
		} else {
			limit = NewLimit(price)
			ob.AsksMap[price] = limit
			limit.AddOrder(o)
			ob.Asks.Put(price, limit)
			ob.totalAskVolume += o.TotalPrice()
		}
	}

}

func (ob *OrderBook) PlaceMarketOrder(price float64, o *Order) []Match {
	var matches []Match

	if o.Bid {
		if o.TotalPrice() > ob.totalAskVolume {
			// market order can't be filled
			panic(fmt.Errorf("market order can't be filled, not enough asks, current totalAskVolume: %f, order.TotalPrice: %f", ob.totalAskVolume, o.TotalPrice()))
		}

		stop := false
		ob.Asks.Each(func(key float64, l *Limit) {

			if stop {
				return
			}
			// we'll match the order with the asks ; incrementally starting from the lowest ask
			limitMatches, flag := l.Fill(o)
			matches = append(matches, limitMatches...)

			if flag {
				ob.DeleteLimit(key, false)
			}

			for _, m := range limitMatches {
				ob.totalAskVolume -= m.Price
			}

			if o.IsFilled() {
				stop = true
			}
		})
	} else {

		if o.TotalPrice() > ob.totalBidVolume {
			// market order can't be filled
			panic(fmt.Errorf("market order can't be consumed, not enough bids, current totalBidVolume: %f, order.TotalPrice: %f", ob.totalBidVolume, o.TotalPrice()))
		}

		stop := false
		ob.Bids.Each(func(key float64, l *Limit) {
			if stop {
				return
			}
			// we'll match the order with the bids ; incrementally starting from the highest ask
			limitMatches, flag := l.Fill(o)
			matches = append(matches, limitMatches...)

			if o.IsFilled() {
				stop = true
			}

			if flag {
				ob.DeleteLimit(key, true)
			}
			for _, m := range limitMatches {
				ob.totalBidVolume -= m.Price
			}

		})
	}

	return matches
}
