package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type FileInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Hash     string `json:"hash"`
	Uploaded string `json:"uploaded"`
}

type FilesList struct {
	Files []FileInfo `json:"files"`
}

func request(method, url string, body interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if body != nil {
		err := json.NewEncoder(&buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func uploadFile(url string, filePath string) (FileInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return FileInfo{}, err
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return FileInfo{}, err
	}

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return FileInfo{}, err
	}

	data := map[string]interface{}{
		"name": fileStat.Name(),
		"size": fileStat.Size(),
		"data": fmt.Sprintf("%x", fileBytes),
	}

	body, err := request("POST", url, data)
	if err != nil {
		return FileInfo{}, err
	}

	var info FileInfo
	err = json.Unmarshal(body, &info)
	if err != nil {
		return FileInfo{}, err
	}

	return info, nil
}

func getFilesList(url string) (FilesList, error) {
	body, err := request("GET", url, nil)
	if err != nil {
		return FilesList{}, err
	}

	var list FilesList
	err = json.Unmarshal(body, &list)
	if err != nil {
		return FilesList{}, err
	}

	return list, nil
}

func deleteFile(url, id string) error {
	data := map[string]string{"id": id}
	_, err := request("DELETE", url, data)
	return err
}

func main() {
	const serverURL = "http://localhost:8080/api"

	// Upload a file
	info, err := uploadFile(serverURL, "path/to/your/file")
	if err != nil {
		fmt.Println("Error uploading file:", err)
	} else {
		fmt.Println("File uploaded successfully:", info)
	}

	// Get files list
	list, err := getFilesList(serverURL)
	if err != nil {
		fmt.Println("Error getting files list:", err)
	} else {
		fmt.Println("Files list:", list)
	}

	// Delete a file
	err = deleteFile(serverURL, "file_id")
	if err != nil {
		fmt.Println("Error deleting file:", err)
	} else {
		fmt.Println("File deleted successfully")
	}
}
