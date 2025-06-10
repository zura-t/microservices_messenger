package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"sync/atomic"

	"buf.build/go/protovalidate"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/zura-t/go_messenger/api-gateway/pkg/api-gateway"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var idSerial uint64

type server struct {
	pb.UnimplementedApiGatewayServiceServer

	mx        sync.RWMutex
	users     map[uint64]string
	validator *protovalidate.Validator
}

func NewServer() (*server, error) {
	server := &server{}

	validator, err := protovalidate.New(
		protovalidate.WithDisableLazy(),
		protovalidate.WithMessages(
			&pb.RegisterRequest{},
			&pb.LoginRequest{},
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize validator: %w", err)
	}

	server.validator = &validator

	return server, nil
}

func protovalidateViolationsToFieldViolations(violations []*protovalidate.Violation) []*errdetails.BadRequest_FieldViolation {
	fieldViolations := make([]*errdetails.BadRequest_FieldViolation, len(violations))
	for i, v := range violations {
		fieldViolations[i] = &errdetails.BadRequest_FieldViolation{
			Field:       v.Proto.Field.String(),
			Description: *v.Proto.Message,
		}
	}
	return fieldViolations
}

func convertProtovalidateValidationErrorToErrorBadRequest(validationError *protovalidate.ValidationError) *errdetails.BadRequest {
	return &errdetails.BadRequest{
		FieldViolations: protovalidateViolationsToFieldViolations(validationError.Violations),
	}
}

func rpcValidationError(err error) error {
	if err == nil {
		return nil
	}

	var validationError *protovalidate.ValidationError
	if ok := errors.As(err, &validationError); ok {
		st, err := status.New(codes.InvalidArgument, codes.InvalidArgument.String()).WithDetails(convertProtovalidateValidationErrorToErrorBadRequest(validationError))
		if err != nil {
			return st.Err()
		}
	}

	return status.Error(codes.Internal, err.Error())
}

func (s *server) Register(_ context.Context, req *pb.RegisterRequest) (*pb.UserRegisterResponse, error) {
	username := req.GetUsername()
	id := atomic.AddUint64(&idSerial, 1)

	s.mx.Lock()
	s.users[id] = username
	s.mx.Unlock()

	return &pb.UserRegisterResponse{
		Id: id,
	}, nil
}

func (s *server) Login(_ context.Context, req *pb.LoginRequest) (*pb.UserLoginResponse, error) {
	return &pb.UserLoginResponse{
		Id: 1,
	}, nil
}

func (s *server) RefreshToken(_ context.Context, _ *emptypb.Empty) (*pb.RefreshTokenResponse, error) {
	return &pb.RefreshTokenResponse{}, nil
}

func (s * server) Logout(_ context.Context, _ *emptypb.Empty) (*pb.LogoutResponse, error) {
	return &pb.LogoutResponse{
		Message: "Logged out",
	}, nil
}

func (s *server) GetUser(_ context.Context, req *pb.GetUserRequest) (*pb.UserProfile, error) {
	return &pb.UserProfile{}, nil
}

func (s *server) GetProfile(_ context.Context, req *pb.GetProfileRequest) (*pb.UserProfile, error) {
	return &pb.UserProfile{}, nil
}

func (s *server) UpdateProfile(_ context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	return &pb.UpdateProfileResponse{}, nil
}

func (s *server) DeleteProfile(_ context.Context, req *pb.DeleteProfileRequest) (*pb.DeleteProfileResponse, error) {
	return &pb.DeleteProfileResponse{}, nil
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	server, err := NewServer()
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		grpcServer := grpc.NewServer()
		pb.RegisterApiGatewayServiceServer(grpcServer, server)

		reflection.Register(grpcServer)

		lis, err := net.Listen("tcp", ":8082")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		log.Printf("server listening on %v", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		mux := runtime.NewServeMux()
		if err = pb.RegisterApiGatewayServiceHandlerServer(ctx, mux, server); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
		httpServer := &http.Server{Handler: mux}

		lis, err := net.Listen("tcp", ":8083")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		log.Printf("server listening on %v", lis.Addr())
		if err := httpServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	wg.Wait()
}
