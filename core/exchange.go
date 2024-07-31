package core

import (
	"crypto/ecdsa"

	"github.com/EggsyOnCode/velho-exchange/auth"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/zyedidia/generic/bimap"
)

type (
	Market string
)

const (
	BTC Market = "BTC"
	ETH Market = "ETH"
)

const DUMMY_PV = "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

type Exchange struct {
	PrivateKey *ecdsa.PrivateKey
	OrderBook  map[Market]*OrderBook
	Users      bimap.Bimap[string, *auth.User]
}

func NewExchange() *Exchange {
	orderbooks := make(map[Market]*OrderBook)
	orderbooks[BTC] = NewOrderBook()
	orderbooks[ETH] = NewOrderBook()

	pv, _ := crypto.HexToECDSA(DUMMY_PV)

	ex := &Exchange{
		PrivateKey: pv,
		OrderBook:  orderbooks,
	}

	orderbooks[BTC].SetExchange(ex)
	orderbooks[ETH].SetExchange(ex)

	return ex

}

func (ex *Exchange) AddUser(user *auth.User) {
	ex.Users.Add(user.ID.String(), user)
}
