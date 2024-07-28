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

func (o *Order) String() string {
	t := time.Unix(o.Timestamp, 0)
	format := t.Format("2006-01-02 15:04:05")
	return fmt.Sprintf("Order{Size: %d, Timestamp: %v, Bid: %v}", o.Size, format, o.Bid)
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
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		Asks:    avl.New[float64, *Limit](g.Greater[float64]),
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

// This is for placing limit orders only
func (ob *OrderBook) PlaceOrder(price float64, o *Order) []Match {
	// find matches and fill as much as possible
	// TODO: Matching logic

	// add the order to a limit bucket ACT its price level
	if o.Size > 0.0 {
		ob.add(price, o)
	}

	// returns the matches
	return []Match{}
}

func (ob *OrderBook) add(price float64, o *Order) {
	var limit *Limit

	if o.Bid {
		if l, ok := ob.BidsMap[price]; ok {
			limit = l
			l.AddOrder(o)
		} else {
			limit = NewLimit(price)
			ob.BidsMap[price] = limit
			limit.AddOrder(o)
			ob.Bids.Put(price, limit)
		}
	} else {
		if l, ok := ob.AsksMap[price]; ok {
			limit = l
			l.AddOrder(o)
		} else {
			limit = NewLimit(price)
			ob.AsksMap[price] = limit
			limit.AddOrder(o)
			ob.Asks.Put(price, limit)
		}
	}

}
