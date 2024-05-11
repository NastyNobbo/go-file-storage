package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"google.golang.org/grpc"

	pb "C/storage"
)

var ErrFileNotFound = errors.New("file not found")

func main() {
	a := app.New()
	w := a.NewWindow("Задание")
	w.Resize(fyne.NewSize(600, 300))

	conn, err := grpc.Dial(":50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	client := pb.NewFileStorageClient(conn)

	fileIDEntry := widget.NewEntry()
	fileIDEntry.SetPlaceHolder("Id файла")

	createFileButton := widget.NewButton("Создание файла", func() {
		createFileResponse, err := client.CreateFile(context.Background(), &pb.CreateFileRequest{
			File: []byte("Привет, я файл"),
		})
		if err != nil {
			log.Printf("Error creating file: %v", err)
			return
		}
		fmt.Printf("File created with ID: %s\n", createFileResponse.Id)
		fileIDEntry.SetText(createFileResponse.Id)
	})

	readFileButton := widget.NewButton("Чтение файла", func() {
		fileID := fileIDEntry.Text
		if fileID == "" {
			dialog.ShowError(errors.New("Пожалуйста, введите ID файла"), w)
			return
		}

		readFileResponse, err := client.ReadFile(context.Background(), &pb.ReadFileRequest{
			Id: fileID,
		})
		if err != nil {
			if err == ErrFileNotFound {
				dialog.ShowError(errors.New("Файл не найден"), w)
			} else {
				log.Printf("Error reading file: %v", err)
				dialog.ShowError(errors.New("Файл не найден"), w)
			}
			return
		}
		dialog.ShowInformation("File Content", string(readFileResponse.File), w)
	})

	updateFileButton := widget.NewButton("Обновить файл", func() {
		fileID := fileIDEntry.Text
		if fileID == "" {
			dialog.ShowError(errors.New("Пожалуйста, введите ID файла"), w)
			return
		}

		updateFileResponse, err := client.UpdateFile(context.Background(), &pb.UpdateFileRequest{
			Id:   fileID,
			File: []byte("Привет, я обновлённый файл"),
		})
		if err != nil {
			log.Printf("Error updating file: %v", err)
			return
		}
		fmt.Printf("File updated: %v\n", updateFileResponse)
	})

	deleteFileButton := widget.NewButton("Удаление файла", func() {
		fileID := fileIDEntry.Text
		if fileID == "" {
			dialog.ShowError(errors.New("Пожалуйста, введите ID файла"), w)
			return
		}

		deleteFileResponse, err := client.DeleteFile(context.Background(), &pb.DeleteFileRequest{
			Id: fileID,
		})
		if err != nil {
			log.Printf("Error deleting file: %v", err)
			dialog.ShowError(errors.New("Файл не найден"), w)
			return
		}
		fmt.Printf("File deleted: %v\n", deleteFileResponse)
		fileIDEntry.SetText("")
	})

	buttons := container.NewGridWithColumns(2,
		createFileButton,
		readFileButton,
		updateFileButton,
		deleteFileButton,
	)

	content := container.NewVBox(
		fileIDEntry,
		buttons,
	)

	border := container.NewBorder(content, buttons, nil, nil)

	w.SetContent(border)
	w.ShowAndRun()
}
