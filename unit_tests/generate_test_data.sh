#!/bin/bash

# variables
REPO_DIR="/root/MP/DS_MP1"
VM_HOSTNAME="fa25-cs425-10"
COUNT="${COUNT:-10}"

# Function to generate test data on a remote host
generate_test_data() {
    local n=$1
    local host="${VM_HOSTNAME}${n}.cs.illinois.edu"
    echo "Connecting to $host (VM$n)..."
    
    ssh -o BatchMode=yes -o StrictHostKeyChecking=accept-new root@"$host" << EOF
        echo "Connected to $host"
        cd $REPO_DIR
        echo "Changed directory to $REPO_DIR"
        
        # Create logs directory if it doesn't exist
        mkdir -p /root/generated_logs
        
        echo "Generating test data for VM$n..."
        go run unit_tests/generate_test_data.go $n
        
        echo "Test data generation completed for VM$n"
EOF
    echo "Disconnected from $host"
    echo "------------------------"
}

# Function to cleanup test data on a remote host
cleanup_test_data() {
    local host="$1"
    local n="$2"
    local host="${VM_HOSTNAME}${n}.cs.illinois.edu"
    echo "Cleaning up test data on $host (VM$n)..."
    
    ssh -o BatchMode=yes -o StrictHostKeyChecking=accept-new root@"$host" << EOF
        echo "Connected to $host"
        rm -f /root/generated_logs/vm$n.log
        echo "Cleaned up test data for VM$n"
EOF
    echo "Disconnected from $host"
    echo "------------------------"
}

# Main execution
if [ "$1" = "cleanup" ]; then
    echo "🧹 Cleaning up test data on all VMs..."
    for n in $(seq 1 $COUNT); do
        cleanup_test_data $n &
    done
    wait
    echo "Cleanup completed on all VMs"
else
    echo "OK, data generation script , Generating test data on all VMs..."
    for XX in $(seq -w 01 "$COUNT"); do
      host="${VM_HOSTNAME}${XX}.cs.illinois.edu"
      n=$((10#$XX))  # 01->1
      generate_test_data "$host" "$n" &
    done
    wait
    echo "Test data generation completed on all VMs"
fi
