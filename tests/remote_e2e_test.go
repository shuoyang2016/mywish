package test

import (
	"context"
	"testing"
	"time"
	"math"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/shuoyang2016/mywish/rpc"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
)

func init() {
	c, conn := setUpClient()
	defer conn.Close()
	_, _ = (*c).TestingDropAll(context.Background(), &rpc.TestingDropRequest{})
}

func setUpClient() (*rpc.MyWishServiceClient, *grpc.ClientConn) {
	addr := "192.168.29.108:8083"
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic("did not connect to db")
	}
	c := rpc.NewMyWishServiceClient(conn)
	return &c, conn
}

func TestBidFlow(t *testing.T) {
	c, conn := setUpClient()
	defer conn.Close()
	mywishbidder := rpc.Bidder{Id: 102, Name: "mywish", TotalAmount: &rpc.Price{Amount: 0}, TotalAmountPending: &rpc.Price{Amount: 0}}
	buyer := rpc.Bidder{Id: 7, TotalAmount: &rpc.Price{Amount: 100}, TotalAmountPending: &rpc.Price{Amount: 100}}
	bidder := rpc.Bidder{Id: 23, TotalAmount: &rpc.Price{Amount: 100}, TotalAmountPending: &rpc.Price{Amount: 100}}
	if _, err := (*c).CreateBidder(context.Background(), &rpc.CreateBidderRequest{Bidder: &bidder}); err != nil {
		t.Errorf("Failed to create bidder %v", buyer)
	}
	if _, err := (*c).CreateBidder(context.Background(), &rpc.CreateBidderRequest{Bidder: &buyer}); err != nil {
		t.Errorf("Failed to create bidder %v", buyer)
	}
	if _, err := (*c).CreateBidder(context.Background(), &rpc.CreateBidderRequest{Bidder: &mywishbidder}); err != nil {
		t.Errorf("Failed to create bidder %v", mywishbidder)
	}
	timestamp, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		t.Error("Failed to convert current time to timestamp proto.")
	}
	product := rpc.Product{Id: 4, Name: "4",
		Duration:      &duration.Duration{Seconds: 1800, Nanos: 1800 * 1000},
		BeginTime:     timestamp,
		BidStartPrice: &rpc.Price{Amount: 50},
		BuyerId: 7,
	}
	createProduct := rpc.CreateProductRequest{NewProduct: &product}
	if _, err = (*c).CreateProduct(context.Background(), &createProduct); err != nil {
		t.Errorf("Create product failed with %v", err)
	}
	res, err := (*c).GetProduct(context.Background(), &rpc.GetProductRequest{ProductId: 4, ProductName: "4"})
	if err != nil {
		t.Errorf("Get product failed with %v", err)
	}
	if res.Product.Id != 4 || res.Product.GetName() != "4" {
		t.Error("Failed to create the product or get the product back")
	}
	_, err = (*c).CreateBidder(context.Background(),
		&rpc.CreateBidderRequest{Bidder: &rpc.Bidder{Id: 23, Name: "first_bidder", TotalAmountPending:&rpc.Price{Amount:100}, TotalAmount:&rpc.Price{Amount:100}}})
	if err != nil {
		t.Errorf("Create bidder failed with %v", err)
	}
	_, err = (*c).BidProduct(context.Background(), &rpc.BidProductRequest{Bid: &rpc.Bid{Id: 1, BidderId: 23, ProductId: 4,
		Price: &rpc.Price{Amount: 13.14}}})
	if err != nil {
		t.Errorf("Bid product failed with %v", err)
	}
	getBidderResponse, err := (*c).GetBidder(context.Background(), &rpc.GetBidderRequest{BidderId: 23})
	if getBidderResponse.GetBidder() == nil {
		t.Error("Unable to get bid")
	}
	if math.Abs(getBidderResponse.GetBidder().GetTotalAmountPending().GetAmount()+1.314 - 100) > 0.0001 {
		t.Error("Failed to update bidder amount")
	}
	if len(getBidderResponse.GetBidder().GetPendingBids()) != 1 || getBidderResponse.GetBidder().GetPendingBids()[0].GetProductId() != 4 {
		t.Error("Failed to update bidder bid entires.")
	}
	getProductResponse, err := (*c).GetProduct(context.Background(), &rpc.GetProductRequest{ProductId: 4})
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
	mywishbidder := rpc.Bidder{Id: 101, Name: "mywish", TotalAmount: &rpc.Price{Amount: 100}, TotalAmountPending: &rpc.Price{Amount: 100}}
	if _, err := (*c).CreateBidder(context.Background(), &rpc.CreateBidderRequest{Bidder: &buyer}); err != nil {
		t.Errorf("Failed to create bidder %v", buyer)
	}
	if _, err := (*c).CreateBidder(context.Background(), &rpc.CreateBidderRequest{Bidder: &bidder}); err != nil {
		t.Errorf("Failed to create bidder %v", bidder)
	}
	if _, err := (*c).CreateBidder(context.Background(), &rpc.CreateBidderRequest{Bidder: &mywishbidder}); err != nil {
		t.Errorf("Failed to create bidder %v", mywishbidder)
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

func TestCloseFlow(t *testing.T) {
	c, conn := setUpClient()
	defer conn.Close()
	mywish_bidder := rpc.Bidder{Id: 100, Name:"mywish", TotalAmount: &rpc.Price{Amount: 200}, TotalAmountPending: &rpc.Price{Amount: 200}}
	buyer := rpc.Bidder{Id: 10, Name:"10", TotalAmount: &rpc.Price{Amount: 200}, TotalAmountPending: &rpc.Price{Amount: 200}}
	bidder := rpc.Bidder{Id: 11, Name:"11", TotalAmount: &rpc.Price{Amount: 200}, TotalAmountPending: &rpc.Price{Amount: 200}}
	if _, err := (*c).CreateBidder(context.Background(), &rpc.CreateBidderRequest{Bidder: &buyer}); err != nil {
		t.Errorf("Failed to create bidder %v", buyer)
	}
	if _, err := (*c).CreateBidder(context.Background(), &rpc.CreateBidderRequest{Bidder: &bidder}); err != nil {
		t.Errorf("Failed to create bidder %v", bidder)
	}
	if _, err := (*c).CreateBidder(context.Background(), &rpc.CreateBidderRequest{Bidder: &mywish_bidder}); err != nil {
		t.Errorf("Failed to create bidder %v", mywish_bidder)
	}
	timestamp, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		t.Error("Failed to convert current time to timestamp proto.")
	}
	product := rpc.Product{Id: 7, Name: "7",
		Duration:  &duration.Duration{Seconds: 1800, Nanos: 1800 * 1000},
		BeginTime: timestamp,
		DealPrice: &rpc.Price{Amount: 50},
		FinalPrice: &rpc.Price{Amount: 50},
		BuyerId:   10,
		FinalBidderId:11,
	}
	if _, err := (*c).CreateProduct(context.Background(), &rpc.CreateProductRequest{NewProduct: &product}); err != nil {
		t.Errorf("Error when create product at server: %v", err)
	}
	if resp, err := (*c).GetProduct(context.Background(), &rpc.GetProductRequest{ProductId: 7, ProductName: "7"}); err != nil {
		t.Errorf("Failed to get created product at server, %v", err)
	} else {
		if resp.Product.GetBuyerId() != 10 || resp.Product.GetDealPrice().GetAmount() != 50 {
			t.Error("Failed to create product at server.")
		}
	}
	if _, err = (*c).CloseProduct(context.Background(), &rpc.CloseProductRequest{ProductId:7, RequesterId:11}); err != nil {
		t.Errorf("Failed to close product %v", err)
	}
	if resp, err := (*c).GetProduct(context.Background(), &rpc.GetProductRequest{ProductId: 7, ProductName: "7"}); err != nil {
		t.Errorf("Failed to get created product at server, %v", err)
	} else {
		if resp.Product.Status != rpc.Product_CLOSED || resp.Product.GetDealPrice().GetAmount() != 50 {
			t.Error("Failed to create product at server.")
		}
	}
	if resp, err := (*c).GetBidder(context.Background(), &rpc.GetBidderRequest{BidderId: 11}); err != nil {
		t.Errorf("Failed to get bidder %v", err)
	} else {
		if math.Abs(resp.Bidder.GetTotalAmount().GetAmount() - 200 - 50) > 0.001 {
			t.Errorf("Failed to transfer money to bidder.")
		}
	}
}

