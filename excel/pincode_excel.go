package excel

import (
	"encoding/csv"
	"errors"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"embed"

	"github.com/19481A1281/go-pincode-service/models"
)

//go:embed indian_pincodes_option2.csv
var pincodeFile embed.FS

var (
	csvMap       sync.Map
	isLoaded     bool
	loadingMutex sync.RWMutex
)

// InitCSVData loads the CSV file in a background goroutine on startup
func InitCSVData() {
	go func() {
		loadingMutex.Lock()
		defer loadingMutex.Unlock()
		if isLoaded {
			return
		}

		log.Println("Starting background load of pincode CSV data...")
		file, err := pincodeFile.Open("indian_pincodes_option2.csv")
		if err != nil {
			log.Println("Error opening embedded CSV file:", err)
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)
		header, err := reader.Read()
		if err != nil {
			log.Println("Error reading CSV header:", err)
			return
		}

		headerMap := map[string]int{}
		for index, column := range header {
			headerMap[column] = index
		}

		count := 0
		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				continue
			}

			if len(row) <= headerMap["Pincode"] {
				continue
			}

			excelPincode, err := strconv.ParseUint(row[headerMap["Pincode"]], 10, 32)
			if err != nil {
				continue
			}

			city := strings.TrimSpace(
				strings.TrimSuffix(
					row[headerMap["DivisionName"]],
					" Division",
				),
			)

			data := models.Pincode{
				Pincode:  uint32(excelPincode),
				City:     city,
				District: row[headerMap["District"]],
				State:    row[headerMap["StateName"]],
			}

			csvMap.Store(uint32(excelPincode), &data)
			count++
		}

		isLoaded = true
		log.Printf("Successfully loaded %d pincode records from CSV into memory map.\n", count)
	}()
}

// SearchPincodeInExcel searches for a pincode in the memory map, falling back to line-by-line streaming search
func SearchPincodeInExcel(
	pincode uint32,
) (*models.Pincode, error) {

	loadingMutex.RLock()
	loaded := isLoaded
	loadingMutex.RUnlock()

	// 1. O(1) Memory map lookup if loaded
	if loaded {
		if val, ok := csvMap.Load(pincode); ok {
			return val.(*models.Pincode), nil
		}
		return nil, errors.New("pincode not found in excel")
	}

	// 2. Fallback to memory-efficient streaming search
	file, err := pincodeFile.Open("indian_pincodes_option2.csv")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}

	headerMap := map[string]int{}
	for index, column := range header {
		headerMap[column] = index
	}

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		if len(row) <= headerMap["Pincode"] {
			continue
		}

		excelPincode, err := strconv.ParseUint(row[headerMap["Pincode"]], 10, 32)
		if err != nil {
			continue
		}

		if uint32(excelPincode) == pincode {
			city := strings.TrimSpace(
				strings.TrimSuffix(
					row[headerMap["DivisionName"]],
					" Division",
				),
			)

			data := models.Pincode{
				Pincode:  uint32(excelPincode),
				City:     city,
				District: row[headerMap["District"]],
				State:    row[headerMap["StateName"]],
			}

			return &data, nil
		}
	}

	return nil, errors.New("pincode not found in excel")
}

