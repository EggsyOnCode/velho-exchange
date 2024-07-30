package api

import (
	"github.com/EggsyOnCode/velho-exchange/api/handlers"
	"github.com/labstack/echo"
)

type Server struct {
	echo     *echo.Echo
	exchange *handlers.Exchange
}

func NewServer() *Server {
	e := echo.New()
	exchange := handlers.NewExchange()
	server := &Server{
		echo:     e,
		exchange: exchange,
	}

	server.registerRoutes()
	return server
}

func (s *Server) registerRoutes() {
	s.echo.POST("/order", s.exchange.HandlePlaceOrder)
	s.echo.GET("/orderbook", s.exchange.HandleGetOrderBook)
	s.echo.DELETE("/order", s.exchange.HandleDeleteOrder)
}

func (s *Server) Start(addr string) {
	s.echo.Logger.Fatal(s.echo.Start(addr))
}
