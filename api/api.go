package api

import (
	"github.com/EggsyOnCode/velho-exchange/api/handlers"
	"github.com/EggsyOnCode/velho-exchange/core"
	"github.com/labstack/echo"
)

type Server struct {
	echo     *echo.Echo
	exchange *core.Exchange
}

func NewServer(exchange *core.Exchange) *Server {
	e := echo.New()
	server := &Server{
		echo:     e,
		exchange: exchange,
	}

	server.registerRoutes()
	return server
}

func (s *Server) registerRoutes() {
	s.echo.POST("/order", func(ctx echo.Context) error {
		return handlers.HandlePlaceOrder(ctx, s.exchange)
	})
	s.echo.GET("/orderbook", func(ctx echo.Context) error {
		return handlers.HandleGetOrderBook(ctx, s.exchange)
	})
	s.echo.DELETE("/order", func(ctx echo.Context) error {
		return handlers.HandleDeleteOrder(ctx, s.exchange)
	})
	s.echo.POST("/user", func(ctx echo.Context) error {
		return handlers.HandleUserRegistration(ctx, s.exchange)
	})

	s.echo.GET("/user/:id", func(ctx echo.Context) error {
		return handlers.HandleGetUser(ctx, s.exchange)
	})
}

func (s *Server) Start(addr string) {
	s.echo.Logger.Fatal(s.echo.Start(addr))
}
