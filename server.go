package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/NastyNobbo/go-file-storage/storage"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const storagePath = "./storage"

type server struct{}

func (s *server) CreateFile(ctx context.Context, req *storage.CreateFileRequest) (*storage.CreateFileResponse, error) {
	fileID := generateFileID()
	filePath := filepath.Join(storagePath, fileID)

	err := ioutil.WriteFile(filePath, req.File, 0644)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create file: %v", err)
	}

	return &storage.CreateFileResponse{Id: fileID}, nil
}

func (s *server) ReadFile(ctx context.Context, req *storage.ReadFileRequest) (*storage.ReadFileResponse, error) {
	filePath := filepath.Join(storagePath, req.Id)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "File not found: %v", err)
	}

	return &storage.ReadFileResponse{File: data}, nil
}

func (s *server) UpdateFile(ctx context.Context, req *storage.UpdateFileRequest) (*empty.Empty, error) {
	filePath := filepath.Join(storagePath, req.Id)

	err := ioutil.WriteFile(filePath, req.File, 0644)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update file: %v", err)
	}

	return &empty.Empty{}, nil
}

func (s *server) DeleteFile(ctx context.Context, req *storage.DeleteFileRequest) (*empty.Empty, error) {
	filePath := filepath.Join(storagePath, req.Id)

	err := os.Remove(filePath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete file: %v", err)
	}

	return &empty.Empty{}, nil
}

func generateFileID() string {
	return strings.ReplaceAll(fmt.Sprintf("%v", time.Now().UnixNano()), "-", "")
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	storage.RegisterFileStorageServer(s, &server{})

	err = os.MkdirAll(storagePath, 0755)
	if err != nil {
		log.Fatalf("Failed to create storage directory: %v", err)
	}

	log.Printf("Server is listening on %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
