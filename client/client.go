package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"

	pb "github.com/NastyNobbo/go-file-storage/storage" // изменен импорт на правильный путь к protobuf-файлу
)

func main() {
	conn, err := grpc.Dial(":50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewFileStorageClient(conn)

	// Пример вызова метода CreateFile
	createFileResponse, err := client.CreateFile(context.Background(), &pb.CreateFileRequest{
		File: []byte("Hello, gRPC!"),
	})
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	fmt.Printf("File created with ID: %s\n", createFileResponse.Id)

	// Пример вызова метода ReadFile
	readFileResponse, err := client.ReadFile(context.Background(), &pb.ReadFileRequest{
		Id: createFileResponse.Id,
	})
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}
	fmt.Printf("File content: %s\n", readFileResponse.File)

	// Пример вызова метода UpdateFile
	updateFileResponse, err := client.UpdateFile(context.Background(), &pb.UpdateFileRequest{
		Id:   createFileResponse.Id,
		File: []byte("Hello, updated gRPC!"),
	})
	if err != nil {
		log.Fatalf("Error updating file: %v", err)
	}
	fmt.Printf("File updated: %v\n", updateFileResponse)

	// Пример вызова метода DeleteFile
	deleteFileResponse, err := client.DeleteFile(context.Background(), &pb.DeleteFileRequest{
		Id: createFileResponse.Id,
	})
	if err != nil {
		log.Fatalf("Error deleting file: %v", err)
	}
	fmt.Printf("File deleted: %v\n", deleteFileResponse)
}
