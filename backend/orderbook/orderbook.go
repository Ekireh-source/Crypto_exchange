package orderbook

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)


type Match struct {
	Ask 	*Order
	Bid 	*Order
	Price 	float64
	SizeFilled 	float64
}


type Order struct {
	ID			int64
	Size 		float64
	Limit 		*Limit `json:"-"`
	Bid 		bool
	Timestamp 	int64
}

func (o *Order) IsFilled() bool {
	return o.Size == 0.0
}

type Orders []*Order

func (o Orders) Len() int           { return len(o) }
func (o Orders) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }
func (o Orders) Less(i, j int) bool { return o[i].Timestamp < o[j].Timestamp }


func NewOrder(bid bool, size float64) *Order {
	return &Order{
		ID: int64(rand.Intn(100000)),
		Size: size,
		Bid: bid,
		Timestamp: time.Now().UnixNano(),
	}
}

func (o *Order) String() string {
	return fmt.Sprintf("%v %v", o.Bid, o.Size)
}

type Limit struct {
	Price float64
	Orders Orders
	TotalVolume float64
}

type Limits []*Limit

type ByBestAsk struct{ Limits }

func (a ByBestAsk) Len() int           { return len(a.Limits) }
func (a ByBestAsk) Swap(i, j int)      { a.Limits[i], a.Limits[j] = a.Limits[j], a.Limits[i] }
func (a ByBestAsk) Less(i, j int) bool { return a.Limits[i].Price < a.Limits[j].Price }

type ByBestBid struct{ Limits }

func (b ByBestBid) Len() int           { return len(b.Limits) }
func (b ByBestBid) Swap(i, j int)      { b.Limits[i], b.Limits[j] = b.Limits[j], b.Limits[i] }
func (b ByBestBid) Less(i, j int) bool { return b.Limits[i].Price > b.Limits[j].Price }


func NewLimit(price float64) *Limit {
	return &Limit{
		Price: price,
		Orders: []*Order{},
		
	}
}

// func (l *Limit) String() string {
// 	return fmt.Sprintf("%v %v", l.Price, l.TotalVolume)
// }



func (l *Limit) AddOrder(order *Order) {
	order.Limit = l
	l.Orders = append(l.Orders, order)
	l.TotalVolume += order.Size
}

func (l *Limit) fill(order *Order) []Match {
	var (
		matches []Match
		ordersToDelete []*Order
	)

	for _, o := range l.Orders {
		match := l.fillOrder(o, order)
		matches = append(matches, match)

		l.TotalVolume -= match.SizeFilled

		if o.IsFilled() {
			ordersToDelete = append(ordersToDelete, o)
		}

		if order.IsFilled() {
			break
		}
	}

	for _, o := range ordersToDelete {
		l.DeleteOrder(o)
	}


	
	
	return matches
}


func (l *Limit) fillOrder(a,b *Order) Match {
	var (
		bid *Order
		ask *Order
		sizeFilled float64
	)

	if a.Bid{
		bid = a
		ask = b
	}else {
		bid = b
		ask = a
	}

	if a.Size >= b.Size {
		a.Size -= b.Size
		sizeFilled = b.Size
		b.Size = 0.0
		
	}else {

		b.Size -= a.Size
		sizeFilled = a.Size
		a.Size = 0.0
	}

	return Match{
		Bid: bid,
		Ask: ask,
		Price: l.Price,
		SizeFilled: sizeFilled,
	}
}

func (l *Limit) DeleteOrder(o *Order) {
	for i := 0; i < len(l.Orders); i++ {
		if l.Orders[i] == o {
			l.Orders[i] = l.Orders[len(l.Orders)-1]
			l.Orders = l.Orders[:len(l.Orders)-1]
		}
	}

	o.Limit = nil
	l.TotalVolume -= o.Size

	sort.Sort(l.Orders)
}



type OrderBook struct {
	asks []*Limit
	bids []*Limit


	AskLimits map[float64]*Limit
	BidLimits map[float64]*Limit
	Orders    map[int64]*Order
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		asks: []*Limit{},
		bids: []*Limit{},
		AskLimits: make(map[float64]*Limit),
		BidLimits: make(map[float64]*Limit),
		Orders:    make(map[int64]*Order),
	}
}

func (ob *OrderBook) PlaceMarketOrder (order *Order) ([]Match, error) {
	matches := []Match{}
	if order.Bid {
		if order.Size > ob.AskTotalVolume() {
			return nil, fmt.Errorf("not enough volume [size: %.2f] for market order [size: %.2f]", ob.AskTotalVolume(), order.Size)
		}
		for _, limit := range ob.Asks() {
			limitMatches := limit.fill(order)
			matches = append(matches, limitMatches...)

			if len(limit.Orders) == 0{
				ob.ClearLimit(false, limit)
			}
		}
	}else {
		if order.Size > ob.BidTotalVolume() {
			return nil, fmt.Errorf("not enough volume [size: %.2f] for market order [size: %.2f]", ob.BidTotalVolume(), order.Size)
		}
		for _, limit := range ob.Bids() {
			limitMatches := limit.fill(order)
			matches = append(matches, limitMatches...)

			if len(limit.Orders) == 0{
				ob.ClearLimit(true, limit)
			}
		}
	}
	return matches, nil
}

func (ob *OrderBook) PlaceLimitOrder (price float64, order *Order) {
	var limit *Limit

	if order.Bid {
		limit = ob.BidLimits[price]
	}else {
		limit = ob.AskLimits[price]
	}

	if limit == nil {
		limit = NewLimit(price)
		
		if order.Bid {
			ob.bids = append(ob.bids, limit)
			ob.BidLimits[price] = limit
		}else {
			ob.asks = append(ob.asks, limit)
			ob.AskLimits[price] = limit
		}
	}

	limit.AddOrder(order)
	ob.Orders[order.ID] = order
	
}


func (ob *OrderBook) ClearLimit(bid bool, l *Limit) {
	if bid {
		delete(ob.BidLimits, l.Price)
		for i := 0; i < len(ob.bids); i++ {
			if ob.bids[i] == l {
				ob.bids[i] = ob.bids[len(ob.bids)-1]
				ob.bids = ob.bids[:len(ob.bids)-1]
			}
		}
		
	}else {
		delete(ob.AskLimits, l.Price)
		for i := 0; i < len(ob.asks); i++ {
			if ob.asks[i] == l {
				ob.asks[i] = ob.asks[len(ob.asks)-1]
				ob.asks = ob.asks[:len(ob.asks)-1]
			}
		}
	}

	fmt.Printf("clearing limit price level [%.2f]\n", l.Price)
	
}

func (ob *OrderBook) CancelOrder(order *Order) {
	if order == nil {
		return
	}
	Limit := order.Limit
	if Limit != nil {
		Limit.DeleteOrder(order)
	}
	delete(ob.Orders, order.ID)
}


func (ob *OrderBook) BidTotalVolume() float64 {
	totalVolume := 0.0
	for i := 0; i < len(ob.bids); i++ {
		totalVolume += ob.bids[i].TotalVolume
	}
	return totalVolume
}


func (ob *OrderBook) AskTotalVolume() float64 {
	totalVolume := 0.0
	for i := 0; i < len(ob.asks); i++ {
		totalVolume += ob.asks[i].TotalVolume
	}
	return totalVolume
}


func (ob *OrderBook) Asks() []*Limit {
	sort.Sort(ByBestAsk{ob.asks})
	return ob.asks
}


func (ob *OrderBook) Bids() []*Limit {
	sort.Sort(ByBestBid{ob.bids})
	return ob.bids
}


