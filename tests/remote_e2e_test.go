package test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/shuoyang2016/mywish/rpc"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
	"github.com/golang/glog"
)

func init() {
	c, conn := setUpClient()
	defer conn.Close()
	_, err := (*c).TestingDropAll(context.Background(), &rpc.TestingDropRequest{})
	if err != nil {
		panic("Failed to clean up the database.")
	}
}

func setUpClient() (*rpc.MyWishServiceClient, *grpc.ClientConn) {
	addr := "192.168.29.108:8083"
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic("did not connect to db")
	}
	c := rpc.NewMyWishServiceClient(conn)
	_, err = c.TestingDropAll(context.Background(), &rpc.TestingDropRequest{})
	if err != nil {
		glog.Infof("%v", err)
		panic("Failed to clean up the database.")
	}
	return &c, conn
}

func TestBidFlow(t *testing.T) {
	c, conn := setUpClient()
	defer conn.Close()
	createProduct := rpc.CreateProductRequest{NewProduct: &rpc.Product{Id: 0, Name: "product0"}}
	_, err := (*c).CreateProduct(context.Background(), &createProduct)
	if err != nil {
		t.Errorf("Create product failed with %v", err)
	}
	res, err := (*c).GetProduct(context.Background(), &rpc.GetProductRequest{ProductId: 0, ProductName: "product0"})
	if err != nil {
		t.Errorf("Get product failed with %v", err)
		return
	}
	if res.Product.Id != 0 || res.Product.GetName() != "product0" {
		t.Error("Failed to create the product or get the product back")
	}
	_, err = (*c).CreateBidder(context.Background(),
		&rpc.CreateBidderRequest{Bidder: &rpc.Bidder{Id: 23, Name: "first_bidder"}})
	if err != nil {
		t.Errorf("Create bidder failed with %v", err)
		return
	}
	_, err = (*c).BidProduct(context.Background(), &rpc.BidProductRequest{Bid: &rpc.Bid{Id: 0, BidderId: 23, ProductId: 1,
		Price: &rpc.Price{Amount: 13.14}}})
	if err != nil {
		t.Errorf("Bid product failed with %v", err)
	}

	getBidderResponse, err := (*c).GetBidder(context.Background(), &rpc.GetBidderRequest{BidderId: 23})
	if getBidderResponse.GetBidder() == nil {
		t.Error("Unable to get bid")
	}
	if getBidderResponse.GetBidder().GetTotalAmountPending().GetAmount()+13.14 > 0.0001 {
		t.Error("Failed to update bidder amount")
	}
	if len(getBidderResponse.GetBidder().GetPendingBids()) != 1 || getBidderResponse.GetBidder().GetPendingBids()[0].GetProductId() != 1 {
		t.Error("Failed to update bidder bid entires.")
	}
	getProductResponse, err := (*c).GetProduct(context.Background(), &rpc.GetProductRequest{ProductId: 1})
	t.Log(getProductResponse)
	if len(getProductResponse.GetProduct().GetBidEntries()) != 1 ||
		getProductResponse.GetProduct().GetBidEntries()[0].BidderId != 23 {
		t.Error("Failed to update product bid entries.")
	}
}

