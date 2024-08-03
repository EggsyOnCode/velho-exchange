package core

import (
	"crypto/ecdsa"

	"github.com/EggsyOnCode/velho-exchange/auth"
	"github.com/ethereum/go-ethereum/crypto"
	g "github.com/zyedidia/generic"
	"github.com/zyedidia/generic/avl"
)

type (
	OrderType string
	Market    string
	ExOrder   struct {
		ID        string
		Size      int64
		Timestamp int64
		Price     float64
		Bid       bool
		UserID    string
		Market    Market
		OrderType OrderType
	}
)

const (
	BTC         Market    = "BTC"
	ETH         Market    = "ETH"
	LimitOrder  OrderType = "LIMIT"
	MarketOrder OrderType = "MARKET"
)

const DUMMY_PV = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

type Exchange struct {
	PrivateKey *ecdsa.PrivateKey
	OrderBook  map[Market]*OrderBook
	Users      map[string]*auth.User
	UsdPool    float64
	// stored against user ID
	orders map[string]*avl.Tree[string, *ExOrder]
}

func NewExchange() *Exchange {
	orderbooks := make(map[Market]*OrderBook)
	orderbooks[BTC] = NewOrderBook(BTC)
	orderbooks[ETH] = NewOrderBook(ETH)

	// priv

	pv, _ := crypto.HexToECDSA(DUMMY_PV)

	ex := &Exchange{
		PrivateKey: pv,
		OrderBook:  orderbooks,
		UsdPool:    0,
		Users:      make(map[string]*auth.User),
		orders:     make(map[string]*avl.Tree[string, *ExOrder]),
	}

	orderbooks[BTC].SetExchange(ex)
	orderbooks[ETH].SetExchange(ex)

	return ex

}

func (ex *Exchange) AddUser(user *auth.User) {
	ex.Users[user.ID.String()] = user
}

func (ex *Exchange) AddOrder(order *ExOrder) {
	if ex.orders[order.UserID] == nil {
		ex.orders[order.UserID] = avl.New[string, *ExOrder](g.Less[string])
		ex.orders[order.UserID].Put(order.ID, order)
		return
	}

	ex.orders[order.UserID].Put(order.ID, order)

}

func (ex *Exchange) GetOrders(userId string) ([]*ExOrder, bool) {
	var orders []*ExOrder
	_, exists := ex.orders[userId]
	if exists {
		ex.orders[userId].Each(func(k string, v *ExOrder) {
			ob := ex.OrderBook[v.Market]
			if ob.GetOrderById(v.ID) == nil {
				ex.orders[userId].Remove(k)
				return
			}
			orders = append(orders, v)
		})
	}

	return orders, exists
}
