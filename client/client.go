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

	extensionOptions := []string{".txt", ".docx", ".pdf"}
	extensionSelect := widget.NewSelect(extensionOptions, func(s string) {
		fmt.Println("Выбрано расширение:", s)
	})
	extensionSelect.PlaceHolder = "Расширение файла"

	fileContent := widget.NewMultiLineEntry()
	fileContent.SetPlaceHolder("Содержимое файла")

	idAndExtension := container.NewGridWithColumns(2, fileIDEntry, extensionSelect)

	createFileButton := widget.NewButton("Создание файла", func() {
		extension := extensionSelect.Selected
		if len(extension) == 0 {
			dialog.ShowError(errors.New("Пожалуйста, выберите расширение файла"), w)
			return
		}

		createFileResponse, err := client.CreateFile(context.Background(), &pb.CreateFileRequest{
			File:      []byte(fileContent.Text),
			Extension: extension,
		})
		if err != nil {
			log.Printf("Error creating file: %v", err)
			return
		}
		fmt.Printf("Файл создан с ID: %s\n", createFileResponse.Id)
		fileIDEntry.SetText(createFileResponse.Id)
		extensionSelect.PlaceHolder = "Расширение файла"
	})

	readFileButton := widget.NewButton("Чтение файла", func() {
		fileID := fileIDEntry.Text
		extension := extensionSelect.Selected
		if len(fileID) == 0 || len(extension) == 0 {
			dialog.ShowError(errors.New("Пожалуйста, введите ID и выберите расширение файла"), w)
			return
		}

		readFileResponse, err := client.ReadFile(context.Background(), &pb.ReadFileRequest{
			Id:        fileID,
			Extension: extension,
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
		dialog.ShowInformation("Содержимое файла", string(readFileResponse.File), w)
	})

	updateFileButton := widget.NewButton("Обновить файл", func() {
		fileID := fileIDEntry.Text
		extension := extensionSelect.Selected
		if len(fileID) == 0 || len(extension) == 0 {
			dialog.ShowError(errors.New("Пожалуйста, введите ID и выберите расширение файла"), w)
			return
		}

		updateFileResponse, err := client.UpdateFile(context.Background(), &pb.UpdateFileRequest{
			Id:        fileID,
			File:      []byte(fileContent.Text),
			Extension: extension,
		})
		if err != nil {
			log.Printf("Error updating file: %v", err)
			return
		}
		fmt.Printf("Файл обновлён: %v\n", updateFileResponse)
	})

	deleteFileButton := widget.NewButton("Удаление файла", func() {
		fileID := fileIDEntry.Text
		extension := extensionSelect.Selected
		if len(fileID) == 0 || len(extension) == 0 {
			dialog.ShowError(errors.New("Пожалуйста, введите ID и выберите расширение файла"), w)
			return
		}

		deleteFileResponse, err := client.DeleteFile(context.Background(), &pb.DeleteFileRequest{
			Id:        fileID,
			Extension: extension,
		})
		if err != nil {
			log.Printf("Error deleting file: %v", err)
			dialog.ShowError(errors.New("Файл не найден"), w)
			return
		}
		fmt.Printf("Файл удалён: %v\n", deleteFileResponse)
		fileIDEntry.SetText("")
		extensionSelect.PlaceHolder = "Расширение файла"
	})

	buttons := container.NewGridWithColumns(2,
		createFileButton,
		readFileButton,
		updateFileButton,
		deleteFileButton,
	)

	content := container.NewVBox(
		idAndExtension,
		fileContent,
		buttons,
	)

	border := container.NewBorder(content, buttons, nil, nil)

	w.SetContent(border)
	w.ShowAndRun()
}
