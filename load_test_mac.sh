#!/bin/bash

# Configuration for Mac (scaled down from 50K)
TOTAL_REQUESTS=10000  # 1K users × 10 requests each
CONCURRENT_USERS=1000
TEST_LINKS_FILE="test_links_mac.txt"

# Read test links into array (Mac compatible)
URLS=()
while IFS= read -r line; do
    URLS+=("$line")
done < "$TEST_LINKS_FILE"

NUM_URLS=${#URLS[@]}

if [ $NUM_URLS -eq 0 ]; then
    echo "❌ No test URLs found in $TEST_LINKS_FILE"
    exit 1
fi

echo "🎯 Starting load test with $NUM_URLS URLs"
echo "📊 Target: $CONCURRENT_USERS concurrent users"
echo "🔗 Testing URLs: ${URLS[@]}"

# Calculate requests per URL
REQUESTS_PER_URL=$((TOTAL_REQUESTS / NUM_URLS))
CONCURRENT_PER_URL=$((CONCURRENT_USERS / NUM_URLS))

echo "📈 $REQUESTS_PER_URL requests per URL with $CONCURRENT_PER_URL concurrent per URL"
echo "🚫 Disabling redirects to test only the link shortener performance"

# Check if hey is installed
if ! command -v hey &> /dev/null; then
    echo "❌ 'hey' load testing tool not found. Installing..."
    if command -v brew &> /dev/null; then
        brew install hey
    else
        echo "Please install hey: https://github.com/rakyll/hey"
        exit 1
    fi
fi

# Run load test against all URLs in parallel (without following redirects)
echo "🚀 Starting load test..."
START_TIME=$(date +%s)

for i in "${!URLS[@]}"; do
    url="${URLS[$i]}"
    hey -disable-redirects -n $REQUESTS_PER_URL -c $CONCURRENT_PER_URL -t 10 "$url" > "results_mactest$(printf "%02d" $((i+1))).txt" 2>&1 &
done

# Wait for all tests to complete
wait

END_TIME=$(date +%s)
DURATION_ACTUAL=$((END_TIME - START_TIME))

echo "✅ Load test completed in ${DURATION_ACTUAL}s"

# Aggregate results
echo "📊 Aggregating results..."
echo "=== LOAD TEST RESULTS (REDIRECT PERFORMANCE ONLY) ==="
echo "Test completed: $(date)"
echo "Duration: ${DURATION_ACTUAL}s"
echo ""

total_requests=0
total_302s=0

for file in results_mactest*.txt; do
    if [ -f "$file" ]; then
        echo "--- $(basename "$file" .txt) ---"
        cat "$file" | head -15
        echo ""
        
        # Extract metrics
        requests=$(grep "Total:" "$file" | awk '{print $2}' 2>/dev/null || echo "0")
        redirects=$(grep "\[302\]" "$file" | awk '{print $2}' 2>/dev/null || echo "0")
        
        total_requests=$((total_requests + requests))
        total_302s=$((total_302s + redirects))
    fi
done

echo "=== SUMMARY ==="
echo "Total Requests Sent: $total_requests"
echo "Total Successful Redirects (302s): $total_302s"
echo "Target Concurrent Users: $CONCURRENT_USERS"
echo "Duration: ${DURATION_ACTUAL}s"
if [ $DURATION_ACTUAL -gt 0 ]; then
    echo "Actual Throughput: $((total_requests / DURATION_ACTUAL)) requests/second"
fi
