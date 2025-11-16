package main

import (
	"fmt"
	"log"

	"github.com/xuri/excelize/v2"
)

func main() {
	f, err := excelize.OpenFile("/home/lcgerke/cuSched.ods")
	if err != nil {
		log.Fatalf("Failed to open ODS file: %v", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	fmt.Printf("Sheets found: %d\n", len(sheets))
	for i, sheet := range sheets {
		fmt.Printf("  %d: %s\n", i+1, sheet)
	}

	// Inspect first sheet
	if len(sheets) > 0 {
		rows, err := f.GetRows(sheets[0])
		if err != nil {
			log.Fatalf("Failed to get rows: %v", err)
		}

		fmt.Printf("\nFirst sheet (%s) has %d rows\n", sheets[0], len(rows))
		if len(rows) > 0 {
			fmt.Printf("Headers: %v\n", rows[0])
			if len(rows) > 1 {
				fmt.Printf("First data row: %v\n", rows[1])
			}
		}
	}
}
