#!/bin/bash

echo "Distributed Grep Testing"
echo "============================="

# Check if go is available
if ! command -v go &> /dev/null; then
    echo "XXXXX Go is not installed XXXXX"
    exit 1
fi

echo "######Go is available: $(go version)######"

# Run the test suite
echo ""
echo "Starting tests ..."
echo ""

if go run test_runner.go; then
    echo ""
    echo "********Testing completed successfully********"
else
    echo ""
    echo " Tests failed to execute!"
    exit 1
fi
