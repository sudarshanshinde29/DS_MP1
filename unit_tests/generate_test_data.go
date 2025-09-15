package main

import (
	"fmt"
	"os"
)

func generateTestData(machineIndex string) {
	// Define simple pattern distribution
	patterns := map[string]int{
		"GET":    40,
		"PUT":    20,
		"POST":   15,
		"DELETE": 5,
		"ERROR":  10,
		"INFO":   10,
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll("/root/generated_logs", 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	// Create log file
	filename := "/root/generated_logs/vm" + machineIndex + ".log"
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", filename, err)
		return
	}
	defer file.Close()

	fmt.Printf("Generating test data for VM%s: %s\n", machineIndex, filename)

	// some entries for each pattern
	for pattern, count := range patterns {
		for i := 0; i < count; i++ {
			fmt.Fprintf(file, "%s request %d\n", pattern, i+1)
		}
	}

	// Adding some regex-friendly patterns
	for i := 0; i < 5; i++ {
		fmt.Fprintf(file, "GET /api/users HTTP/1.1 200 %d\n", 1000+i)
		fmt.Fprintf(file, "PUT /api/users HTTP/1.1 201 %d\n", 2000+i)
		fmt.Fprintf(file, "POST /api/login HTTP/1.1 200 %d\n", 3000+i)
	}

	// Adding some status code patterns for regex testing
	for i := 0; i < 3; i++ {
		fmt.Fprintf(file, "GET /api/data HTTP/1.1 200 %d\n", 4000+i)
		fmt.Fprintf(file, "GET /api/data HTTP/1.1 404 %d\n", 5000+i)
		fmt.Fprintf(file, "GET /api/data HTTP/1.1 500 %d\n", 6000+i)
	}

	// Add unique patterns for this VM
	fmt.Fprintf(file, "VM%s_UNIQUE_PATTERN\n", machineIndex)
	fmt.Fprintf(file, "VM%s_SPECIFIC_ERROR\n", machineIndex)

	fmt.Printf("Generated test data for VM%s\n", machineIndex)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("run using :  go run generate_test_data.go <machine_index>")
		fmt.Println("ex: go run generate_test_data.go 1")
		os.Exit(1)
	}

	machineIndex := os.Args[1]
	generateTestData(machineIndex)
}
