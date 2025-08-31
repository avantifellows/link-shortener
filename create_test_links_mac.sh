#!/bin/bash

BASE_URL="https://lnk.avantifellows.org"
AUTH_TOKEN="8E8E3C30-19B6-4DCB-AEC0-84F8CB9485FF"

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
)

echo "Creating test links..."
> test_links_mac.txt

for i in "${!TEST_URLS[@]}"; do
    CUSTOM_CODE="test$(printf "%02d" $((i+1)))"
    
    RESPONSE=$(curl -s -X POST "$BASE_URL/shorten" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -d "original_url=${TEST_URLS[$i]}&custom_code=$CUSTOM_CODE&created_by=load-test")
    
    echo "$RESPONSE"
    if [[ "$RESPONSE" == *"$BASE_URL"* ]]; then
        echo "$BASE_URL/$CUSTOM_CODE" >> test_links_mac.txt
        echo "Created: $BASE_URL/$CUSTOM_CODE -> ${TEST_URLS[$i]}"
    else
        echo "Failed to create link for ${TEST_URLS[$i]}: $RESPONSE"
    fi
done

echo "Test links saved to test_links_mac.txt"
