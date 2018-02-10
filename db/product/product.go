package product

import "time"
import "github.com/shuoyang2016/mywish/db/user"

type Product struct {
	ID   int64
	Name string
	Buyer *user.Player
	Brand string
	DoubleConfirm bool
	AcceptRatio int
	SellerMinLevel uint
	ExtraDescription string
	BeginTime time.Time
	LastDuration time.Duration
}

type SupriseProduct struct {
	*Product
	usage *Usage
	function *Function
	Positivefilter *Filter
	NegFilter *Filter
}

type VagueProduct struct {
	*Product
	usage *Usage
	function *Function
}

type Usage struct {
	description string
}

type Function struct {
	description string
}

type Price struct {
	NoBidPrice float64
	FinalPrice float64
	BidBeginPrice float64
	BidHistory BidHistory
}

type BidHistory []*Bid

type Bid struct {
	Price float64
	Bider *user.Player
	Timesteamp time.Time
}