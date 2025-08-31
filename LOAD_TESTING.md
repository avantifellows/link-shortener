# Load Testing Guide for Link Shortener

## Overview

Load testing setup for validating 50K concurrent users accessing the link shortener service. This guide covers infrastructure setup, test execution, and cleanup.

## Test Scenario

- **Target Load**: 50,000 concurrent users
- **Test Links**: 10-20 pre-created short links
- **Duration**: 2-3 minutes sustained load
- **Purpose**: Validate Redis caching performance and t3.medium capacity

## Infrastructure Requirements

### Load Testing Instance: c5.xlarge
- **Specs**: 4 vCPU, 8GB RAM
- **Purpose**: Generate 50K concurrent connections
- **Location**: Same AWS region as target server
- **Cost**: ~$0.17/hour (us-east-1)

### Why c5.xlarge
- **High CPU**: Required for 50K connection generation
- **Enhanced Networking**: Better bandwidth and packet rates
- **Memory**: Sufficient for concurrent connection overhead
- **Alternative**: 5 x c5.large instances for distributed testing

## Cost Estimates

| Duration | c5.xlarge Cost | Notes |
|----------|----------------|-------|
| 1 hour   | $0.17         | Setup + multiple test runs |
| 2 hours  | $0.34         | Extended testing |
| 4 hours  | $0.68         | Full day of testing |

**Recommendation**: Budget 1-2 hours ($0.17-$0.34) for complete testing cycle.

## Setup Commands

### 1. Launch Load Testing Instance

```bash
# Create c5.xlarge instance
aws ec2 run-instances \
  --image-id ami-0c7217cdde317cfec \
  --count 1 \
  --instance-type c5.xlarge \
  --key-name your-key-pair \
  --security-group-ids sg-xxxxxxxxx \
  --subnet-id subnet-xxxxxxxxx \
  --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=load-test-instance},{Key=Purpose,Value=redis-load-test}]'

# Get instance ID from output, then get public IP
aws ec2 describe-instances --instance-ids i-xxxxxxxxx --query 'Reservations[0].Instances[0].PublicIpAddress'
```

### 2. Configure Load Testing Instance

```bash
# SSH into instance
ssh -i your-key.pem ubuntu@INSTANCE_IP

# Install dependencies
sudo apt-get update
sudo apt-get install -y hey

# Increase system limits for high concurrency
echo '* soft nofile 100000' | sudo tee -a /etc/security/limits.conf
echo '* hard nofile 100000' | sudo tee -a /etc/security/limits.conf
echo 'net.core.somaxconn = 65535' | sudo tee -a /etc/sysctl.conf
echo 'net.ipv4.tcp_max_syn_backlog = 65535' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

# Logout and login again for limits to take effect
exit
ssh -i your-key.pem ubuntu@INSTANCE_IP
```

## Pre-Test Setup

### 1. Create Test Links

```bash
# Create test script to generate 10-20 short links
cat > create_test_links.sh << 'EOF'
#!/bin/bash

BASE_URL="https://lnk.avantifellows.org"
ADMIN_USER="admin"
ADMIN_PASS="your_admin_password"

# Array of test URLs
TEST_URLS=(
    "https://example.com/session1"
    "https://example.com/session2" 
    "https://example.com/session3"
    "https://example.com/session4"
    "https://example.com/session5"
    "https://example.com/session6"
    "https://example.com/session7"
    "https://example.com/session8"
    "https://example.com/session9"
    "https://example.com/session10"
    "https://example.com/session11"
    "https://example.com/session12"
    "https://example.com/session13"
    "https://example.com/session14"
    "https://example.com/session15"
    "https://example.com/session16"
    "https://example.com/session17"
    "https://example.com/session18"
    "https://example.com/session19"
    "https://example.com/session20"
)

echo "Creating test links..."
> test_links.txt

for i in "${!TEST_URLS[@]}"; do
    CUSTOM_CODE="test$(printf "%02d" $((i+1)))"
    
    RESPONSE=$(curl -s -u "$ADMIN_USER:$ADMIN_PASS" \
        -X POST "$BASE_URL/shorten" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -H "Accept: application/json" \
        -d "original_url=${TEST_URLS[$i]}&custom_code=$CUSTOM_CODE&created_by=load-test")
    
    SHORT_URL=$(echo "$RESPONSE" | jq -r '.short_url')
    echo "$SHORT_URL" >> test_links.txt
    echo "Created: $SHORT_URL -> ${TEST_URLS[$i]}"
done

echo "Test links saved to test_links.txt"
EOF

chmod +x create_test_links.sh
```

