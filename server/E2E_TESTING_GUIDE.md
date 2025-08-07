# Manual E2E Testing Guide for RLS Implementation

This guide provides step-by-step instructions for manually testing the Row Level Security (RLS) implementation to ensure proper authentication and data isolation.

## 📋 Prerequisites

### 1. Server Setup
```bash
# Start the server
go run cmd/api/main.go
```
The server should start on the configured port (typically `:8080`)

### 2. Test Users Setup
You'll need **at least 2 different user accounts** with valid JWT tokens from Stack Auth:
- **User A**: `user-a-id` with JWT token
- **User B**: `user-b-id` with JWT token

### 3. Tools Required
- **cURL** (recommended for this guide)
- Or **Postman**, **HTTPie**, or any HTTP client

## 🧪 Test Scenarios

### **Scenario 1: User A Creates and Retrieves Their Own Workout**

#### Step 1.1: Create a workout as User A
```bash
curl -X POST http://localhost:8080/api/workouts \
  -H "Authorization: Bearer YOUR_USER_A_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2025-01-07T10:00:00Z",
    "notes": "User A's private workout"
  }'
```

**Expected Result:**
- ✅ Status: `200 OK` or `201 Created`
- ✅ Response contains workout data with an ID
- ✅ The workout should be associated with User A

#### Step 1.2: Retrieve workouts as User A
```bash
curl -X GET http://localhost:8080/api/workouts \
  -H "Authorization: Bearer YOUR_USER_A_JWT_TOKEN"
```

**Expected Result:**
- ✅ Status: `200 OK`
- ✅ Response contains the workout created in Step 1.1
- ✅ Only User A's workouts are returned

#### Step 1.3: Get specific workout by ID as User A
```bash
curl -X GET http://localhost:8080/api/workouts/{WORKOUT_ID} \
  -H "Authorization: Bearer YOUR_USER_A_JWT_TOKEN"
```
*(Replace `{WORKOUT_ID}` with the ID from Step 1.1)*

**Expected Result:**
- ✅ Status: `200 OK`
- ✅ Returns the specific workout data

---

### **Scenario 2: User B Cannot Access User A's Data**

#### Step 2.1: Try to list workouts as User B
```bash
curl -X GET http://localhost:8080/api/workouts \
  -H "Authorization: Bearer YOUR_USER_B_JWT_TOKEN"
```

