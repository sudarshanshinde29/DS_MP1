package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// TestCase represents a simple test case
type TestCase struct {
	Name          string
	GrepArgs      []string
	Mode          string
	ExpectedCount int
	ExpectedPerVM map[string]int // Expected count per VM
	Description   string
}

// TestResult represents the result of a test case
type TestResult struct {
	TestCase     TestCase
	ActualCount  int
	ActualPerVM  map[string]int // Actual count per VM
	Passed       bool
	ErrorMessage string
	Duration     time.Duration
}

// GetTestCases returns our test cases
func GetTestCases() []TestCase {
	return []TestCase{
		{
			Name:          "Frequent Pattern - GET",
			GrepArgs:      []string{"-i", "-e", "GET"},
			Mode:          "count",
			ExpectedCount: 450, // 40 + 5 + 3 per VM × 10 VMs
			ExpectedPerVM: map[string]int{
				"vm1": 48, "vm2": 48, "vm3": 48, "vm4": 48, "vm5": 48,
				"vm6": 48, "vm7": 48, "vm8": 48, "vm9": 48, "vm10": 48,
			},
			Description: "Test frequent pattern matching",
		},
		{
			Name:          "Somewhat Frequent - PUT",
			GrepArgs:      []string{"-i", "-e", "PUT"},
			Mode:          "count",
			ExpectedCount: 250, // 20 + 5 per VM × 10 VMs
			ExpectedPerVM: map[string]int{
				"vm1": 25, "vm2": 25, "vm3": 25, "vm4": 25, "vm5": 25,
				"vm6": 25, "vm7": 25, "vm8": 25, "vm9": 25, "vm10": 25,
			},
			Description: "Test somewhat frequent pattern",
		},
		{
			Name:          "Rare Pattern - DELETE",
			GrepArgs:      []string{"-i", "-e", "DELETE"},
			Mode:          "count",
			ExpectedCount: 50, // 5 per VM × 10 VMs
			ExpectedPerVM: map[string]int{
				"vm1": 5, "vm2": 5, "vm3": 5, "vm4": 5, "vm5": 5,
				"vm6": 5, "vm7": 5, "vm8": 5, "vm9": 5, "vm10": 5,
			},
			Description: "Test rare pattern matching",
		},
		{
			Name:          "Regex Pattern - Status Codes",
			GrepArgs:      []string{"-E", "-e", "200|201"},
			Mode:          "count",
			ExpectedCount: 130, // 5+5+3 per VM × 10 VMs
			ExpectedPerVM: map[string]int{
				"vm1": 13, "vm2": 13, "vm3": 13, "vm4": 13, "vm5": 13,
				"vm6": 13, "vm7": 13, "vm8": 13, "vm9": 13, "vm10": 13,
			},
			Description: "Test regex with status codes",
		},
		{
			Name:          "Regex Pattern - HTTP Paths",
			GrepArgs:      []string{"-E", "-e", "/api/users|/api/login"},
			Mode:          "count",
			ExpectedCount: 100, // 5+5 per VM × 10 VMs
			ExpectedPerVM: map[string]int{
				"vm1": 10, "vm2": 10, "vm3": 10, "vm4": 10, "vm5": 10,
				"vm6": 10, "vm7": 10, "vm8": 10, "vm9": 10, "vm10": 10,
			},
			Description: "Test regex with HTTP paths",
		},
		{
			Name:          "VM1 Only Pattern",
			GrepArgs:      []string{"-F", "-e", "VM1_UNIQUE_PATTERN"},
			Mode:          "count",
			ExpectedCount: 1, // Only in VM1
			ExpectedPerVM: map[string]int{
				"vm1": 1, "vm2": 0, "vm3": 0, "vm4": 0, "vm5": 0,
				"vm6": 0, "vm7": 0, "vm8": 0, "vm9": 0, "vm10": 0,
			},
			Description: "Pattern should only appear in VM1",
		},
		{
			Name:          "Non-existent Pattern",
			GrepArgs:      []string{"-i", "-e", "NONEXISTENT"},
			Mode:          "count",
			ExpectedCount: 0,
			ExpectedPerVM: map[string]int{
				"vm1": 0, "vm2": 0, "vm3": 0, "vm4": 0, "vm5": 0,
				"vm6": 0, "vm7": 0, "vm8": 0, "vm9": 0, "vm10": 0,
			},
			Description: "Pattern should not exist anywhere",
		},
	}
}

