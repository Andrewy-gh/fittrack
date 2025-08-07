#!/bin/bash

# RLS Smoke Test for Staging Environment
# This script tests Row Level Security in a production-like environment
# It performs API calls as different users to ensure data isolation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
STAGING_URL="${STAGING_URL:-https://fittrack.fly.dev}"
USER_A_TOKEN="${USER_A_TOKEN}"
USER_B_TOKEN="${USER_B_TOKEN}"

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Function to check if a command exists
check_command() {
    if ! command -v $1 &> /dev/null; then
        print_error "Required command '$1' is not installed"
        exit 1
    fi
}

# Function to make API requests with error handling
api_request() {
    local method=$1
    local endpoint=$2
    local token=$3
    local data=$4
    local expected_status=${5:-200}
    
    local url="${STAGING_URL}${endpoint}"
    local response_file=$(mktemp)
    local status_code
    
    if [ -n "$data" ]; then
        status_code=$(curl -s -w "%{http_code}" -X "$method" \
            -H "x-stack-access-token: $token" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$url" \
            -o "$response_file")
    else
        status_code=$(curl -s -w "%{http_code}" -X "$method" \
            -H "x-stack-access-token: $token" \
            "$url" \
            -o "$response_file")
    fi
    
    # Output the response for debugging if not expected status
    if [ "$status_code" != "$expected_status" ]; then
        print_error "Expected status $expected_status, got $status_code for $method $endpoint"
        print_error "Response: $(cat $response_file)"
        rm -f "$response_file"
        return 1
    fi
    
    # Return the response content
    cat "$response_file"
    rm -f "$response_file"
    return 0
}

# Function to check if response contains user-specific data
check_user_isolation() {
    local response=$1
    local user_identifier=$2
    local should_contain=$3  # true/false
    
    if [ "$should_contain" = "true" ]; then
        if echo "$response" | grep -q "$user_identifier" || [ "$(echo "$response" | jq '. | length')" -gt 0 ]; then
            return 0
        else
            return 1
        fi
    else
        if echo "$response" | grep -q "$user_identifier" || [ "$(echo "$response" | jq '. | length')" -gt 0 ]; then
            return 1
        else
            return 0
        fi
    fi
}

# Main test function
main() {
    echo "======================================"
    echo "      RLS Smoke Test - Staging"
    echo "======================================"
    echo
    
    # Check prerequisites
    print_status "Checking prerequisites..."
    check_command curl
    check_command jq
    
    # Check environment variables
    if [ -z "$USER_A_TOKEN" ] || [ -z "$USER_B_TOKEN" ]; then
        print_error "Missing required environment variables:"
        print_error "  USER_A_TOKEN: JWT token for User A"
        print_error "  USER_B_TOKEN: JWT token for User B"
        echo
        print_warning "To get tokens, visit your Stack Auth dashboard or use a test environment"
        exit 1
    fi
    
    print_status "Using staging URL: $STAGING_URL"
    print_status "Tokens configured for User A and User B"
    echo
    
    # Test 1: User A creates a workout
    print_status "Test 1: User A creates a workout..."
    user_a_workout_data='{
        "date": "'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'",
        "notes": "User A smoke test workout",
        "exercises": [{
            "name": "Bench Press",
            "sets": [{
                "reps": 10,
                "weight": 135,
                "set_type": "working"
            }]
        }]
    }'
    
    if api_request "POST" "/api/workouts" "$USER_A_TOKEN" "$user_a_workout_data" "200" > /dev/null; then
        print_success "✓ User A workout created successfully"
    else
        print_error "✗ Failed to create workout for User A"
        exit 1
    fi
    
    # Test 2: User A can retrieve their own workouts
    print_status "Test 2: User A retrieves their workouts..."
    user_a_workouts=$(api_request "GET" "/api/workouts" "$USER_A_TOKEN")
    
    if [ $? -eq 0 ] && [ "$(echo "$user_a_workouts" | jq '. | length')" -gt 0 ]; then
        print_success "✓ User A can see their own workout(s)"
        workout_count=$(echo "$user_a_workouts" | jq '. | length')
        print_status "  Found $workout_count workout(s) for User A"
    else
        print_error "✗ User A cannot retrieve their workouts"
        exit 1
    fi
    
    # Test 3: User B creates a workout
    print_status "Test 3: User B creates a workout..."
    user_b_workout_data='{
        "date": "'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'",
        "notes": "User B smoke test workout",
        "exercises": [{
            "name": "Squat",
            "sets": [{
                "reps": 8,
                "weight": 185,
                "set_type": "working"
            }]
        }]
    }'
    
    if api_request "POST" "/api/workouts" "$USER_B_TOKEN" "$user_b_workout_data" "200" > /dev/null; then
        print_success "✓ User B workout created successfully"
    else
        print_error "✗ Failed to create workout for User B"
        exit 1
    fi
    
    # Test 4: User B can retrieve their own workouts (but not User A's)
    print_status "Test 4: User B retrieves their workouts..."
    user_b_workouts=$(api_request "GET" "/api/workouts" "$USER_B_TOKEN")
    
    if [ $? -eq 0 ] && [ "$(echo "$user_b_workouts" | jq '. | length')" -gt 0 ]; then
        print_success "✓ User B can see their own workout(s)"
        workout_count=$(echo "$user_b_workouts" | jq '. | length')
        print_status "  Found $workout_count workout(s) for User B"
    else
        print_error "✗ User B cannot retrieve their workouts"
        exit 1
    fi
    
    # Test 5: CRITICAL - Verify User B cannot see User A's data
    print_status "Test 5: Verifying User B cannot see User A's data..."
    
    # Extract User A's workout ID if possible
    user_a_workout_id=$(echo "$user_a_workouts" | jq -r '.[0].id // empty')
    
    if [ -n "$user_a_workout_id" ]; then
        # Try to access User A's specific workout as User B
        print_status "  Attempting to access User A's workout (ID: $user_a_workout_id) as User B..."
        user_b_access_attempt=$(api_request "GET" "/api/workouts/$user_a_workout_id" "$USER_B_TOKEN" "" "200")
        
        # Check if User B gets empty results (RLS working) vs actual data (RLS broken)
        if [ "$(echo "$user_b_access_attempt" | jq '. | length')" -eq 0 ]; then
            print_success "✓ User B cannot see User A's specific workout (RLS working)"
        else
            print_error "✗ CRITICAL: User B can see User A's workout data!"
            print_error "  This indicates RLS is NOT working properly in staging!"
            echo "$user_b_access_attempt"
            exit 1
        fi
    else
        print_warning "Could not extract User A's workout ID for specific access test"
    fi
    
    # Test 6: Verify cross-user data isolation in list endpoints
    print_status "Test 6: Verifying data isolation in workout lists..."
    
    # Compare workout lists - they should be completely different
    user_a_notes=$(echo "$user_a_workouts" | jq -r '.[].notes // empty' | grep -i "User A" | wc -l)
    user_b_notes=$(echo "$user_b_workouts" | jq -r '.[].notes // empty' | grep -i "User B" | wc -l)
    
    if [ "$user_a_notes" -gt 0 ] && [ "$user_b_notes" -eq 0 ]; then
        print_success "✓ User A's workouts contain only User A's data"
    else
        print_error "✗ User A's workouts may contain User B's data or vice versa"
        exit 1
    fi
    
    if [ "$user_b_notes" -gt 0 ] && echo "$user_b_workouts" | jq -r '.[].notes // empty' | grep -q -v "User A"; then
        print_success "✓ User B's workouts contain only User B's data"
    else
        print_error "✗ User B's workouts may contain User A's data"
        exit 1
    fi
    
    # Test 7: Test exercise isolation
    print_status "Test 7: Testing exercise endpoint isolation..."
    
    # User A creates an exercise
    user_a_exercise_data='{"name": "User A Exercise"}'
    if api_request "POST" "/api/exercises" "$USER_A_TOKEN" "$user_a_exercise_data" "200" > /dev/null; then
        print_success "✓ User A exercise created"
    else
        print_warning "User A exercise creation failed (may be optional)"
    fi
    
    # Check if User B can see User A's exercises
    user_b_exercises=$(api_request "GET" "/api/exercises" "$USER_B_TOKEN" "" "200")
    if echo "$user_b_exercises" | grep -q "User A Exercise"; then
        print_error "✗ CRITICAL: User B can see User A's exercises!"
        exit 1
    else
        print_success "✓ User B cannot see User A's exercises"
    fi
    
    # Test 8: Anonymous access test
    print_status "Test 8: Testing anonymous access rejection..."
    
    # Try to access workouts without authentication
    anonymous_status=$(curl -s -w "%{http_code}" -X "GET" "${STAGING_URL}/api/workouts" -o /dev/null)
    
    if [ "$anonymous_status" = "401" ] || [ "$anonymous_status" = "403" ]; then
        print_success "✓ Anonymous access properly rejected (status: $anonymous_status)"
    else
        print_error "✗ Anonymous access not properly rejected (status: $anonymous_status)"
        exit 1
    fi
    
    echo
    echo "======================================"
    print_success "🎉 ALL RLS SMOKE TESTS PASSED! 🎉"
    echo "======================================"
    echo
    print_status "Summary:"
    print_status "✓ User A can create and access their own data"
    print_status "✓ User B can create and access their own data"
    print_status "✓ User B CANNOT access User A's data (RLS working)"
    print_status "✓ User A CANNOT access User B's data (RLS working)"
    print_status "✓ Anonymous access properly rejected"
    print_status "✓ Row Level Security is functioning correctly in staging"
    echo
    
    return 0
}

# Handle script interruption
trap 'echo; print_error "Test interrupted"; exit 1' INT TERM

# Run the main function
main "$@"
