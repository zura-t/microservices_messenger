package main

import (
	"context"
	"log"
	"net"
	"strings"
	"sync"
	"sync/atomic"

	pb "github.com/zura-t/go_messenger/accounts/pkg/accounts"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var idSerial uint64

type server struct {
	pb.UnimplementedAccountsServiceServer

	mx    sync.RWMutex
	users map[uint64]string
}

func NewServer() *server {
	return &server{users: make(map[uint64]string)}
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

	if err := validateRegisterRequest(req); err != nil {
		return nil, err
	}

	id := atomic.AddUint64(&idSerial, 1)

	s.mx.Lock()
	s.users[id] = username
	s.mx.Unlock()

	return &pb.UserRegisterResponse{
		Id: id,
	}, nil
}

func validateRegisterRequest(req *pb.RegisterRequest) error {
	username := req.GetUsername()
	if err := validateUsername(username); err != nil {
		return err
	}

	email := req.GetEmail()
	if err := validateEmail(email); err != nil {
		return err
	}

	password := req.GetPassword()
	if err := validatePassword(password); err != nil {
		return err
	}

	description := req.GetDescription()
	if err := validateDescription(description); err != nil {
		return err
	}

	name := req.GetName()
	if err := validateName(name); err != nil {
		return err
	}

	return nil
}

func validateName(name string) error {
	if name == "" {
		return status.Error(codes.InvalidArgument, "name is required")
	}
	if len(name) < 2 || len(name) > 50 {
		return status.Error(codes.InvalidArgument, "name must be between 2 and 50 characters")
	}
	if !isValidUsername(name) {
		return status.Error(codes.InvalidArgument, "name can only contain letters, numbers, and underscores")
	}
	return nil
}

func validateDescription(description string) error {
	if description == "" {
		return status.Error(codes.InvalidArgument, "description is required")
	}
	if len(description) < 10 || len(description) > 200 {
		return status.Error(codes.InvalidArgument, "description must be between 10 and 200 characters")
	}
	if !isValidUsername(description) {
		return status.Error(codes.InvalidArgument, "description can only contain letters, numbers, and underscores")
	}
	return nil
}

func validatePassword(password string) error {
	if password == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}
	if len(password) < 8 || len(password) > 20 {
		return status.Error(codes.InvalidArgument, "password must be between 8 and 20 characters")
	}
	if !isValidPassword(password) {
		return status.Error(codes.InvalidArgument, "password must contain at least one letter, one number, and one special character")
	}
	return nil
}

func validateEmail(email string) error {
	if email == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}

	if len(email) < 5 || len(email) > 50 {
		return status.Error(codes.InvalidArgument, "email must be between 5 and 50 characters")
	}

	if !isValidEmail(email) {
		return status.Error(codes.InvalidArgument, "email must be a valid email address")
	}

	return nil
}

func validateUsername(username string) error {
	if username == "" {
		return status.Error(codes.InvalidArgument, "username is required")
	}

	if len(username) < 3 || len(username) > 20 {
		return status.Error(codes.InvalidArgument, "username must be between 3 and 20 characters")
	}

	if !isValidUsername(username) {
		return status.Error(codes.InvalidArgument, "username can only contain letters, numbers, and underscores")
	}

	return nil
}

func isValidPassword(password string) bool {
	return len(password) >= 8 && len(password) <= 20
}

func isValidUsername(username string) bool {
	for _, char := range username {
		if !(('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z') || ('0' <= char && char <= '9') || char == '_') {
			return false
		}
	}
	return true
}

func isValidEmail(email string) bool {
	// Простейшая проверка на наличие символа "@" и "."
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func (s *server) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.UserProfile, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	id := req.GetId()
	if id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	return &pb.UserProfile{
		Id:       id,
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

func main() {
	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	implementation := NewServer() // наша реализация сервера

	server := grpc.NewServer()
	pb.RegisterAccountsServiceServer(server, implementation) // регистрация обработчиков

	reflection.Register(server) // регистрируем дополнительные обработчики

	log.Printf("server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
