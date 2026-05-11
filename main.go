package main

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Simplified structures for Base64 only
type PinMediaSource struct {
	SourceType  string `json:"source_type"`
	Data        string `json:"data"`
	ContentType string `json:"content_type"`
}

type PinRequest struct {
	BoardID        string         `json:"board_id"`
	BoardSectionID string         `json:"board_section_id,omitempty"`
	Title          string         `json:"title,omitempty"`
	Description    string         `json:"description,omitempty"`
	Link           string         `json:"link,omitempty"`
	AltText        string         `json:"alt_text,omitempty"`
	Note           string         `json:"note,omitempty"`
	MediaSource    PinMediaSource `json:"media_source"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

// CSV row structure for batch pin creation
type PinData struct {
	FilePath    string
	Title       string
	Description string
	Link        string
	AltText     string
	SectionID   string
	Note        string
}

func main() {
	// Check if this is being called as image validator
	if len(os.Args) > 1 && os.Args[1] == "check-images" {
		if len(os.Args) < 3 {
			log.Fatal("Usage: go run main.go check-images <csv_file>")
		}
		checkImagesInCSV(os.Args[2])
		return
	}

	// 1. Gather Environment Variables
	appID := os.Getenv("PINTEREST_APP_ID")
	appSecret := os.Getenv("PINTEREST_APP_SECRET")
	refreshToken := os.Getenv("PINTEREST_REFRESH_TOKEN")
	boardID := os.Getenv("PINTEREST_BOARD_ID")

	if appID == "" || refreshToken == "" {
		log.Fatal("❌ Missing critical secrets (App ID or Refresh Token)")
	}

	if boardID == "" {
		log.Fatal("❌ Missing Pinterest Board ID")
	}

	// 2. Check for CSV batch processing or single pin
	csvPath := os.Getenv("INPUT_CSV_PATH")

	log.Println("🔄 Authenticating...")
	token := getAccessToken(appID, appSecret, refreshToken)

	if csvPath != "" {
		// Batch processing from CSV
		log.Printf("📂 Processing CSV file: %s", csvPath)
		processBatchPins(token, boardID, csvPath)
	} else {
		// Single pin processing (legacy mode)
		log.Println("📌 Processing single pin...")
		processSinglePin(token, boardID)
	}
}

func processSinglePin(token, boardID string) {
	// Gather single pin inputs
	filePath := os.Getenv("INPUT_FILE_PATH")
	title := os.Getenv("INPUT_TITLE")
	desc := os.Getenv("INPUT_DESCRIPTION")
	link := os.Getenv("INPUT_LINK")
	altText := os.Getenv("INPUT_ALT_TEXT")
	sectionID := os.Getenv("INPUT_SECTION_ID")
	note := os.Getenv("INPUT_NOTE")

	if filePath == "" {
		log.Fatal("❌ INPUT_FILE_PATH is required for single pin creation")
	}

	pinData := PinData{
		FilePath:    filePath,
		Title:       title,
		Description: desc,
		Link:        link,
		AltText:     altText,
		SectionID:   sectionID,
		Note:        note,
	}

	err := createPinFromData(token, boardID, pinData, 1, 1)
	if err != nil {
		log.Fatalf("❌ Failed to create pin: %v", err)
	}
}

func processBatchPins(token, boardID, csvPath string) {
	pins, err := readPinsFromCSV(csvPath)
	if err != nil {
		log.Fatalf("❌ Error reading CSV file: %v", err)
	}

	if len(pins) == 0 {
		log.Fatal("❌ No pins found in CSV file")
	}

	log.Printf("📊 Found %d pins to process from batch: %s", len(pins), filepath.Base(csvPath))
	log.Printf("🎯 All pins will defautlt link here unless otherwise specified: https://www.loveofsalt.com (default)")

	successCount := 0
	failCount := 0

	for i, pin := range pins {
		title := pin.Title
		if title == "" {
			title = filepath.Base(pin.FilePath)
		}
		log.Printf("🔄 Processing pin %d/%d: %s", i+1, len(pins), title)

		err := createPinFromData(token, boardID, pin, i+1, len(pins))
		if err != nil {
			log.Printf("❌ Failed to create pin %d (%s): %v", i+1, title, err)
			failCount++
		} else {
			successCount++
		}
	}

	log.Printf("✅ Batch processing complete! Success: %d, Failed: %d", successCount, failCount)

	if failCount > 0 {
		log.Printf("⚠️  Some pins failed. Check logs above for details.")
		if successCount == 0 {
			log.Fatal("❌ All pins failed - batch processing unsuccessful")
		}
	}

	log.Printf("🎉 Batch %s processed successfully!", filepath.Base(csvPath))
}

func readPinsFromCSV(csvPath string) ([]PinData, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// Check if first row is header (contains "file_path" or "title")
	hasHeader := false
	if len(records) > 0 {
		firstRow := strings.ToLower(strings.Join(records[0], "|"))
		if strings.Contains(firstRow, "file_path") || strings.Contains(firstRow, "title") {
			hasHeader = true
		}
	}

	startIdx := 0
	if hasHeader {
		startIdx = 1
		log.Println("📋 CSV header detected, skipping first row")
	}

	var pins []PinData
	for i := startIdx; i < len(records); i++ {
		row := records[i]

		// Skip empty rows
		if len(row) == 0 || (len(row) == 1 && row[0] == "") {
			continue
		}

		// Ensure we have at least the file_path column
		if len(row) < 1 || row[0] == "" {
			log.Printf("⚠️ Skipping row %d: missing file_path", i+1)
			continue
		}

		pin := PinData{
			FilePath: row[0],
		}

		// Optional columns (with safe indexing)
		if len(row) > 1 {
			pin.Title = row[1]
		}
		if len(row) > 2 {
			pin.Description = row[2]
		}
		if len(row) > 3 {
			pin.Link = row[3]
		}
		if len(row) > 4 {
			pin.AltText = row[4]
		}
		if len(row) > 5 {
			pin.SectionID = row[5]
		}
		if len(row) > 6 {
			pin.Note = row[6]
		}

		pins = append(pins, pin)
	}

	return pins, nil
}

func createPinFromData(token, boardID string, pinData PinData, current, total int) error {
	// Read and process the image file
	fileBytes, err := os.ReadFile(pinData.FilePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", pinData.FilePath, err)
	}

	// Detect Content-Type
	mimeType := http.DetectContentType(fileBytes)
	if mimeType != "image/jpeg" && mimeType != "image/png" {
		return fmt.Errorf("invalid file type %s for %s: Pinterest only accepts jpg or png", mimeType, pinData.FilePath)
	}

	// Encode to Base64
	base64Str := base64.StdEncoding.EncodeToString(fileBytes)

	// Set default link if not provided
	link := pinData.Link
	if link == "" {
		link = "https://www.loveofsalt.com"
	}

	// Create the pin
	err = createPin(token, boardID, pinData.SectionID, pinData.Title, pinData.Description,
		link, pinData.AltText, pinData.Note, base64Str, mimeType)
	if err != nil {
		return err
	}

	log.Printf("✅ Pin %d/%d created successfully: %s", current, total, filepath.Base(pinData.FilePath))
	return nil
}

func getAccessToken(clientID, clientSecret, refreshToken string) string {
	auth := clientID + ":" + clientSecret
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	req, _ := http.NewRequest("POST", "https://api.pinterest.com/v5/oauth/token", strings.NewReader(data.Encode()))
	req.Header.Add("Authorization", "Basic "+encodedAuth)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Network error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("❌ API Auth Error: %s", string(body))
	}

	var tokenResp TokenResponse
	json.NewDecoder(resp.Body).Decode(&tokenResp)
	return tokenResp.AccessToken
}

func createPin(token, boardID, sectionID, title, desc, link, altText, note, base64Data, contentType string) error {
	payload := PinRequest{
		BoardID:        boardID,
		BoardSectionID: sectionID,
		Title:          title,
		Description:    desc,
		Link:           link,
		AltText:        altText,
		Note:           note,
		MediaSource: PinMediaSource{
			SourceType:  "image_base64",
			Data:        base64Data,
			ContentType: contentType,
		},
	}

	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "https://api.pinterest.com/v5/pins", bytes.NewBuffer(jsonData))
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error creating pin (Status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func checkImagesInCSV(csvPath string) {
	file, err := os.Open(csvPath)
	if err != nil {
		log.Fatalf("Failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}

	if len(records) == 0 {
		log.Fatal("CSV file is empty")
	}

	// Skip header if it exists
	startIdx := 0
	if len(records) > 0 {
		firstRow := strings.ToLower(strings.Join(records[0], "|"))
		if strings.Contains(firstRow, "file_path") || strings.Contains(firstRow, "title") {
			startIdx = 1
		}
	}

	allExist := true
	for i := startIdx; i < len(records); i++ {
		row := records[i]
		if len(row) == 0 || row[0] == "" {
			continue
		}

		filePath := row[0]
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Printf("❌ Missing file: %s\n", filePath)
			allExist = false
		} else {
			fmt.Printf("✅ Found: %s\n", filePath)
		}
	}

	if !allExist {
		os.Exit(1)
	}

	fmt.Println("🎉 All image files found!")
}
