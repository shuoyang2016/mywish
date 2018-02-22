package test

import (
	"context"
	"testing"

	"github.com/shuoyang2016/mywish/rpc"
	"google.golang.org/grpc"
)

func TestBidFlow(t *testing.T) {
	addr := "192.168.29.108:8083"
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := rpc.NewMyWishServiceClient(conn)
	_, err = c.TestingDropAll(context.Background(), &rpc.TestingDropRequest{})
	if err != nil {
		t.Error("Failed to clean up the database.")
	}
	createProduct := rpc.CreateProductRequest{NewProduct: &rpc.Product{Id: 1, Name: "product1"}}
	_, err = c.CreateProduct(context.Background(), &createProduct)
	if err != nil {
		t.Errorf("Create product failed with %v", err)
	}
	res, err := c.GetProduct(context.Background(), &rpc.GetProductRequest{ProductId: 1, ProductName:"product1"})
	if err != nil {
		t.Errorf("Get product failed with %v", err)
		return
	}
	if res.Product.Id != 1 || res.Product.GetName() != "product1" {
		t.Error("Failed to create the product or get the product back")
	}
	_, err = c.CreateBidder(context.Background(), &rpc.CreateBidderRequest{BidderId: 23, BidderName:"first_bidder"})
	if err != nil {
		t.Errorf("Create bidder failed with %v", err)
		return
	}
	_, err = c.BidProduct(context.Background(), &rpc.BidProductRequest{Bid:&rpc.Bid{Id:0, BidderId: 23, ProductId:1, Price:13.14}})
	if err != nil {
		t.Errorf("Bid product failed with %v", err)
	}

	getBidderResponse, err := c.GetBidder(context.Background(), &rpc.GetBidderRequest{BidderId:23})
	if getBidderResponse.GetBidder() == nil {
		t.Error("Unable to get bid")
	}
	if getBidderResponse.GetBidder().GetTotalAmountPending().GetAmount() + 13.14 > 0.0001 {
		t.Error("Failed to update bidder amount")
	}
	if len(getBidderResponse.GetBidder().Bids) != 1 || getBidderResponse.GetBidder().Bids[0].ProductId != 1 {
		t.Error("Failed to update bidder bid entires.")
	}
	getProductResponse, err := c.GetProduct(context.Background(), &rpc.GetProductRequest{ProductId:1})
	t.Log(getProductResponse)
	if (len(getProductResponse.GetProduct().GetBidEntries()) != 1 ||
		getProductResponse.GetProduct().GetBidEntries()[0].BidderId != 23) {
		t.Error("Failed to update product bid entries.")
	}
}
