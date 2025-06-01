package main

import (
	"context"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	pb "github.com/zura-t/go_messenger/mailer/pkg/mailer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type server struct {
	pb.UnimplementedMailerServiceServer

	mx sync.RWMutex
}

func NewServer() *server {
	return &server{}
}

func (s *server) SendEmail(ctx context.Context, req *pb.SendEmailRequest) (*pb.SendEmailResponse, error) {
	err := validateSendEmailRequest(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}

	return &pb.SendEmailResponse{Message: "Email sent successfully"}, nil
}

type Attachment struct {
	filename string
	content  []byte
}

func (s *server) SendEmailWithAttachment(ctx context.Context, req *pb.SendEmailWithAttachmentRequest) (*pb.SendEmailResponse, error) {
	err := validateSendEmailWithAttachmentRequest(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}

	return &pb.SendEmailResponse{Message: "Email with attachment sent successfully"}, nil
}

func validateSendEmailWithAttachmentRequest(req *pb.SendEmailWithAttachmentRequest) error {
	to := req.GetTo()
	if err := validateTo(to); err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid email: %v", err)
	}

	subject := req.GetSubject()
	if err := validateSubject(subject); err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid subject: %v", err)
	}

	message := req.GetMessage()
	if err := validateMessage(message); err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid message: %v", err)
	}

	var attachment *pb.Attachment = req.GetAttachment()
	if attachment == nil {
		return status.Error(codes.InvalidArgument, "attachments cannot be empty")
	}

	return nil
}

func validateSendEmailRequest(req *pb.SendEmailRequest) error {
	to := req.GetTo()
	if err := validateTo(to); err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid email: %v", err)
	}

	subject := req.GetSubject()
	if err := validateSubject(subject); err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid subject: %v", err)
	}

	message := req.GetMessage()
	if err := validateMessage(message); err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid message: %v", err)
	}

	return nil
}

func validateTo(email string) error {
	if email == "" {
		return status.Error(codes.InvalidArgument, "email cannot be empty")
	}
	if !strings.Contains(email, "@") {
		return status.Error(codes.InvalidArgument, "email must contain '@'")
	}
	if !strings.Contains(email, ".") {
		return status.Error(codes.InvalidArgument, "email must contain a domain")
	}
	if len(email) < 5 || len(email) > 254 {
		return status.Error(codes.InvalidArgument, "email length must be between 5 and 254 characters")
	}
	return nil
}

func validateSubject(subject string) error {
	if subject == "" {
		return status.Error(codes.InvalidArgument, "subject cannot be empty")
	}
	if len(subject) < 1 || len(subject) > 100 {
		return status.Error(codes.InvalidArgument, "subject length must be between 1 and 100 characters")
	}
	return nil
}

func validateMessage(message string) error {
	if message == "" {
		return status.Error(codes.InvalidArgument, "message cannot be empty")
	}
	if len(message) < 1 || len(message) > 1000 {
		return status.Error(codes.InvalidArgument, "message length must be between 1 and 1000 characters")
	}
	return nil
}

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// for startup probe
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, struct{ Status string }{Status: "OK"})
	})

	// for readiness probe
	e.GET("/ready", func(c echo.Context) error {
		return c.JSON(http.StatusOK, struct{ Status string }{Status: "OK"})
	})

	e.GET("/hello", func(c echo.Context) error {
		return c.HTML(http.StatusOK, "Hello, Docker!")
	})

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8082"
	}

	time.Sleep(5 * time.Second)

	e.Logger.Fatal(e.Start(":" + httpPort))
}
