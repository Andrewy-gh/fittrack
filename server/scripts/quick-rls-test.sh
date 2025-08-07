#!/bin/bash

# Quick RLS Test Script
# Tests RLS against an existing staging/production environment
# Usage: ./quick-rls-test.sh [staging-url]

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Configuration
STAGING_URL="${1:-${STAGING_URL:-https://fittrack.fly.dev}}"
USER_A_TOKEN="${USER_A_TOKEN}"
USER_B_TOKEN="${USER_B_TOKEN}"

echo "========================================="
echo "     Quick RLS Test - User Isolation"
echo "========================================="
echo "Testing: $STAGING_URL"
echo

# Check if tokens are provided
if [ -z "$USER_A_TOKEN" ] || [ -z "$USER_B_TOKEN" ]; then
    echo -e "${RED}ERROR:${NC} Missing test user tokens!"
    echo "Please set environment variables:"
    echo "  export USER_A_TOKEN='your-test-user-a-jwt'"
    echo "  export USER_B_TOKEN='your-test-user-b-jwt'"
    echo
    echo "To get tokens:"
    echo "1. Visit your Stack Auth dashboard"
    echo "2. Create/login as two different test users"
    echo "3. Extract JWT tokens from browser dev tools or API"
    exit 1
fi

# Test function
test_api() {
    local method=$1
    local endpoint=$2
    local token=$3
    local data=$4
    
    if [ -n "$data" ]; then
        curl -s -X "$method" \
            -H "x-stack-access-token: $token" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "${STAGING_URL}${endpoint}"
    else
        curl -s -X "$method" \
            -H "x-stack-access-token: $token" \
            "${STAGING_URL}${endpoint}"
    fi
}

echo "🧪 Running RLS isolation tests..."
echo

# Test 1: Create workout for User A
echo -n "Test 1: User A creates workout... "
user_a_data='{
  "date": "'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'",
  "notes": "User A test workout",
  "exercises": [{
    "name": "Test Exercise A",
    "sets": [{"reps": 10, "weight": 100, "set_type": "working"}]
  }]
}'

if test_api "POST" "/api/workouts" "$USER_A_TOKEN" "$user_a_data" >/dev/null 2>&1; then
    echo -e "${GREEN}✓ Success${NC}"
else
    echo -e "${RED}✗ Failed${NC}"
    exit 1
fi

# Test 2: Create workout for User B
echo -n "Test 2: User B creates workout... "
user_b_data='{
  "date": "'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'",
  "notes": "User B test workout", 
  "exercises": [{
    "name": "Test Exercise B",
    "sets": [{"reps": 8, "weight": 150, "set_type": "working"}]
  }]
}'

if test_api "POST" "/api/workouts" "$USER_B_TOKEN" "$user_b_data" >/dev/null 2>&1; then
    echo -e "${GREEN}✓ Success${NC}"
else
    echo -e "${RED}✗ Failed${NC}"
    exit 1
fi

# Test 3: Check User A can see their workouts
echo -n "Test 3: User A retrieves workouts... "
user_a_workouts=$(test_api "GET" "/api/workouts" "$USER_A_TOKEN")
user_a_count=$(echo "$user_a_workouts" | jq '. | length' 2>/dev/null || echo "0")

if [ "$user_a_count" -gt 0 ]; then
    echo -e "${GREEN}✓ Found $user_a_count workout(s)${NC}"
else
    echo -e "${RED}✗ No workouts found${NC}"
    exit 1
fi

# Test 4: Check User B can see their workouts
echo -n "Test 4: User B retrieves workouts... "
user_b_workouts=$(test_api "GET" "/api/workouts" "$USER_B_TOKEN")
user_b_count=$(echo "$user_b_workouts" | jq '. | length' 2>/dev/null || echo "0")

if [ "$user_b_count" -gt 0 ]; then
    echo -e "${GREEN}✓ Found $user_b_count workout(s)${NC}"
else
    echo -e "${RED}✗ No workouts found${NC}"
    exit 1
fi

# Test 5: Critical - Check if User B's results contain User A's data
echo -n "Test 5: Checking User B cannot see User A's data... "
if echo "$user_b_workouts" | grep -q "User A test workout"; then
    echo -e "${RED}✗ CRITICAL: User B can see User A's workout!${NC}"
    echo "RLS IS NOT WORKING PROPERLY!"
    exit 1
else
    echo -e "${GREEN}✓ Data isolation working${NC}"
fi

# Test 6: Critical - Check if User A's results contain User B's data  
echo -n "Test 6: Checking User A cannot see User B's data... "
if echo "$user_a_workouts" | grep -q "User B test workout"; then
    echo -e "${RED}✗ CRITICAL: User A can see User B's workout!${NC}"
    echo "RLS IS NOT WORKING PROPERLY!"
    exit 1
else
    echo -e "${GREEN}✓ Data isolation working${NC}"
fi

# Test 7: Check anonymous access is blocked
echo -n "Test 7: Anonymous access blocked... "
anonymous_status=$(curl -s -w "%{http_code}" -X "GET" "${STAGING_URL}/api/workouts" -o /dev/null)

if [ "$anonymous_status" = "401" ] || [ "$anonymous_status" = "403" ]; then
    echo -e "${GREEN}✓ Properly rejected (${anonymous_status})${NC}"
else
    echo -e "${RED}✗ Not blocked (${anonymous_status})${NC}"
    exit 1
fi

echo
echo "========================================="
echo -e "${GREEN}🎉 ALL RLS TESTS PASSED! 🎉${NC}"
echo "========================================="
echo
echo "✅ User data isolation is working correctly"
echo "✅ Users can only see their own workouts"
echo "✅ Anonymous access is properly blocked"
echo "✅ Row Level Security is functioning properly"
echo
echo "Environment parity confirmed - RLS works in staging!"