**Expected Result:**
- ✅ Status: `200 OK`
- ✅ Empty array `[]` or no workouts (User B hasn't created any)
- ✅ User A's workout is NOT visible to User B

#### Step 2.2: Try to access User A's specific workout by ID as User B
```bash
curl -X GET http://localhost:8080/api/workouts/{USER_A_WORKOUT_ID} \
  -H "Authorization: Bearer YOUR_USER_B_JWT_TOKEN"
```

**Expected Result:**
- ✅ Status: `404 Not Found` or `200 OK` with empty result
- ✅ User B cannot see User A's workout data
- ✅ No sensitive information leaked in error messages

---

### **Scenario 3: User B Creates Their Own Data**

#### Step 3.1: Create a workout as User B
```bash
curl -X POST http://localhost:8080/api/workouts \
  -H "Authorization: Bearer YOUR_USER_B_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2025-01-07T14:00:00Z",
    "notes": "User B's private workout"
  }'
```

**Expected Result:**
- ✅ Status: `200 OK` or `201 Created`
- ✅ Workout created successfully for User B

#### Step 3.2: Verify User B can see their own data
```bash
curl -X GET http://localhost:8080/api/workouts \
  -H "Authorization: Bearer YOUR_USER_B_JWT_TOKEN"
```

**Expected Result:**
- ✅ Status: `200 OK`
- ✅ Returns User B's workout only
- ✅ User A's workout is still not visible

#### Step 3.3: Verify User A still cannot see User B's data
```bash
curl -X GET http://localhost:8080/api/workouts \
  -H "Authorization: Bearer YOUR_USER_A_JWT_TOKEN"
```

**Expected Result:**
- ✅ Status: `200 OK`
- ✅ Returns only User A's workout(s)
- ✅ User B's workout is not visible to User A

---

### **Scenario 4: Anonymous/Unauthenticated Access**

#### Step 4.1: Try to access workouts without authentication
```bash
curl -X GET http://localhost:8080/api/workouts
```

**Expected Result:**
- ✅ Status: `401 Unauthorized`
- ✅ Access denied without valid JWT token

#### Step 4.2: Try to access with invalid/expired token
```bash
curl -X GET http://localhost:8080/api/workouts \
  -H "Authorization: Bearer invalid.jwt.token"
```

**Expected Result:**
- ✅ Status: `401 Unauthorized`
- ✅ Invalid token rejected

---

### **Scenario 5: Exercise API Testing (Same Pattern)**

#### Step 5.1: Test exercise isolation for User A
```bash
# Create exercise as User A
curl -X POST http://localhost:8080/api/exercises \
  -H "Authorization: Bearer YOUR_USER_A_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "User A Exercise",
    "category": "strength"
  }'

# List exercises as User A
curl -X GET http://localhost:8080/api/exercises \
  -H "Authorization: Bearer YOUR_USER_A_JWT_TOKEN"
```

#### Step 5.2: Verify User B cannot see User A's exercises
```bash
curl -X GET http://localhost:8080/api/exercises \
  -H "Authorization: Bearer YOUR_USER_B_JWT_TOKEN"
```

**Expected Result:**
- ✅ User A's exercises not visible to User B
- ✅ Proper isolation maintained

---

## 🔍 What to Look For During Testing

### ✅ **Success Indicators**
1. **Data Isolation**: Users can only see their own data
2. **Authentication Required**: All endpoints require valid JWT tokens  
3. **Consistent Behavior**: RLS works across all API endpoints (workouts, exercises)
4. **No Data Leakage**: Error messages don't reveal other users' data
5. **Performance**: Requests complete in reasonable time (< 1 second typically)

### ❌ **Red Flags** (Security Issues)
1. **Cross-User Data Access**: User B can see User A's data
2. **Authentication Bypass**: Endpoints accessible without tokens
3. **Information Leakage**: Error messages reveal sensitive data
4. **Inconsistent Policies**: Some endpoints don't enforce RLS
5. **Performance Issues**: Requests take too long (> 5 seconds)

## 🚨 Troubleshooting

### Common Issues:

#### "Empty Results When Expected Data"
- Check JWT token is valid and not expired
- Verify the `sub` (subject) claim in JWT matches expected user ID
- Ensure RLS policies are enabled (not running as database superuser)

#### "Can See Other Users' Data" 
- **CRITICAL SECURITY ISSUE** - RLS is not working
- Check database connection is not using superuser role
- Verify RLS policies are enabled and applied
- Check middleware is setting `app.current_user_id` correctly

#### "Authentication Errors"
- Verify JWT token format: `Authorization: Bearer TOKEN`
- Check token is not expired
- Ensure JWKS URL is accessible and correct

## 📊 Test Results Documentation

Create a simple checklist to track your results:

```
□ Scenario 1: User A CRUD operations work
□ Scenario 2: User B cannot access User A's data  
□ Scenario 3: User B CRUD operations work (isolated)
□ Scenario 4: Anonymous access properly blocked
□ Scenario 5: Exercise API isolation works
□ Performance: All requests < 1 second
□ Security: No cross-user data leakage observed
```

## 🎯 Next Steps After Testing

1. **If all tests pass**: Update `todo.md` to mark E2E tests complete
2. **If issues found**: Document them and fix before moving to production
3. **Consider automation**: Convert successful manual tests to automated scripts

## 💡 Advanced Testing (Optional)

### Concurrent User Testing
Test multiple users simultaneously using tools like:
- **Apache Bench (ab)**: `ab -n 100 -c 10 -H "Authorization: Bearer TOKEN" http://localhost:8080/api/workouts`
- **Hey**: `hey -n 100 -c 10 -H "Authorization: Bearer TOKEN" http://localhost:8080/api/workouts`

### Token Edge Cases
- Test with malformed JWT tokens
- Test with valid format but invalid signature  
- Test with expired tokens
- Test token refresh scenarios