func TestCreateFlow(t *testing.T) {
	c, conn := setUpClient()
	defer conn.Close()
	buyer := rpc.Bidder{Id: 123, TotalAmount: &rpc.Price{Amount: 100}, TotalAmountPending: &rpc.Price{Amount: 100}}
	bidder := rpc.Bidder{Id: 456, TotalAmount: &rpc.Price{Amount: 100}, TotalAmountPending: &rpc.Price{Amount: 100}}
	if _, err := (*c).CreateBidder(context.Background(), &rpc.CreateBidderRequest{Bidder: &buyer}); err != nil {
		t.Errorf("Failed to create bidder %v", buyer)
	}
	if _, err := (*c).CreateBidder(context.Background(), &rpc.CreateBidderRequest{Bidder: &bidder}); err != nil {
		t.Errorf("Failed to create bidder %v", bidder)
	}
	timestamp, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		t.Error("Failed to convert current time to timestamp proto.")
	}
	product := rpc.Product{Id: 1, Name: "1",
		Duration:      &duration.Duration{Seconds: 1800, Nanos: 1800 * 1000},
		BeginTime:     timestamp,
		BidStartPrice: &rpc.Price{Amount: 50},
		BuyerId:       123,
	}
	if _, err := (*c).CreateProduct(context.Background(), &rpc.CreateProductRequest{NewProduct: &product}); err != nil {
		t.Errorf("Error when create product at server: %v", err)
	}
	if resp, err := (*c).GetProduct(context.Background(), &rpc.GetProductRequest{ProductId: 1, ProductName: "1"}); err != nil {
		t.Errorf("Failed to get created product at server, %v", err)
	} else {
		if resp.Product.GetBuyerId() != 123 || resp.Product.GetBidStartPrice().GetAmount() != 50 {
			t.Error("Failed to create product at server.")
		}
	}
}

func TestBuyFlow(t *testing.T) {
	c, conn := setUpClient()
	defer conn.Close()
	buyer := rpc.Bidder{Id: 2, TotalAmount: &rpc.Price{Amount: 200}, TotalAmountPending: &rpc.Price{Amount: 200}}
	bidder := rpc.Bidder{Id: 3, TotalAmount: &rpc.Price{Amount: 200}, TotalAmountPending: &rpc.Price{Amount: 200}}
	if _, err := (*c).CreateBidder(context.Background(), &rpc.CreateBidderRequest{Bidder: &buyer}); err != nil {
		t.Errorf("Failed to create bidder %v", buyer)
	}
	if _, err := (*c).CreateBidder(context.Background(), &rpc.CreateBidderRequest{Bidder: &bidder}); err != nil {
		t.Errorf("Failed to create bidder %v", bidder)
	}
	timestamp, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		t.Error("Failed to convert current time to timestamp proto.")
	}
	product := rpc.Product{Id: 2, Name: "2",
		Duration:  &duration.Duration{Seconds: 1800, Nanos: 1800 * 1000},
		BeginTime: timestamp,
		DealPrice: &rpc.Price{Amount: 50},
		BuyerId:   2,
	}
	if _, err := (*c).CreateProduct(context.Background(), &rpc.CreateProductRequest{NewProduct: &product}); err != nil {
		t.Errorf("Error when create product at server: %v", err)
	}
	if resp, err := (*c).GetProduct(context.Background(), &rpc.GetProductRequest{ProductId: 2, ProductName: "2"}); err != nil {
		t.Errorf("Failed to get created product at server, %v", err)
	} else {
		if resp.Product.GetBuyerId() != 123 || resp.Product.GetBidStartPrice().GetAmount() != 50 {
			t.Error("Failed to create product at server.")
		}
	}
	if _, err := (*c).PayOff(context.Background(), &rpc.PayOffRequest{ProductId: 2, BidderId: 3, TakeDealPrice: true}); err != nil {
		t.Errorf("Error when pay off product at server: %v", err)
	}
	fProduct := &rpc.Product{}
	fBidder := &rpc.Bidder{}
	fBuyer := &rpc.Bidder{}
	resp, err := (*c).GetBidder(context.Background(), &rpc.GetBidderRequest{BidderId: 3})
	if err != nil {
		t.Errorf("Error when create product at server: %v", err)
	}
	fBidder = resp.GetBidder()
	resp, err = (*c).GetBidder(context.Background(), &rpc.GetBidderRequest{BidderId: 2})
	if err != nil {
		t.Errorf("Error when create product at server: %v", err)
	}
	fBuyer = resp.GetBidder()
	getProduct, err := (*c).GetProduct(context.Background(), &rpc.GetProductRequest{ProductId: 2})
	if err != nil {
		t.Errorf("Error when create product at server: %v", err)
	}
	fProduct = getProduct.GetProduct()
	if fBidder.GetInShippingProductIds()[0] != 2 {
		t.Error("Failed to update bidder's action to in shipping.")
	}
	if fBuyer.GetTotalAmountPending().GetAmount()-150 > 0.0001 {
		t.Error("Failed to charge buyer and return deposit.")
	}
	if fProduct.GetStatus() != rpc.Product_PENDING {
		t.Error("Failed to charge buyer and return deposit.")
	}
}

