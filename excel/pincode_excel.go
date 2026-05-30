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

// cleanAreaName strips trailing B.O, S.O, BO, SO suffixes and cleans up names
func cleanAreaName(name string) string {
	name = strings.TrimSpace(name)
	upperName := strings.ToUpper(name)
	suffixes := []string{" B.O.", " B.O", " BO", " S.O.", " S.O", " SO", " H.O.", " H.O", " HO"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(upperName, suffix) {
			name = name[:len(name)-len(suffix)]
			break
		}
	}
	return strings.TrimSpace(name)
}

// uniqueStrings removes duplicate entries from a slice of strings
func uniqueStrings(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if entry == "" {
			continue
		}
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

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

		pincodeMap := make(map[uint32]*models.Pincode)
		areaMap := make(map[uint32][]string)
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

			pin := uint32(excelPincode)
			officeName := cleanAreaName(row[headerMap["OfficeName"]])

			if _, exists := pincodeMap[pin]; !exists {
				city := strings.TrimSpace(strings.TrimSuffix(row[headerMap["DivisionName"]], " Division"))
				pincodeMap[pin] = &models.Pincode{
					Pincode:  pin,
					City:     city,
					District: cleanAreaName(row[headerMap["District"]]),
					State:    row[headerMap["StateName"]],
				}
			}
			areaMap[pin] = append(areaMap[pin], officeName)
			count++
		}

		// Group and serialize areas
		loadedCount := 0
		for pin, pinModel := range pincodeMap {
			areas := areaMap[pin]
			unique := uniqueStrings(areas)
			pinModel.Areas = strings.Join(unique, ", ")
			csvMap.Store(pin, pinModel)
			loadedCount++
		}

		isLoaded = true
		log.Printf("Successfully loaded %d records grouped into %d unique pincodes from CSV.\n", count, loadedCount)
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

	// 2. Fallback to memory-efficient streaming search accumulating all areas
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

	var matchedPincode *models.Pincode
	var areas []string

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
			if matchedPincode == nil {
				city := strings.TrimSpace(strings.TrimSuffix(row[headerMap["DivisionName"]], " Division"))
				matchedPincode = &models.Pincode{
					Pincode:  pincode,
					City:     city,
					District: cleanAreaName(row[headerMap["District"]]),
					State:    row[headerMap["StateName"]],
				}
			}
			areas = append(areas, cleanAreaName(row[headerMap["OfficeName"]]))
		}
	}

	if matchedPincode != nil {
		matchedPincode.Areas = strings.Join(uniqueStrings(areas), ", ")
		return matchedPincode, nil
	}

	return nil, errors.New("pincode not found in excel")
}

