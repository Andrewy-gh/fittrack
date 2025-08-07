# RLS Smoke Test Scripts

This directory contains scripts for testing Row Level Security (RLS) in staging and production-like environments to ensure proper user data isolation.

## Overview

The smoke tests verify that:
1. User A can create and access their own data
2. User B can create and access their own data  
3. **CRITICAL:** User B cannot see User A's data (and vice versa)
4. Anonymous access is properly blocked
5. RLS policies are working correctly in the target environment

## Scripts

### 1. `quick-rls-test.sh` - Simple Manual Test

A lightweight script for quick RLS verification against an existing environment.

**Usage:**
```bash
# Export test user tokens
export USER_A_TOKEN="your-test-user-a-jwt-token"
export USER_B_TOKEN="your-test-user-b-jwt-token"

# Run against staging
./scripts/quick-rls-test.sh https://fittrack.fly.dev

# Or use environment variable
export STAGING_URL="https://your-staging.fly.dev"
./scripts/quick-rls-test.sh
```

**Features:**
- Quick 7-test validation
- Clear pass/fail output
- Tests both workout and exercise endpoints
- Verifies anonymous access blocking

### 2. `rls-smoke-test.sh` - Comprehensive Test Suite

A thorough test suite with detailed reporting and edge case testing.

**Usage:**
```bash
export USER_A_TOKEN="your-test-user-a-jwt-token"
export USER_B_TOKEN="your-test-user-b-jwt-token"
export STAGING_URL="https://fittrack.fly.dev"

./scripts/rls-smoke-test.sh
```

**Features:**
- 8 comprehensive test scenarios
- Detailed logging and error reporting
- Tests specific workout access by ID
- Concurrent user testing
- Exercise endpoint isolation testing
- Full cross-user data leakage verification

### 3. GitHub Action: `rls-smoke-test.yml`

Automated CI/CD pipeline that deploys to staging and runs smoke tests.

**Features:**
- Deploys current branch to staging
- Waits for deployment readiness
- Runs comprehensive RLS tests
- Fails pipeline if User B can see User A's data
- Automatic cleanup

## Getting Test Tokens

To run these tests, you need JWT tokens for two different test users:

### Method 1: Stack Auth Dashboard
1. Visit your Stack Auth project dashboard
2. Create two test user accounts
3. Generate JWT tokens for each user
4. Store as environment variables or GitHub secrets

### Method 2: Browser Dev Tools
1. Login to your staging app as Test User A
2. Open browser dev tools → Network tab
3. Make an API request and copy the `x-stack-access-token` header value
4. Repeat for Test User B in incognito/different browser

### Method 3: API Token Generation
```bash
# Example using your auth endpoint
curl -X POST https://api.stack-auth.com/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test-user-a@example.com", "password": "testpass"}'
```

## Environment Setup

### Local Development
```bash
# Required environment variables
export STAGING_URL="https://fittrack.fly.dev"
export USER_A_TOKEN="eyJhbGciOiJSUzI1NiIs..."
export USER_B_TOKEN="eyJhbGciOiJSUzI1NiIs..."

# Make script executable
chmod +x scripts/quick-rls-test.sh
chmod +x scripts/rls-smoke-test.sh

# Run tests
./scripts/quick-rls-test.sh
```

### GitHub Actions
Add these secrets to your repository:

**Required Secrets:**
- `TEST_USER_A_TOKEN` - JWT token for test user A
- `TEST_USER_B_TOKEN` - JWT token for test user B
- `STAGING_DATABASE_URL` - Database connection for staging migrations
- `FLY_API_TOKEN` - Fly.io deployment token

**Optional Secrets:**
- `STAGING_FLY_APP_NAME` - If using separate staging app (default: `fittrack`)
- `VITE_PROJECT_ID` - Frontend environment variable
- `VITE_PUBLISHABLE_CLIENT_KEY` - Frontend environment variable

## Expected Test Results

### ✅ Success (RLS Working)
```
Test 1: User A creates workout... ✓ Success
Test 2: User B creates workout... ✓ Success  
Test 3: User A retrieves workouts... ✓ Found 1 workout(s)
Test 4: User B retrieves workouts... ✓ Found 1 workout(s)
Test 5: Checking User B cannot see User A's data... ✓ Data isolation working
Test 6: Checking User A cannot see User B's data... ✓ Data isolation working
Test 7: Anonymous access blocked... ✓ Properly rejected (401)

🎉 ALL RLS TESTS PASSED! 🎉
```

### ❌ Failure (RLS Broken)
```
Test 5: Checking User B cannot see User A's data... ✗ CRITICAL: User B can see User A's workout!
RLS IS NOT WORKING PROPERLY!
```

## Troubleshooting

### Common Issues

#### "Missing test user tokens"
- Ensure `USER_A_TOKEN` and `USER_B_TOKEN` are set
- Verify tokens are valid and not expired
- Check tokens are for different users

#### "401 Unauthorized" for valid tokens
- Verify `STAGING_URL` is correct
- Check if staging deployment is ready
- Ensure Stack Auth configuration matches

#### "RLS IS NOT WORKING PROPERLY!"
🚨 **CRITICAL SECURITY ISSUE** 🚨
- RLS policies may not be enabled
- Database user might be a superuser (bypasses RLS)
- Session variable `app.current_user_id` not being set
- Review RLS implementation immediately

#### Tests pass locally but fail in staging
- Verify staging database has RLS policies applied
- Check staging database user is NOT a superuser
- Ensure migrations ran successfully
- Verify authentication middleware is working

### Debug Mode

Run with verbose output:
```bash
set -x  # Enable bash debug mode
./scripts/quick-rls-test.sh
```

View API responses:
```bash
# Test API manually
curl -H "x-stack-access-token: $USER_A_TOKEN" \
  https://fittrack.fly.dev/api/workouts
```

## Security Checklist

Before deploying to production, ensure:

- [ ] All RLS smoke tests pass
- [ ] User A cannot see User B's data
- [ ] User B cannot see User A's data
- [ ] Anonymous access returns 401/403
- [ ] Database user is NOT a superuser
- [ ] RLS policies are enabled on all tables
- [ ] Session variables are set correctly
- [ ] Tests pass in staging environment

## CI/CD Integration

### Basic Pipeline
```yaml
- name: Run RLS Smoke Tests
  run: |
    export USER_A_TOKEN="${{ secrets.TEST_USER_A_TOKEN }}"
    export USER_B_TOKEN="${{ secrets.TEST_USER_B_TOKEN }}"
    ./scripts/quick-rls-test.sh
```

### Full Pipeline
Use the provided `rls-smoke-test.yml` GitHub Action for complete deployment and testing.

## Support

If RLS tests fail:
1. Check the [RLS documentation](../docs/rls.md)
2. Review [E2E testing guide](../E2E_TESTING_GUIDE.md)  
3. Run integration tests: `go test ./...`
4. Verify database RLS policies are active

**Remember:** RLS failures are critical security issues that can lead to data breaches. Never deploy to production if RLS tests are failing.