func TestCloseFlow(t *testing.T) {
	c, conn := setUpClient()
	defer conn.Close()
	buyer := rpc.Bidder{Id: 5, TotalAmount: &rpc.Price{Amount: 200}, TotalAmountPending: &rpc.Price{Amount: 200}}
	bidder := rpc.Bidder{Id: 6, TotalAmount: &rpc.Price{Amount: 200}, TotalAmountPending: &rpc.Price{Amount: 200}}
	if _, err := (*c).CreateBidder(context.Background(), &rpc.CreateBidderRequest{Bidder: &buyer}); err != nil {
		t.Errorf("Failed to create bidder %v", buyer)
	}
	if _, err := (*c).CreateBidder(context.Background(), &rpc.CreateBidderRequest{Bidder: &bidder}); err != nil {
		t.Errorf("Failed to create bidder %v", bidder)
	}
	timestamp, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		t.Error("Failed to convert current time to timestamp proto.")
	}
	product := rpc.Product{Id: 3, Name: "3",
		Duration:  &duration.Duration{Seconds: 1800, Nanos: 1800 * 1000},
		BeginTime: timestamp,
		DealPrice: &rpc.Price{Amount: 50},
		BuyerId:   5,
	}
	if _, err := (*c).CreateProduct(context.Background(), &rpc.CreateProductRequest{NewProduct: &product}); err != nil {
		t.Errorf("Error when create product at server: %v", err)
	}
	if resp, err := (*c).GetProduct(context.Background(), &rpc.GetProductRequest{ProductId: 3, ProductName: "3"}); err != nil {
		t.Errorf("Failed to get created product at server, %v", err)
	} else {
		if resp.Product.GetBuyerId() != 5 || resp.Product.GetBidStartPrice().GetAmount() != 50 {
			t.Error("Failed to create product at server.")
		}
	}
	if _, err := (*c).PayOff(context.Background(), &rpc.PayOffRequest{ProductId: 3, BidderId: 6, TakeDealPrice: true}); err != nil {
		t.Errorf("Error when pay off product at server: %v", err)
	}
	fProduct := &rpc.Product{}
	fBidder := &rpc.Bidder{}
	fBuyer := &rpc.Bidder{}
	resp, err := (*c).GetBidder(context.Background(), &rpc.GetBidderRequest{BidderId: 6})
	if err != nil {
		t.Errorf("Error when create product at server: %v", err)
	}
	fBidder = resp.GetBidder()
	resp, err = (*c).GetBidder(context.Background(), &rpc.GetBidderRequest{BidderId: 5})
	if err != nil {
		t.Errorf("Error when create product at server: %v", err)
	}
	fBuyer = resp.GetBidder()
	getProduct, err := (*c).GetProduct(context.Background(), &rpc.GetProductRequest{ProductId: 3})
	if err != nil {
		t.Errorf("Error when create product at server: %v", err)
	}
	fProduct = getProduct.GetProduct()
	if fBidder.GetClosedProductIds()[0] != 3 || fBidder.GetTotalAmountPending().GetAmount() != 250 ||
		fBidder.GetTotalAmount().GetAmount() != 250 || len(fBidder.GetPendingBids()) != 0 {
		t.Error("Failed to update bidder after closing the product.")
	}
	if fBuyer.GetTotalAmountPending().GetAmount()-150 > 0.0001 ||
		fBuyer.GetTotalAmountPending().GetAmount()-150 > 0.0001 {
		t.Error("Failed to charge buyer after closing the product.")
	}
	if fProduct.GetStatus() != rpc.Product_CLOSED {
		t.Error("Failed to upate product after closing the transaction.")
	}

}
