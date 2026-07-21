package orderbook

import (
	"fmt"
	"reflect"
	"testing"
)

func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%+v != %+v", a, b)
	}
}

func TestLimit(t *testing.T) {
	l := NewLimit(10_000)
	buyOrderA := NewOrder(true, 5)
	buyOrderB := NewOrder(true, 2)
	buyOrderC := NewOrder(true, 3)

	l.AddOrder(buyOrderA)
	l.AddOrder(buyOrderB)
	l.AddOrder(buyOrderC)

	l.DeleteOrder(buyOrderA)

	fmt.Println(l)

}

func TestPlaceLimitOrder(t *testing.T) {
	ob := NewOrderBook()

	sellOrder := NewOrder(false, 5)
	sellOrderB := NewOrder(false, 2)

	ob.PlaceLimitOrder(10_000, sellOrder)
	ob.PlaceLimitOrder(9_000, sellOrderB)

	// assert(t, len(ob.asks), 1)

}

func TestPlaceMarketOrder(t *testing.T) {
	ob := NewOrderBook()

	sellOrder := NewOrder(false, 20)

	ob.PlaceLimitOrder(10_000, sellOrder)

	buyOrder := NewOrder(true, 10)

	matches := ob.PlaceMarketOrder(buyOrder)

	assert(t, len(matches), 1)
	assert(t, len(ob.asks), 1)
	assert(t, ob.AskTotalVolume(), 10.0)
	assert(t, matches[0].Ask, sellOrder)
	assert(t, matches[0].Bid, buyOrder)
	assert(t, matches[0].SizeFilled, 10.0)
	assert(t, matches[0].Price, 10_000.0)
	assert(t, buyOrder.IsFilled(), true)

}


func TestPlaceMarketOrderMultiFill(t *testing.T) {
	ob := NewOrderBook()
	
	buyOrderA := NewOrder(true, 5)
	buyOrderB := NewOrder(true, 8)
	buyOrderC := NewOrder(true, 10)
	buyOrderD := NewOrder(true, 1)

	ob.PlaceLimitOrder(9_000, buyOrderB)
	ob.PlaceLimitOrder(5_000, buyOrderC)
	ob.PlaceLimitOrder(10_000, buyOrderA)
	ob.PlaceLimitOrder(5_000, buyOrderD)
	

	assert(t, ob.BidTotalVolume(), 24.0)
	assert(t, ob.AskTotalVolume(), 0.0)

	sellOrder := NewOrder(false, 20)

	matches := ob.PlaceMarketOrder(sellOrder)

	assert(t, len(matches), 3)
	assert(t, ob.BidTotalVolume(), 4.0)
	assert(t, len(ob.bids), 1)
	


	fmt.Printf("%+v\n",matches)
	
}

func TestCancelOrder(t *testing.T) {
	ob := NewOrderBook()
	order := NewOrder(true, 4)
	ob.PlaceLimitOrder(10_000, order)


	ob.CancelOrder(order)
	assert(t, ob.BidTotalVolume(), 0.0)

	fmt.Printf("%+v\n",ob)
	fmt.Println(ob.bids)

}
