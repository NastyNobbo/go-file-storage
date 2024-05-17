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

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/NastyNobbo/go-file-storage/storage"
	pb "github.com/NastyNobbo/go-file-storage/storage"
)

const storagePath = "./files"

type server struct {
	storage.UnimplementedFileStorageServer
}

func (s *server) CreateFile(ctx context.Context, req *pb.CreateFileRequest) (*pb.CreateFileResponse, error) {
	fileID := generateFileID()
	fileExt := req.Extension

	if len(fileExt) == 0 {
		fileExt = ".txt"
	}

	if len(fileExt) > 1 && fileExt[0] != '.' {
		fileExt = "." + fileExt
	}

	filePath := filepath.Join(storagePath, fileID+fileExt)

	err := ioutil.WriteFile(filePath, req.File, 0644)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create file: %v", err)
	}

	return &pb.CreateFileResponse{Id: fileID, Extension: fileExt}, nil
}

func (s *server) ReadFile(ctx context.Context, req *pb.ReadFileRequest) (*pb.ReadFileResponse, error) {
	filePath := filepath.Join(storagePath, req.Id+req.Extension)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "File not found: %v", err)
	}

	return &pb.ReadFileResponse{File: data}, nil
}

func (s *server) UpdateFile(ctx context.Context, req *pb.UpdateFileRequest) (*pb.UpdateFileResponse, error) {
	filePath := filepath.Join(storagePath, req.Id+req.Extension)
	err := ioutil.WriteFile(filePath, req.File, 0644)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update file: %v", err)
	}

	return &pb.UpdateFileResponse{}, nil
}

func (s *server) DeleteFile(ctx context.Context, req *pb.DeleteFileRequest) (*pb.DeleteFileResponse, error) {
	filePath := filepath.Join(storagePath, req.Id+req.Extension)

	err := os.Remove(filePath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete file: %v", err)
	}

	return &pb.DeleteFileResponse{}, nil
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
	pb.RegisterFileStorageServer(s, &server{})

	err = os.MkdirAll(storagePath, 0755)
	if err != nil {
		log.Fatalf("Failed to create storage directory: %v", err)
	}

	log.Printf("Server is listening on %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
