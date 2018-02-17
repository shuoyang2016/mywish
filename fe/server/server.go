package server

import (
	"fmt"
	"log"
	"net"

	"github.com/golang/glog"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/shuoyang2016/mywish/auth"
	"github.com/shuoyang2016/mywish/db"
	rpcpb "github.com/shuoyang2016/mywish/rpc"
	"go.uber.org/zap"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

func StartServer(port string) {
	glog.Info("Configure and start new server.")
	lis, err := net.Listen("tcp", port)
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
	serverIns, err := NewServer()
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
	rpcpb.RegisterMyWishServiceServer(s, serverIns)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	<-serverIns.stop

}

type Server struct {
	Auth  *auth.AuthModule
	Mongo *db.MongoConnection
	stop  chan struct{}
}

func NewServer() (*Server, error) {
	auth_module, err := auth.NewAuthModule()
	if err != nil {
		return nil, err
	}
	server := Server{
		Auth: auth_module,
		stop: make(chan struct{}),
	}
	mongoSession := db.StartMongoConnection("mywish_mongo", "")
	server.Mongo = mongoSession
	return &server, nil
}

func (s *Server) CreateProduct(ctx context.Context, req *rpcpb.CreateProductRequest) (*rpcpb.CreateProductResponse, error) {
	glog.V(3)
	product := req.NewProduct
	response := rpcpb.CreateProductResponse{Status: rpcpb.Error_SUCCESS}
	if product.Id == 0 || product.Name == "" {
		response.Status = rpcpb.Error_GENERIC_FAILURE
		response.Msg = "Either ID or name of the product is empty"
		return &response, status.Error(codes.InvalidArgument, "Either ID or name of the product is empty")
	}
	session := s.Mongo.BaseSession.Clone()
	c := session.DB(s.Mongo.DB).C("product")
	c.Insert(product)
	return &response, status.Error(codes.OK, " ")
}

func (s *Server) GetProduct(ctx context.Context, req *rpcpb.GetProductRequest) (*rpcpb.GetProductResponse, error) {
	glog.Info(*req)
	response := rpcpb.GetProductResponse{}
	return &response, nil
}

func (s *Server) CheckOrCreateUser(ctx context.Context, req *rpcpb.CheckOrCreateUserRequest) (*rpcpb.CheckOrCreateUserResponse, error) {
	_ = ctx
	response := rpcpb.CheckOrCreateUserResponse{}
	err := s.Auth.CheckOrCreateUser(req.UserName, req.Password)
	if err == auth.ErrUserNameExist {
		response.Succeed = false
		response.Details = fmt.Sprintf("The user name %v is already exist.", req.UserName)
	}
	return &response, err
}

func (s *Server) UpdateProduct(ctx context.Context, in *rpcpb.UpdateProductRequest) (*rpcpb.UpdateProductResponse, error) {
	return &rpcpb.UpdateProductResponse{}, nil
}
func (s *Server) GetProducts(ctx context.Context, in *rpcpb.GetProductsRequest) (*rpcpb.GetProductsResponse, error) {
	return &rpcpb.GetProductsResponse{}, nil
}
func (s *Server) CreateUser(ctx context.Context, in *rpcpb.CreateUserRequest) (*rpcpb.CreateUserResponse, error) {
	return &rpcpb.CreateUserResponse{}, nil
}
func (s *Server) GetUser(ctx context.Context, in *rpcpb.GetUserRequest) (*rpcpb.GetUserResponse, error) {
	return &rpcpb.GetUserResponse{}, nil
}
func (s *Server) DeleteUser(ctx context.Context, in *rpcpb.DeleteUserRequest) (*rpcpb.DeleteUserResponse, error) {
	return &rpcpb.DeleteUserResponse{}, nil
}
func (s *Server) UpdateUser(ctx context.Context, in *rpcpb.UpdateUserRequest) (*rpcpb.UpdateUserResponse, error) {
	return &rpcpb.UpdateUserResponse{}, nil
}
func (s *Server) AuthUser(ctx context.Context, in *rpcpb.AuthUserRequest) (*rpcpb.AuthUserResponse, error) {
	return &rpcpb.AuthUserResponse{}, nil
}
