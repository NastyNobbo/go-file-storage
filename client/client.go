package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"image"
	_ "image/jpeg"
	_ "image/png"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"google.golang.org/grpc"

	pb "C/storage"
)

var ErrFileNotFound = errors.New("file not found")
var fileList = make(map[string]string)

func main() {
	dirPath := "C:/praktika/go-file-storage/files"

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Fatalf("Error reading directory: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() {
			fileName := filepath.Base(file.Name())
			ext := filepath.Ext(fileName)
			fileList[strings.TrimSuffix(fileName, ext)] = ext
			fmt.Printf("Added file: %s\n", fileName)
		}
	}

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

	fileSelect := widget.NewSelect(getFileList(), func(s string) {
		fileIDEntry.SetText(s)
		if ext, ok := fileList[s]; ok {
			if !contains(extensionOptions, ext) {
				extensionOptions = append(extensionOptions, ext)
				extensionSelect.Options = extensionOptions
			}
			extensionSelect.SetSelected(ext)
		}
	})
	fileSelect.PlaceHolder = "Выберите файл"

	idAndExtension := container.NewGridWithColumns(3, fileSelect, fileIDEntry, extensionSelect)

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
		fileList[createFileResponse.Id] = extension
		fileSelect.Options = getFileList()
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

		if isImage(readFileResponse.File, extension) {
			showImage(readFileResponse.File, extension, w)
		} else {
			dialog.ShowInformation("Содержимое файла", string(readFileResponse.File), w)
		}
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
		fileList[fileID] = extension
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
		delete(fileList, fileID)
		fileIDEntry.SetText("")
		extensionSelect.PlaceHolder = "Расширение файла"
		fileSelect.Options = getFileList()
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

func getFileList() []string {
	var list []string
	for id := range fileList {
		list = append(list, id)
	}
	return list
}

func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func isImage(data []byte, ext string) bool {
	img, _, err := image.DecodeConfig(bytes.NewReader(data))
	return err == nil && (ext == ".jpg" || ext == ".jpeg" || ext == ".png") && img.Width > 0 && img.Height > 0
}

func showImage(data []byte, ext string, w fyne.Window) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Printf("Error decoding image: %v", err)
		dialog.ShowError(errors.New("Не удалось открыть изображение"), w)
		return
	}

	fyneImg := fyne.NewStaticResource("image", data)
	imgFile := canvas.NewImageFromResource(fyneImg)

	width := float32(img.Bounds().Dx())
	height := float32(img.Bounds().Dy())

	imgWin := fyne.CurrentApp().NewWindow("Изображение")
	imgWin.Resize(fyne.NewSize(width, height))

	imgContainer := container.NewMax(imgFile)

	imgWin.SetContent(imgContainer)
	imgWin.Show()
}