### 2. Load Test Execution Script

```bash
cat > load_test.sh << 'EOF'
#!/bin/bash

# Configuration
TOTAL_REQUESTS=500000  # 50K users × 10 requests each
CONCURRENT_USERS=50000
DURATION=180  # 3 minutes
TEST_LINKS_FILE="test_links.txt"

# Read test links into array
mapfile -t URLS < "$TEST_LINKS_FILE"
NUM_URLS=${#URLS[@]}

if [ $NUM_URLS -eq 0 ]; then
    echo "❌ No test URLs found in $TEST_LINKS_FILE"
    echo "   Run create_test_links.sh first"
    exit 1
fi

echo "🎯 Starting load test with $NUM_URLS URLs"
echo "📊 Target: $CONCURRENT_USERS concurrent users for ${DURATION}s"
echo "🔗 URLs: ${URLS[@]}"

# Calculate requests per URL
REQUESTS_PER_URL=$((TOTAL_REQUESTS / NUM_URLS))
CONCURRENT_PER_URL=$((CONCURRENT_USERS / NUM_URLS))

echo "📈 $REQUESTS_PER_URL requests per URL with $CONCURRENT_PER_URL concurrent per URL"

# Start monitoring
echo "📊 Starting monitoring..."
(
    while true; do
        echo "$(date): Load test running..."
        sleep 10
    done
) &
MONITOR_PID=$!

# Run load test against all URLs in parallel
echo "🚀 Starting load test..."
START_TIME=$(date +%s)

for url in "${URLS[@]}"; do
    hey -n $REQUESTS_PER_URL -c $CONCURRENT_PER_URL -t 30 "$url" > "results_$(basename "$url").txt" 2>&1 &
done

# Wait for all tests to complete
wait

# Stop monitoring
kill $MONITOR_PID 2>/dev/null

END_TIME=$(date +%s)
DURATION_ACTUAL=$((END_TIME - START_TIME))

echo "✅ Load test completed in ${DURATION_ACTUAL}s"

# Aggregate results
echo "📊 Aggregating results..."
cat > aggregate_results.sh << 'INNER_EOF'
#!/bin/bash
echo "=== LOAD TEST RESULTS ==="
echo "Test completed: $(date)"
echo "Duration: ${DURATION_ACTUAL}s"
echo ""

total_requests=0
total_success=0
total_errors=0

for file in results_*.txt; do
    if [ -f "$file" ]; then
        echo "--- $(basename "$file" .txt) ---"
        
        # Extract key metrics
        requests=$(grep "Total:" "$file" | awk '{print $2}' || echo "0")
        success=$(grep "Success:" "$file" | awk '{print $2}' || echo "0") 
        errors=$(grep -E "(Error|Failed)" "$file" | wc -l || echo "0")
        
        echo "Requests: $requests"
        echo "Success: $success"  
        echo "Errors: $errors"
        echo ""
        
        total_requests=$((total_requests + requests))
        total_success=$((total_success + success))
        total_errors=$((total_errors + errors))
    fi
done

echo "=== TOTALS ==="
echo "Total Requests: $total_requests"
echo "Total Success: $total_success"
echo "Total Errors: $total_errors"
success_rate=$(( total_success * 100 / total_requests ))
echo "Success Rate: ${success_rate}%"
INNER_EOF

chmod +x aggregate_results.sh
./aggregate_results.sh
EOF

chmod +x load_test.sh
```

## Machine Requirements

**Your Mac**: ❌ Cannot generate 50K concurrent connections
- File descriptor limits (~16K)
- Network stack limitations
- ISP bandwidth constraints

**c5.xlarge**: ✅ Perfect for this load
- **4 vCPU**: Handle connection overhead
- **Enhanced networking**: High packet rates
- **Same region**: Realistic network latency

**Cost for testing session**: ~$0.50 (2-3 hours including setup)

The key insight: Load testing at this scale requires infrastructure designed for high concurrency - your local machine isn't built for this.