// RunTestCase executes a single test case
func RunTestCase(testCase TestCase) (*TestResult, error) {
	fmt.Printf(" Running: %s\n", testCase.Name)
	fmt.Printf(" Expected: %d\n", testCase.ExpectedCount)

	startTime := time.Now()

	// Build coordinator command
	args := []string{
		"run", "../coordinator/main.go",
		"-props", "../cluster.properties",
		"-mode", testCase.Mode,
		"--",
	}
	args = append(args, testCase.GrepArgs...)

	// Execute coordinator
	cmd := exec.Command("go", args...)
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	if err != nil {
		return &TestResult{
			TestCase:     testCase,
			Passed:       false,
			ErrorMessage: fmt.Sprintf("Command failed: %v", err),
			Duration:     duration,
		}, nil
	}

	// Parse output to get count
	result := &TestResult{
		TestCase:    testCase,
		Duration:    duration,
		ActualPerVM: make(map[string]int),
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Look for VM count results: [vm1] count=123
		if strings.Contains(line, "] count=") {
			parts := strings.Split(line, "] count=")
			if len(parts) == 2 {
				vmName := strings.TrimPrefix(parts[0], "[")
				if count, err := strconv.Atoi(parts[1]); err == nil {
					result.ActualPerVM[vmName] = count
					result.ActualCount += count
				}
			}
		}

		// Look for total count: TOTAL_COUNT=123
		if strings.HasPrefix(line, "TOTAL_COUNT=") {
			if count, err := strconv.Atoi(strings.TrimPrefix(line, "TOTAL_COUNT=")); err == nil {
				result.ActualCount = count
			}
		}
	}

	// Validate result
	if result.ActualCount == testCase.ExpectedCount {
		result.Passed = true
	} else {
		result.Passed = false
		result.ErrorMessage = fmt.Sprintf("Expected %d, got %d",
			testCase.ExpectedCount, result.ActualCount)
	}

	// Also validate per-VM results
	if result.Passed {
		for vm, expected := range testCase.ExpectedPerVM {
			if actual, exists := result.ActualPerVM[vm]; !exists || actual != expected {
				result.Passed = false
				result.ErrorMessage = fmt.Sprintf("VM %s: expected %d, got %d",
					vm, expected, actual)
				break
			}
		}
	}

	if result.Passed {
		fmt.Printf("******** PASSED: Found %d (expected %d)*********\n", result.ActualCount, testCase.ExpectedCount)
	} else {
		fmt.Printf("XXXXXXXXXFAILED: %s\n", result.ErrorMessage)
	}

	return result, nil
}

// RunAllTests executes all test cases
func RunAllTests() ([]*TestResult, error) {
	testCases := GetTestCases()
	var results []*TestResult

	fmt.Printf("#########Running %d test cases\n", len(testCases))
	fmt.Println(strings.Repeat("=", 50))

	for i, testCase := range testCases {
		fmt.Printf("Test %d/%d: %s\n", i+1, len(testCases), testCase.Name)

		result, err := RunTestCase(testCase)
		if err != nil {
			return nil, fmt.Errorf("test case %s failed: %v", testCase.Name, err)
		}

		results = append(results, result)
		fmt.Println()
	}

	return results, nil
}

// GenerateTestReport creates a test report
func GenerateTestReport(results []*TestResult) {
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("........ TEST REPORT.........")
	fmt.Println(strings.Repeat("=", 50))

	totalTests := len(results)
	passedTests := 0
	failedTests := 0

	for _, result := range results {
		if result.Passed {
			passedTests++
		} else {
			failedTests++
		}
	}

	fmt.Printf("Total Tests: %d\n", totalTests)
	fmt.Printf("Passed: %d\n", passedTests)
	fmt.Printf("Failed: %d\n", failedTests)

	if failedTests > 0 {
		fmt.Println("\n Failed Tests:")
		for _, result := range results {
			if !result.Passed {
				fmt.Printf("   • %s: %s\n", result.TestCase.Name, result.ErrorMessage)
			}
		}
	}

	fmt.Println(strings.Repeat("=", 50))
	if failedTests == 0 {
		fmt.Println("Successfully All tests passed!")
	} else {
		fmt.Printf("  %d test(s) failed\n", failedTests)
	}
	fmt.Println(strings.Repeat("=", 50))
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "cleanup" {
		fmt.Println("🧹 Cleaning up test data...")
		cmd := exec.Command("./generate_test_data.sh", "cleanup")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatalf("Cleanup failed: %v", err)
		}
		return
	}

	fmt.Println("Starting Distributed Grep Testing")
	fmt.Println(strings.Repeat("=", 60))

	// Step 1: Generate test data on all VMs
	fmt.Println("First, generating test data on all VMs...")
	cmd := exec.Command("./generate_test_data.sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to generate test data: %v", err)
	}

	// Step 2: Start workers on all VMs
	fmt.Println("Starting workers on all VMs...")
	cmd = exec.Command("./start_test_workers.sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to start workers: %v", err)
	}

	// Step 3: Wait for workers to be ready
	fmt.Println("Waiting for workers to be ready...")
	time.Sleep(5 * time.Second)

	// Step 4: Run tests
	fmt.Println("@@@@@@@Running test cases...@@@@@@@")
	results, err := RunAllTests()
	if err != nil {
		log.Fatalf("Test execution failed: %v", err)
	}

	// Step 5: Generate report
	fmt.Println("Generating test report...")
	GenerateTestReport(results)

	fmt.Println("All Test end to end completed!")
}
