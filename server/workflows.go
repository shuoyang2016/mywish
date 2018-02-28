package server

import (
	"errors"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/golang/glog"
	"github.com/shuoyang2016/mywish/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ glog.Level

const (
	buyerDepositPCT  = 0.2
	bidderDepositPCT = 0.1
)

/* Bid flow (done)
1. Create a bid object with bid ID, user ID, product ID.
2. If this is the first bid from the user, charge the deposit.
3. Update the product with new lowest price and Bid ID.
*/
func BidFlow(s *Server, req *rpc.BidProductRequest) error {
	bid := req.GetBid()
	bidderId := bid.GetBidderId()
	bidId := bid.GetId()
	productId := bid.GetProductId()
	price := bid.GetPrice().GetAmount()
	if price <= 0 {
		return errors.New("Price must be larger than 0.")
	}
	if bidderId == 0 || bidId == 0 || productId == 0 {
		return errors.New("The ids must exists.")
	}
	mongo := s.Mongo
	session := mongo.BaseSession.Clone()
	cb := session.DB(s.Mongo.DB).C(mongo.PlayerSCollection)
	cp := session.DB(s.Mongo.DB).C(mongo.ProductsCollection)
	oldProduct := rpc.Product{}
	if err := cp.Find(bson.M{"id": productId}).One(&oldProduct); err != nil {
		return err
	}
	if price <= oldProduct.GetDealPrice().GetAmount() {
		return errors.New("Bid price must be larger than deal price.")
	}
	if price >= oldProduct.GetBidStartPrice().GetAmount() {
		return errors.New("Bid price must be less than bid start price.")
	}
	if oldProduct.GetBestBid() != nil && price >= oldProduct.GetBestBid().GetPrice().GetAmount() {
		return errors.New("Bid price must be less than best bid price.")
	}
	// 1. Create a new bid in the bidder
	bidder := rpc.Bidder{}
	mywishBidder := rpc.Bidder{}
	if err := cb.Find(bson.M{"id": bid.GetBidderId()}).One(&bidder); err != nil {
		return err
	}
	if err := cb.Find(bson.M{"name": "mywish"}).One(&mywishBidder); err != nil {
		return err
	}
	if price > bidder.GetTotalAmountPending().GetAmount() {
		return errors.New("Bidder doesn't have enough money for the bid.")
	}

	// 2. If this is the first bid from the user, charge the deposit.
	var be *rpc.Bid
	for _, item := range bidder.GetPendingBids() {
		if item.GetProductId() == productId {
			be = item
			break
		}
	}
	if be != nil {
		currentBidPrice := be.GetPrice().GetAmount()
		bidder.TotalAmountPending.Amount += (currentBidPrice - price) * bidderDepositPCT
		mywishBidder.TotalAmountPending.Amount -= (currentBidPrice - price) * bidderDepositPCT
		be.Price = &rpc.Price{Amount: price}
	} else {
		be = &rpc.Bid{Id: 123, BidderId: bidderId, Price: &rpc.Price{Amount: price}}
		bidder.TotalAmountPending.Amount -= price * bidderDepositPCT
		mywishBidder.TotalAmountPending.Amount += price * bidderDepositPCT
		bidder.PendingBids = append(bidder.PendingBids, be)
	}
	if err := cb.Update(bson.M{"id": bid.BidderId}, &bidder); err != nil {
		return err
	}
	if err := cb.Update(bson.M{"name": "mywish"}, &mywishBidder); err != nil {
		return err
	}

	// 3. Update the product with new lowest price and Bid ID.
	if len(oldProduct.BidEntries) == 0 {
		oldProduct.Status = rpc.Product_BIDDING
	}
	oldProduct.BidEntries = append(oldProduct.BidEntries, be)
	oldProduct.BestBid = be
	if err := cp.Update(bson.M{"id": bid.GetProductId()}, oldProduct); err != nil {
		return err
	}
	return status.Error(codes.OK, " ")
}

/* Create product flow (done)
1. Create a product with product ID, user ID.
2. Add the product expiration time and add a call back to buy transaction at Scheduler.
3. Charge buyer's deposit
*/
func CreateProductFlow(s *Server, req *rpc.CreateProductRequest) error {
	cb := s.Mongo.BaseSession.DB(s.Mongo.DB).C(s.Mongo.PlayerSCollection)
	cp := s.Mongo.BaseSession.DB(s.Mongo.DB).C(s.Mongo.ProductsCollection)
	// 1. Create a product with the product info in request.
	product := req.GetNewProduct()
	if product.GetId() == 0 || product.GetName() == "" || product.GetDuration().GetNanos() <= 0 ||
		product.GetBeginTime == nil {
		return errors.New("Product has invalid fields.")
	}
	if product.GetBidStartPrice().GetAmount() == 0 && product.GetDealPrice().GetAmount() == 0 {
		return errors.New("Either start bid price or deal price need to be set")
	}
	basePriceForDeposit := product.GetBidStartPrice().GetAmount()
	if product.GetDealPrice().GetAmount() > basePriceForDeposit {
		basePriceForDeposit = product.GetDealPrice().GetAmount()
	}
	deposit := basePriceForDeposit * buyerDepositPCT
	product.Deposit = &rpc.Price{Amount: deposit}
	if err := cb.Insert(product); err != nil {
		return err
	}

	// 2. Add the product expiration time and add a call back to close transaction at Scheduler.
	t := NewTimer(time.Duration(product.Duration.Seconds), func(product_id int64, s *Server) func() {
		return func() {
			req := rpc.PayOffRequest{ProductId: product_id, TakeHighestBid: true}
			BuyProductFlow(s, &req)
		}
	}(product.GetId(), s))
	s.ProductIdToTimerMap[product.GetId()] = t

	// 3. Charge buyer's deposit
	buyer := rpc.Bidder{}
	if err := cp.Find(bson.M{"id": product.GetBuyerId()}).One(&buyer); err != nil {
		return err
	}
	buyer.ProductIdsInRequest = append(buyer.ProductIdsInRequest, product.GetId())
	if buyer.TotalAmountPending.GetAmount() < deposit {
		return errors.New("You need to have at least 20% for the deposit.")
	}
	buyer.TotalAmountPending.Amount -= deposit
	mywishBidder := rpc.Bidder{}
	if err := cp.Find(bson.M{"name": "mywish"}).One(&mywishBidder); err != nil {
		return err
	}
	mywishBidder.TotalAmountPending.Amount += deposit
	if err := cp.Update(bson.M{"id": product.GetBuyerId()}, &buyer); err != nil {
		return err
	}
	if err := cp.Update(bson.M{"name": "mywish"}, &buyer); err != nil {
		return err
	}
	return nil
}

/* Buy product flow (done)
1. Update the product to deal, final price and everything.
2. Return buyer's deposit and charge buyer based on final price.
3. Update the bidder for offline action.
*/
func BuyProductFlow(s *Server, req *rpc.PayOffRequest) error {
	cb := s.Mongo.BaseSession.DB(s.Mongo.DB).C(s.Mongo.PlayerSCollection)
	cp := s.Mongo.BaseSession.DB(s.Mongo.DB).C(s.Mongo.ProductsCollection)
	// 1. Update the product to pending, added final bidder id.
	product_id := req.GetProductId()
	product := rpc.Product{}
	if err := cp.Find(bson.M{"id": req.GetProductId()}).One(&product); err != nil {
		return err
	}
	product.Status = rpc.Product_PENDING
	product.FinalBidderId = req.GetBidderId()
	if product.FinalBidderId == 0 {
		product.Status = rpc.Product_CLOSED
	}
	// 2. Return buyer's deposit and charge buyer based on final price.
	deposit := product.GetDeposit().GetAmount()
	buyer := rpc.Bidder{}
	if err := cb.Find(bson.M{"id": product.GetBuyerId()}).One(&buyer); err != nil {
		return err
	}
	mywishBidder := rpc.Bidder{}
	if err := cb.Find(bson.M{"name": "mywish"}).One(&mywishBidder); err != nil {
		return err
	}
	buyer.TotalAmountPending.Amount += deposit
	mywishBidder.TotalAmountPending.Amount -= deposit
	if req.GetBidderId() == 0 {
		product.FinalPrice.Amount = 0
	} else if req.GetTakeDealPrice() {
		product.FinalPrice.Amount = product.GetDealPrice().GetAmount()
	} else if product.GetBestBid() == nil {
		product.FinalPrice.Amount = 0
	} else {
		product.FinalPrice.Amount = product.GetBestBid().GetPrice().GetAmount()
		product.FinalBidderId = product.GetBestBid().GetBidderId()
	}

	buyer.TotalAmountPending.Amount -= product.GetFinalPrice().GetAmount()
	mywishBidder.TotalAmountPending.Amount += product.GetFinalPrice().GetAmount()
	if err := cb.Update(bson.M{"id": product.GetBuyerId()}, &buyer); err != nil {
		return err
	}
	if err := cb.Update(bson.M{"name": "mywish"}, &mywishBidder); err != nil {
		return err
	}
	if err := cp.Update(bson.M{"id": product_id}, &product); err != nil {
		return err
	}
	if product.GetFinalBidderId() == 0 {
		return nil
	}
	// 3. Update the bidder for offline action.
	finalBidder := rpc.Bidder{}
	if err := cb.Find(bson.M{"id": product.GetFinalBidderId()}).One(&mywishBidder); err != nil {
		return err
	}
	finalBidder.InShippingProductIds = append(mywishBidder.InShippingProductIds, product.GetId())
	if err := cb.Update(bson.M{"id": finalBidder.GetId()}, &finalBidder); err != nil {
		return err
	}
	return nil
}

/* Close product flow （done)
1. Close product to status close
2. Transfer money to bidder's account
*/
func CloseProductFlow(s *Server, req *rpc.CloseProductRequest) error {
	cb := s.Mongo.BaseSession.DB(s.Mongo.DB).C(s.Mongo.PlayerSCollection)
	cp := s.Mongo.BaseSession.DB(s.Mongo.DB).C(s.Mongo.ProductsCollection)
	// 1. Close product to status close
	product := rpc.Product{}
	if err := cp.Find(bson.M{"id": req.GetProductId()}).One(&product); err != nil {
		return err
	}
	product.Status = rpc.Product_CLOSED
	if err := cp.Update(bson.M{"id": product.GetId()}, &product); err != nil {
		return err
	}
	if product.GetFinalBidderId() == 0 {
		return nil
	}

	// 2. Transfer money to bidder's account
	mywishBidder := rpc.Bidder{}
	if err := cb.Find(bson.M{"name": "mywish"}).One(&mywishBidder); err != nil {
		return err
	}
	finalBidder := rpc.Bidder{}
	if err := cb.Find(bson.M{"id": product.GetFinalBidderId()}).One(&mywishBidder); err != nil {
		return err
	}
	finalPrice := product.GetFinalPrice().GetAmount()
	finalBidder.TotalAmount.Amount += finalPrice
	for index, be := range finalBidder.GetPendingBids() {
		if be.GetProductId() == product.GetId() {
			finalBidder.TotalAmountPending.Amount += finalPrice * (1 + bidderDepositPCT) // Deposit + final price
			finalBidder.TotalAmount.Amount += finalPrice
			finalBidder.PendingBids = append(finalBidder.PendingBids[:index], finalBidder.PendingBids[index+1:]...)
			finalBidder.ClosedProductIds = append(finalBidder.ClosedProductIds, product.GetId())
			break
		}
	}
	finalBidder.TotalAmountPending.Amount -= finalPrice * (1 + bidderDepositPCT) // Transfer both deposit and sale price to bidder.
	if err := cb.Update(bson.M{"name": "mywish"}, &mywishBidder); err != nil {
		return err
	}
	if err := cb.Update(bson.M{"id": product.GetFinalBidderId()}, &finalBidder); err != nil {
		return err
	}
	return nil
}

func NewTimer(d time.Duration, fn func()) *time.Timer {
	t := time.NewTimer(d)
	go func() {
		<-t.C
		fn()
	}()
	return t
}
