package core

import (
	"crypto/ecdsa"

	"github.com/EggsyOnCode/velho-exchange/auth"
	"github.com/ethereum/go-ethereum/crypto"
)

type (
	Market string
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
	}

	orderbooks[BTC].SetExchange(ex)
	orderbooks[ETH].SetExchange(ex)

	return ex

}

func (ex *Exchange) AddUser(user *auth.User) {
	ex.Users[user.ID.String()] = user
}
