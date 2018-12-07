package csv

import (
	"os"
	"encoding/csv"
)

// CSVData is type for data who will write to csv file.
type CSVData [][]string

// Write create csv file by provides path and write provided data to this file.
func Write(path string, data CSVData) (error) {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, value := range data {
		err := writer.Write(value)
		if err != nil {
			return err
		}
	}

	return nil
}