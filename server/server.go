package server

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/globalsign/mgo/bson"
	"github.com/golang/glog"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/shuoyang2016/mywish/rpc"
	"github.com/shuoyang2016/mywish/server/auth"
	"github.com/shuoyang2016/mywish/server/config"
	"github.com/shuoyang2016/mywish/server/db"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"strconv"
)

func runRestService(restPort int, grpcPort int) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	echoEndpoint := "localhost:" + string(grpcPort)

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := rpc.RegisterMyWishServiceHandlerFromEndpoint(ctx, mux, echoEndpoint, opts)
	if err != nil {
		return err
	}
	return http.ListenAndServe(":"+strconv.Itoa(restPort), mux)
}

func StartServer(cfg *config.Config) chan struct{} {
	glog.Info("Start REST service gateway")
	go runRestService(cfg.RestPort, cfg.GrpcPort)
	glog.Info("Start GRPC service.")
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.GrpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_opentracing.StreamServerInterceptor(),
			grpc_prometheus.StreamServerInterceptor,
			grpc_zap.StreamServerInterceptor(zap.New(nil)),
			grpc_auth.StreamServerInterceptor(auth.AuthFunc),
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_opentracing.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			grpc_zap.UnaryServerInterceptor(zap.New(nil)),
			grpc_auth.UnaryServerInterceptor(auth.AuthFunc),
			grpc_recovery.UnaryServerInterceptor(),
		)),
	)
	serverIns, err := NewServer(cfg)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
	rpc.RegisterMyWishServiceServer(s, serverIns)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	return serverIns.stop
}

type Server struct {
	Config *config.Config
	Auth   *auth.AuthModule
	Mongo  *db.MongoConnection
	stop   chan struct{}
}

func NewServer(cfg *config.Config) (*Server, error) {
	server := Server{
		Config: cfg,
		stop:   make(chan struct{}),
	}
	auth_module, err := auth.NewAuthModule(cfg.SqlAddress)
	if err != nil {
		return nil, err
	}
	server.Auth = auth_module
	option := db.Option{DB: cfg.DBName, PlayerSCollection: cfg.PlayersCollection,
		ProductsCollection: cfg.ProductsCollection, URL: cfg.MongoAddress}
	mongoSession, err := db.StartMongoConnection(&option)
	if err != nil {
		return nil, err
	}
	server.Mongo = mongoSession
	return &server, nil
}

func (s *Server) CreateProduct(ctx context.Context, req *rpc.CreateProductRequest) (*rpc.CreateProductResponse, error) {
	glog.V(3).Info(*req)
	product := req.NewProduct
	response := rpc.CreateProductResponse{Status: rpc.Error_SUCCESS}
	if product.Id == 0 || product.Name == "" {
		response.Status = rpc.Error_GENERIC_FAILURE
		response.Msg = "Either ID or name of the product is empty"
		return &response, status.Error(codes.InvalidArgument, "Either ID or name of the product is empty")
	}
	session := s.Mongo.BaseSession.Clone()
	c := session.DB(s.Mongo.DB).C(s.Mongo.ProductsCollection)
	c.Insert(product)
	return &response, status.Error(codes.OK, " ")
}

func (s *Server) GetProduct(ctx context.Context, req *rpc.GetProductRequest) (*rpc.GetProductResponse, error) {
	glog.V(3).Info(*req)
	c := s.Mongo.BaseSession.Clone()
	ret := rpc.GetProductResponse{Product: &rpc.Product{}}
	err := c.DB(s.Mongo.DB).C(s.Mongo.ProductsCollection).Find(bson.M{"name": req.GetProductName()}).One(ret.GetProduct())
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func (s *Server) CheckOrCreateUser(ctx context.Context, req *rpc.CheckOrCreateUserRequest) (*rpc.CheckOrCreateUserResponse, error) {
	_ = ctx
	response := rpc.CheckOrCreateUserResponse{}
	err := s.Auth.CheckOrCreateUser(req.UserName, req.Password)
	if err == auth.ErrUserNameExist {
		response.Succeed = false
		response.Details = fmt.Sprintf("The user name %v is already exist.", req.UserName)
	}
	return &response, err
}

func (s *Server) UpdateProduct(ctx context.Context, in *rpc.UpdateProductRequest) (*rpc.UpdateProductResponse, error) {

	return &rpc.UpdateProductResponse{}, nil
}
func (s *Server) GetProducts(ctx context.Context, in *rpc.GetProductsRequest) (*rpc.GetProductsResponse, error) {
	return &rpc.GetProductsResponse{}, nil
}
func (s *Server) CreateUser(ctx context.Context, in *rpc.CreateUserRequest) (*rpc.CreateUserResponse, error) {
	return &rpc.CreateUserResponse{}, nil
}
func (s *Server) GetUser(ctx context.Context, in *rpc.GetUserRequest) (*rpc.GetUserResponse, error) {
	return &rpc.GetUserResponse{}, nil
}
func (s *Server) DeleteUser(ctx context.Context, in *rpc.DeleteUserRequest) (*rpc.DeleteUserResponse, error) {
	return &rpc.DeleteUserResponse{}, nil
}
func (s *Server) UpdateUser(ctx context.Context, in *rpc.UpdateUserRequest) (*rpc.UpdateUserResponse, error) {
	return &rpc.UpdateUserResponse{}, nil
}
func (s *Server) AuthUser(ctx context.Context, in *rpc.AuthUserRequest) (*rpc.AuthUserResponse, error) {
	return &rpc.AuthUserResponse{}, nil
}
func (s *Server) CreateBidder(ctx context.Context, in *rpc.CreateBidderRequest) (*rpc.CreateBidderResponse, error) {
	newBidder := rpc.Bidder{Id:in.GetBidderId(), Name:in.GetBidderName()}
	s.Mongo.BaseSession.DB(s.Mongo.DB).C(s.Mongo.PlayerSCollection).Insert(&newBidder)
	return &rpc.CreateBidderResponse{}, nil
}
func (s *Server) UpdateBidder(ctx context.Context, in *rpc.UpdateBidderRequest) (*rpc.UpdateBidderResponse, error) {
	return &rpc.UpdateBidderResponse{}, nil
}
func (s *Server) GetBidder(ctx context.Context, in *rpc.GetBidderRequest) (*rpc.GetBidderResponse, error) {
	return &rpc.GetBidderResponse{}, nil
}
func (s *Server) BidProduct(ctx context.Context, in *rpc.BidProductRequest) (*rpc.BidProductResponse, error) {
	err := BidFlow(s, in)
	status := rpc.Error_SUCCESS
	if err != nil {
		status = rpc.Error_GENERIC_FAILURE
	}
	return &rpc.BidProductResponse{Error: status}, err
}