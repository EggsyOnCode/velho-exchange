package core

import (
	"crypto/ecdsa"

	"github.com/EggsyOnCode/velho-exchange/auth"
	"github.com/ethereum/go-ethereum/crypto"
)

type (
	Market  string
	ExOrder struct {
		ID        string
		Size      int64
		Timestamp int64
		Price     float64
		Bid       bool
		UserID    string
	}
)

const (
	BTC Market = "BTC"
	ETH Market = "ETH"
)

const DUMMY_PV = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

type Exchange struct {
	PrivateKey *ecdsa.PrivateKey
	OrderBook  map[Market]*OrderBook
	Users      map[string]*auth.User
	UsdPool    float64
	// stored against user ID
	orders map[string][]*ExOrder
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
		orders:     make(map[string][]*ExOrder, 0),
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
		ex.orders[order.UserID] = []*ExOrder{order}
	}
	ex.orders[order.UserID] = append(ex.orders[order.UserID], order)

}

func (ex *Exchange) GetOrders(userId string) ([]*ExOrder, bool) {
	orders, ok := ex.orders[userId]
	return orders, ok
}
