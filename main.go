package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"exchange/orderbook"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/echo/v4"
)


const (
	exchangePrivatekey = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80" 
	LimitOrder  OrderType = "LIMIT"
	MarketOrder OrderType = "MARKET"
	MarketETH Market = "ETH"
)

type (
	OrderType string
	Market string

	Exchange struct {
		privateKey *ecdsa.PrivateKey
		orderbooks map[Market]*orderbook.OrderBook
	}

	PlaceOrderRequest struct {
		Type   OrderType
		Bid    bool
		Size   float64
		Price  float64
		Market Market
	}

	Order struct {
		ID        int64
		Price     float64
		Size      float64
		Bid       bool
		Timestamp int64
	}

	OrderBookData struct {
		TotalBidVolume float64
		TotalAskVolume float64
		Bids           []*Order
		Asks           []*Order
	}

	CancelOrderRequest struct {
		Bid bool
		ID int64
	}
)

func main() {
	e := echo.New()
	e.HTTPErrorHandler = httpErrorHandler

	ex, err := NewExchange(exchangePrivatekey)
	if err != nil {
		log.Fatal(err)
	}

	e.GET("/book/:market", ex.handleGetBook)
	e.POST("/orders", ex.handlePlaceOrder)
	e.DELETE("/orders/:id", ex.CancelOrder)


	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatal(err)
	}
	balance, err := client.BalanceAt(context.Background(), common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"), nil)
	if err != nil {
		log.Fatal(err)
	}


	privateKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	if err != nil {
	log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
	log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
	log.Fatal(err)
	}
	value := big.NewInt(1000000000000000000) // in wei (1 eth)

	gasLimit := uint64(21000) // in units

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
	log.Fatal(err)
	}

	toAddress := common.HexToAddress("0xa0Ee7A142d267C1f36714E4a8F75612F20a79720")

	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)

	chainID, err := client.ChainID(context.Background())
	fmt.Println("Chain ID", chainID)
	if err != nil {
	log.Fatal(err)
	}

	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), privateKey)
	if err != nil {
	log.Fatal(err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}


	fmt.Printf("tx sent: %s", signedTx.Hash().Hex())

	balance, err = client.BalanceAt(context.Background(), common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"), nil)
	if err != nil {
		log.Fatal(err)
	}
	

	fmt.Println("Balance:", balance)

	e.Start(":3000")

}


func httpErrorHandler(err error, c echo.Context) {
	if he, ok := err.(*echo.HTTPError); ok {
		c.JSON(he.Code, he.Message)
		return
	}
	c.JSON(http.StatusBadRequest, map[string]any{
		"error": err.Error(),
	})
}



type User struct {
	PrivateKey *ecdsa.PrivateKey
}


func NewUser(pk string) *User {
	pk, err := crypto.HexToECDSA(pk)
	if err != nil {
		log.Fatal(err)
	}
	return &User{
		PrivateKey: pk
	}
}





func NewExchange(privateKeyHex string) (*Exchange, error) {

	orderbooks := make(map[Market]*orderbook.OrderBook)

	orderbooks[MarketETH] = orderbook.NewOrderBook()

	pk, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, err
	}

	return &Exchange{
		privateKey: pk,
		orderbooks: orderbooks,
	}, nil
}




func (ex *Exchange) CancelOrder(c echo.Context) error {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	ob := ex.orderbooks[MarketETH]
	order := ob.Orders[int64(id)]
	ob.CancelOrder(order)

	log.Println("order canceled id => ", id)

	return c.JSON(200, map[string]any{"msg": "order deleted"})
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
				ID: 	   order.ID,
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
				ID:		   order.ID,
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
		
		matches, err := ob.PlaceMarketOrder(order)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]any{
				"error": err.Error(),
			})
		}
		return c.JSON(http.StatusOK, map[string]any{
			"matches": matches,
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
