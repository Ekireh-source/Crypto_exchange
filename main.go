package main

import (
	"encoding/json"
	"exchange/orderbook"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	ex := NewExchange()

	e.GET("/book/:market", ex.handleGetBook)
	e.POST("/orders", ex.handlePlaceOrder)

	e.Start(":3000")

}

type OrderType string

const (
	LimitOrder  OrderType = "LIMIT"
	MarketOrder OrderType = "MARKET"
)

type Market string

const (
	MarketETH Market = "ETH"
)

type Exchange struct {
	orderbooks map[Market]*orderbook.OrderBook
}

func NewExchange() *Exchange {

	orderbooks := make(map[Market]*orderbook.OrderBook)

	orderbooks[MarketETH] = orderbook.NewOrderBook()

	return &Exchange{
		orderbooks: orderbooks,
	}
}

type PlaceOrderRequest struct {
	Type   OrderType
	Bid    bool
	Size   float64
	Price  float64
	Market Market
}

type Order struct {
	Price     float64
	Size      float64
	Bid       bool
	Timestamp int64
}

type OrderBookData struct {
	TotalBidVolume float64
	TotalAskVolume float64
	Bids           []*Order
	Asks           []*Order
}

func (ex *Exchange) handleGetBook(c echo.Context) error {
	market := c.Param("market")

	ob, ok := ex.orderbooks[Market(market)]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"message": "Market not found",
		})
	}

	orderBookData := OrderBookData{
		TotalBidVolume: ob.BidTotalVolume(),
		TotalAskVolume: ob.AskTotalVolume(),
		Bids:           []*Order{},
		Asks:           []*Order{},
	}

	for _, limit := range ob.Asks() {
		for _, order := range limit.Orders {
			o := Order{
				Price:     limit.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			}
			orderBookData.Asks = append(orderBookData.Asks, &o)
		}
	}

	for _, limit := range ob.Bids() {
		for _, order := range limit.Orders {
			o := Order{
				Price:     limit.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			}
			orderBookData.Bids = append(orderBookData.Bids, &o)
		}
	}

	return c.JSON(http.StatusOK, orderBookData)
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {
	var placeOrderData PlaceOrderRequest

	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return err
	}

	market := placeOrderData.Market

	ob := ex.orderbooks[market]

	order := orderbook.NewOrder(placeOrderData.Bid, placeOrderData.Size)

	if placeOrderData.Type == MarketOrder {
		fmt.Printf("Placing market order: %+v\n", placeOrderData)
		matches := ob.PlaceMarketOrder(order)
		return c.JSON(http.StatusOK, map[string]any{
			"matches": len(matches),
		})
	}

	if placeOrderData.Type == LimitOrder {
		ob.PlaceLimitOrder(placeOrderData.Price, order)
		return c.JSON(http.StatusOK, map[string]any{
			"message": "Limit order placed successfully",
		})
	}

	return nil
}
