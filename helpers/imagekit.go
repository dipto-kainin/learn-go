package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

type ImageKitUploadResponse struct {
	FileID       string `json:"fileId"`
	Name         string `json:"name"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnailUrl"`
}

// UploadToImageKit uploads a file (binary bytes) to ImageKit.io
func UploadToImageKit(fileBytes []byte, fileName string) (*ImageKitUploadResponse, error) {
	endpoint := os.Getenv("IMAGEKIT_IO_ENDPOINT")
	privateKey := os.Getenv("IMAGEKIT_IO_SK_KEY")

	if endpoint == "" || privateKey == "" {
		return nil, fmt.Errorf("ImageKit environment variables are not set")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file field
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, err
	}
	_, err = part.Write(fileBytes)
	if err != nil {
		return nil, err
	}

	// Add fileName field
	err = writer.WriteField("fileName", fileName)
	if err != nil {
		return nil, err
	}

	// Add useUniqueFileName field (false to overwrite if filename matches)
	err = writer.WriteField("useUniqueFileName", "false")
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://upload.imagekit.io/api/v1/files/upload", body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetBasicAuth(privateKey, "")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("ImageKit upload failed with status %d: %s", resp.StatusCode, string(respBytes))
	}

	var uploadResp ImageKitUploadResponse
	err = json.Unmarshal(respBytes, &uploadResp)
	if err != nil {
		return nil, err
	}

	return &uploadResp, nil
}



// DeleteFromImageKit deletes a file from ImageKit.io by its fileId
func DeleteFromImageKit(fileID string) error {
	privateKey := os.Getenv("IMAGEKIT_IO_SK_KEY")
	if privateKey == "" {
		return fmt.Errorf("ImageKit private key is not set")
	}

	if fileID == "" {
		return nil // Nothing to delete
	}

	url := fmt.Sprintf("https://api.imagekit.io/v1/files/%s", fileID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(privateKey, "")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		respBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ImageKit delete failed with status %d: %s", resp.StatusCode, string(respBytes))
	}

	return nil
}
