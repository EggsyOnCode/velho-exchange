package api

import (
	"github.com/EggsyOnCode/velho-exchange/api/handlers"
	"github.com/EggsyOnCode/velho-exchange/core"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type Server struct {
	echo     *echo.Echo
	exchange *core.Exchange
}

func NewServer(exchange *core.Exchange) *Server {
	e := echo.New()

	// Configure CORS middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3001", "http://localhost:5173", "http://127.0.0.1:5173", "*"},
		AllowMethods:     []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowCredentials: true,
	}))
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

	s.echo.GET("/book/bid", func(ctx echo.Context) error {
		return handlers.HandleGetBestBidPrice(ctx, s.exchange)
	})

	s.echo.GET("/book/ask", func(ctx echo.Context) error {
		return handlers.HandleGetBestAskPrice(ctx, s.exchange)
	})

	s.echo.GET("/order", func(ctx echo.Context) error {
		return handlers.HandleGetOrders(ctx, s.exchange)
	})

	s.echo.GET("/trade", func(ctx echo.Context) error {
		return handlers.HandleGetTrades(ctx, s.exchange)
	})

	s.echo.GET("/marketPrice/:id", func(ctx echo.Context) error {
		return handlers.HandleGetMarketPrice(ctx, s.exchange)
	})
}

func (s *Server) Start(addr string) {
	s.echo.Logger.Fatal(s.echo.Start(addr))
}
