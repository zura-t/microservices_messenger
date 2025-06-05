package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"buf.build/go/protovalidate"
	pb "github.com/zura-t/go_messenger/mailer/pkg/mailer"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type server struct {
	pb.UnimplementedMailerServiceServer

	mx        sync.RWMutex
	validator *protovalidate.Validator
}

func NewServer() (*server, error) {
	server := &server{}

	validator, err := protovalidate.New(
		protovalidate.WithDisableLazy(),
		protovalidate.WithMessages(
			&pb.SendEmailRequest{},
			&pb.SendEmailWithAttachmentRequest{},
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

func (s *server) SendEmail(ctx context.Context, req *pb.SendEmailRequest) (*pb.SendEmailResponse, error) {
	validator := *s.validator
	if err := validator.Validate(req); err != nil {
		return nil, rpcValidationError(err)
	}

	s.mx.Lock()
	// * send email
	s.mx.Unlock()
	return &pb.SendEmailResponse{Message: "Email sent successfully"}, nil
}

type Attachment struct {
	filename string
	content  []byte
}

func (s *server) SendEmailWithAttachment(ctx context.Context, req *pb.SendEmailWithAttachmentRequest) (*pb.SendEmailResponse, error) {
	validator := *s.validator
	if err := validator.Validate(req); err != nil {
		return nil, rpcValidationError(err)
	}

	return &pb.SendEmailResponse{Message: "Email with attachment sent successfully"}, nil
}

func main() {
	server, err := NewServer()
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	lis, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	validator, err := protovalidate.New(
		protovalidate.WithDisableLazy(),
		protovalidate.WithMessages(
			&pb.SendEmailRequest{},
			&pb.SendEmailWithAttachmentRequest{},
		),
	)
	if err != nil {

	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
				if err := validator.Validate(req.(proto.Message)); err != nil {
					return nil, rpcValidationError(err)
				}
				return handler(ctx, req)
			},
		),
		grpc.ChainStreamInterceptor(),
	)

	pb.RegisterMailerServiceServer(grpcServer, server)

	reflection.Register(grpcServer)

	log.Printf("server listening on %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