func TestBuyFlow(t *testing.T) {
	c, conn := setUpClient()
	defer conn.Close()
	mywish_bidder := rpc.Bidder{Id: 99, Name:"mywish", TotalAmount: &rpc.Price{Amount: 200}, TotalAmountPending: &rpc.Price{Amount: 200}}
	buyer := rpc.Bidder{Id: 5, Name:"5", TotalAmount: &rpc.Price{Amount: 200}, TotalAmountPending: &rpc.Price{Amount: 200}}
	bidder := rpc.Bidder{Id: 6, Name:"6", TotalAmount: &rpc.Price{Amount: 200}, TotalAmountPending: &rpc.Price{Amount: 200}}
	if _, err := (*c).CreateBidder(context.Background(), &rpc.CreateBidderRequest{Bidder: &buyer}); err != nil {
		t.Errorf("Failed to create bidder %v", buyer)
	}
	if _, err := (*c).CreateBidder(context.Background(), &rpc.CreateBidderRequest{Bidder: &bidder}); err != nil {
		t.Errorf("Failed to create bidder %v", bidder)
	}
	if _, err := (*c).CreateBidder(context.Background(), &rpc.CreateBidderRequest{Bidder: &mywish_bidder}); err != nil {
		t.Errorf("Failed to create bidder %v", mywish_bidder)
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
		if resp.Product.GetBuyerId() != 5 || resp.Product.GetDealPrice().GetAmount() != 50 {
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
	if fBidder.GetInShippingProductIds()[0] != 3 || math.Abs(fBidder.GetTotalAmountPending().GetAmount() - 200) > 0.0001 ||
		len(fBidder.GetPendingBids()) != 0 {
		t.Error("Failed to update bidder after closing the product.")
	}
	if math.Abs(fBuyer.GetTotalAmountPending().GetAmount()-150) > 0.0001 {
		t.Error("Failed to charge buyer after closing the product.")
	}
	if fProduct.GetStatus() != rpc.Product_PENDING {
		t.Error("Failed to upate product after closing the transaction.")
	}

}
