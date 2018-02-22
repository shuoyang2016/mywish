package server

import (
	"github.com/globalsign/mgo/bson"
	"github.com/shuoyang2016/mywish/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/golang/glog"
)

var _ glog.Level

/* Bid flow
1. Create a bid object with bid ID, user ID, product ID.
2. Update the product with new lowest price and Bid ID.
3. Bidder account has pending amount.
*/
func BidFlow(s *Server, req *rpc.BidProductRequest) error {

	// Create a new bid in the bidder
	bid := req.Bid
	mongo := s.Mongo
	session := mongo.BaseSession.Clone()
	c := session.DB(s.Mongo.DB).C(mongo.PlayerSCollection)
	bidder := rpc.Bidder{}
	if err := c.Find(bson.M{"id": bid.GetBidderId()}).One(&bidder); err != nil {
		return err
	}
	newBid := *bid
	bidder.Bids = append(bidder.Bids, &newBid)
	if bidder.GetTotalAmountPending() == nil {
		bidder.TotalAmountPending = &rpc.Price{}
	}

	// Update bidder's money
	newPrice := rpc.Price{Amount: bid.GetPrice()}
	bidder.PendingBidsAmount = append(bidder.PendingBidsAmount, &newPrice)
	bidder.GetTotalAmountPending().Amount -= bid.GetPrice()

	// Update bidder
	c.Update(bson.M{"id": bid.BidderId}, &bidder)

	// Update product based on new bids
	c = session.DB(mongo.DB).C(mongo.ProductsCollection)
	oldProduct := rpc.Product{}
	if err := c.Find(bson.M{"id": bid.GetProductId()}).One(&oldProduct); err != nil {
		return err
	}
	newEntry := rpc.BidEntry{BidderId: newBid.GetBidderId()}
	oldProduct.BidEntries = append(oldProduct.BidEntries, &newEntry)
	c.Update(bson.M{"id": bid.GetProductId()}, oldProduct)

	return status.Error(codes.OK, " ")
	// Create a bid object in bidder object
}

/* Create product flow
1. Create a product with product ID, user ID.
2. Added the production expiration time and add a call back to close transaction at Scheduler.
*/

/* Buy product flow
1. Update the product to deal.
2. Update the buyer for offline action, minus the amount to zhifubao.
*/

/* Close product flow
1. Bidder close product.
2. Buyer confirms
*/
