package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"

	pb "github.com/zura-t/go_messenger/accounts/pkg/accounts"

	"buf.build/go/protovalidate"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var idSerial uint64

type server struct {
	pb.UnimplementedAccountsServiceServer

	mx        sync.RWMutex
	users     map[uint64]string
	validator *protovalidate.Validator
}

func NewServer() (*server, error) {
	server := &server{users: make(map[uint64]string)}
	validator, err := protovalidate.New(
		protovalidate.WithDisableLazy(),
		protovalidate.WithMessages(
			&pb.RegisterRequest{},
			&pb.LoginRequest{},
			&pb.CreateUserRequest{},
			&pb.GetProfileRequest{},
			&pb.GetUserRequest{},
		),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create validator: %w", err)
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

func convertProtovalidateValidationErrorToErrorBadRequest(err *protovalidate.ValidationError) *errdetails.BadRequest {
	return &errdetails.BadRequest{
		FieldViolations: protovalidateViolationsToFieldViolations(err.Violations),
	}
}

func rpcValidationError(err error) error {
	if err == nil {
		return nil
	}

	var valErr *protovalidate.ValidationError
	if ok := errors.As(err, &valErr); ok {
		status, err := status.New(codes.InvalidArgument, codes.InvalidArgument.String()).
			WithDetails(convertProtovalidateValidationErrorToErrorBadRequest(valErr))
		if err == nil {
			return status.Err()
		}
	}

	return status.Error(codes.Internal, err.Error())
}

func (s *server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.UserRegisterResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Заголовков нет")
	} else {
		const key = "x-header"
		log.Println(key, md.Get(key))
	}

	username := req.GetUsername()

	validator := *s.validator

	if err := validator.Validate(req); err != nil {
		return nil, rpcValidationError(err)
	}

	id := atomic.AddUint64(&idSerial, 1)

	s.mx.Lock()
	s.users[id] = username
	s.mx.Unlock()

	return &pb.UserRegisterResponse{
		Id: id,
	}, nil
}

func (s *server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.UserLoginResponse, error) {
	email := req.GetEmail()

	return &pb.UserLoginResponse{
		Id:       1,
		Email:    email,
		Name:     "John Doe",
		Username: "johndoe",
	}, nil
}

func (s *server) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.UserProfile, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	id := req.GetId()
	if id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	return &pb.UserProfile{
		Id: id,
	}, nil
}

func (s *server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserProfile, error) {
	s.mx.RLock()
	// закрываем блокировку после выполнения функции
	// чтобы другие горутины могли получить доступ к данным
	// блокировка на чтение
	// блокировка на запись

	// блокировка на чтение позволяет другим горутинам читать данные
	// блокировка на запись позволяет только одной горутине изменять данные

	defer s.mx.RUnlock()

	id := req.GetId()
	if id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	username, ok := s.users[id]
	if !ok {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &pb.UserProfile{
		Id:       id,
		Username: username,
	}, nil
}

func CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserProfile, error) {
	// Здесь должна быть логика создания пользователя
	// Например, сохранение пользователя в базу данных
	// или в память сервера

	// Для примера просто возвращаем пользователя с ID 1
	return &pb.UserProfile{
		Id:       1,
		Username: req.GetUsername(),
	}, nil
}

func (s *server) UpdateUser(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	return &pb.UpdateProfileResponse{
		Id:       1,
		Username: req.GetUsername(),
	}, nil
}

func (s *server) DeleteUser(ctx context.Context, req *pb.DeleteProfileRequest) (*pb.DeleteProfileResponse, error) {
	id := req.GetId()
	return &pb.DeleteProfileResponse{
		Message: fmt.Sprintf("Profile with ID %d deleted", id),
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server, err := NewServer() // наша реализация сервера
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	validator, err := protovalidate.New(
		protovalidate.WithDisableLazy(),
		protovalidate.WithMessages(
			&pb.RegisterRequest{},
			&pb.LoginRequest{},
			&pb.CreateUserRequest{},
			&pb.GetProfileRequest{},
			&pb.GetUserRequest{},
		),
	)
	if err != nil {

	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (response any, err error) {
				if err := validator.Validate(req.(proto.Message)); err != nil {
					return nil, rpcValidationError(err)
				}
				return handler(ctx, req)
			},
		),
		grpc.ChainStreamInterceptor(),
	)
	pb.RegisterAccountsServiceServer(grpcServer, server) // регистрация обработчиков

	reflection.Register(grpcServer) // регистрируем дополнительные обработчики

	log.Printf("server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
