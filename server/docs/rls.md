# Row Level Security (RLS) Documentation

This document covers the Row Level Security implementation in FitTrack, which ensures users can only access their own data through database-level access controls.

## Overview

Row Level Security (RLS) is implemented using PostgreSQL's native RLS features to provide multi-tenant data isolation. Each user can only access their own workouts, exercises, and sets through policies enforced at the database level.

## Session Variable: `app.current_user_id`

### How It's Set

The session variable `app.current_user_id` is automatically set by the authentication middleware for each request:

1. **Authentication Flow**: When a user makes an API request, the `Authenticator.Middleware()` extracts the user ID from the JWT token
2. **Session Setup**: The middleware calls `setSessionUserID()` which executes:
   ```sql
   SELECT set_config('app.current_user_id', $1, false) WHERE $1 IS NOT NULL
   ```
3. **Per-Connection**: Each database connection from the connection pool gets its own session variable, ensuring proper isolation between concurrent users

### Implementation Details

- **Location**: `internal/auth/middleware.go`
- **Trigger**: Every API request to paths starting with `/api/`
- **Scope**: Session-scoped (per database connection)
- **Isolation**: Connection pool ensures different users don't interfere with each other

## Protected Tables and Policies

### Core Function

All RLS policies use the `current_user_id()` function:

```sql
CREATE OR REPLACE FUNCTION current_user_id() 
RETURNS TEXT AS $$
    SELECT current_setting('app.current_user_id', true);
$$ LANGUAGE SQL STABLE;
```

### Protected Tables

#### 1. `users` Table
- **Policy**: `users_policy`
- **Logic**: Users can only access their own user record
- **SQL**: `user_id = current_user_id()`

#### 2. `workout` Table
- **Policies**: 
  - `workout_select_policy` (SELECT)
  - `workout_insert_policy` (INSERT) 
  - `workout_update_policy` (UPDATE)
  - `workout_delete_policy` (DELETE)
- **Logic**: Users can only access workouts they created
- **SQL**: `user_id = current_user_id()`

#### 3. `exercise` Table  
- **Policies**:
  - `exercise_select_policy` (SELECT)
  - `exercise_insert_policy` (INSERT)
  - `exercise_update_policy` (UPDATE) 
  - `exercise_delete_policy` (DELETE)
- **Logic**: Users can only access exercises they created
- **SQL**: `user_id = current_user_id()`

#### 4. `"set"` Table (quoted due to SQL keyword)
- **Policies**:
  - `set_select_policy` (SELECT)
  - `set_insert_policy` (INSERT)
  - `set_update_policy` (UPDATE)
  - `set_delete_policy` (DELETE)
- **Logic**: Users can only access sets that belong to their workouts (indirect relationship)
- **SQL**: 
  ```sql
  EXISTS (
      SELECT 1 FROM workout w 
      WHERE w.id = "set".workout_id 
      AND w.user_id = current_user_id()
  )
  ```

### Policy Behavior

- **Filtering**: RLS policies filter results automatically - unauthorized data is simply not returned
- **No Errors**: Blocked access returns empty result sets rather than permission errors
- **All Operations**: Policies cover SELECT, INSERT, UPDATE, and DELETE operations
- **Automatic**: No application code changes needed - policies are enforced at the database level

## Testing

### Running Tests Locally

Run all tests including RLS integration tests:

```bash
go test ./...
```

### Test Categories

1. **Unit Tests**: Mock-based tests for individual components
2. **Integration Tests**: Real database tests with RLS enabled
3. **Performance Tests**: Connection pool and RLS overhead validation
4. **Security Tests**: Bypass prevention and edge case handling

### Test Utilities

For testing with RLS, use the provided utility:

```go
import "github.com/Andrewy-gh/fittrack/server/internal/testutils"

// Set user context for testing
ctx = testutils.SetTestUserContext(ctx, t, pool, "test-user-id")
```

Or set directly in SQL:
```sql
SELECT set_config('app.current_user_id', 'test-user-id', false);
```

### Test Coverage

- **Connection Pool Isolation**: Verifies session variables don't leak between users
- **Policy Enforcement**: Confirms users can only access their own data
- **Bypass Prevention**: Tests against SQL injection and privilege escalation
- **Edge Cases**: Handles missing/invalid session variables gracefully
- **Performance**: Validates acceptable RLS overhead

## Important Notes

### Superuser Behavior
⚠️ **Critical**: RLS policies do **not apply to database superusers** by design in PostgreSQL. 

- **Development**: Local development may use superuser, which bypasses RLS
- **Production**: Must use non-superuser database role for RLS to function
- **Testing**: Tests automatically skip when running as superuser

### Production Requirements

⚠️ **CRITICAL**: For RLS to work in production, the `DATABASE_URL` environment variable **MUST NOT** use a superuser role.

#### Step 1: Verify Current DATABASE_URL is Not Using a Superuser

Before deployment, verify that your current database connection does not use a superuser:

```sql
-- Connect using your current DATABASE_URL and run this query
SELECT 
    current_user as current_role,
    rolsuper as is_superuser,
    rolcanlogin as can_login
FROM pg_roles 
WHERE rolname = current_user;

-- Expected result for production:
-- current_role | is_superuser | can_login
-- fittrack_app | f           | t
-- 
-- If is_superuser = 't', you MUST create a non-superuser role!
```

**If the result shows `is_superuser = t` (true), your current DATABASE_URL is using a superuser and RLS will NOT work in production.**

#### Production Database Role Creation

Create a dedicated non-superuser role for the application:

```sql
-- Create application role (non-superuser)
CREATE ROLE fittrack_app WITH LOGIN PASSWORD 'your_secure_password_here';

-- Grant database connection permission
GRANT CONNECT ON DATABASE fittrack TO fittrack_app;

-- Grant table permissions (SELECT, INSERT, UPDATE, DELETE only)
GRANT SELECT, INSERT, UPDATE, DELETE ON users TO fittrack_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON workout TO fittrack_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON exercise TO fittrack_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON "set" TO fittrack_app;

-- Grant sequence usage (required for INSERT with SERIAL columns)
GRANT USAGE ON SEQUENCE users_id_seq TO fittrack_app;
GRANT USAGE ON SEQUENCE workout_id_seq TO fittrack_app;
GRANT USAGE ON SEQUENCE exercise_id_seq TO fittrack_app;
GRANT USAGE ON SEQUENCE set_id_seq TO fittrack_app;

-- Verify role is not a superuser (should return false)
SELECT rolsuper FROM pg_roles WHERE rolname = 'fittrack_app';
```

#### Environment Variable Configuration

Update your production `DATABASE_URL` to use the application role:

```bash
# Example production DATABASE_URL
DATABASE_URL="postgresql://fittrack_app:your_secure_password_here@your-db-host:5432/fittrack?sslmode=require"
```

#### Production Security Checklist

Before deploying to production, verify all of the following:

- [ ] **Step 1**: Database role is **NOT** a superuser (run verification query above)
- [ ] **Step 2**: Role has been granted **ONLY** the minimum required permissions:
  - [ ] `CONNECT` on the database
  - [ ] `SELECT, INSERT, UPDATE, DELETE` on application tables only (no DDL permissions)
  - [ ] `USAGE` on sequences for SERIAL columns only
- [ ] **Step 3**: DATABASE_URL secret uses the non-superuser role
- [ ] **Step 4**: RLS policies are active on all tables:
  ```sql
  -- Verify RLS is enabled (should return 4 rows with enabled=true)
  SELECT schemaname, tablename, rowlsecurity as enabled
  FROM pg_tables 
  WHERE tablename IN ('users', 'workout', 'exercise', 'set')
  AND schemaname = 'public';
  ```
- [ ] **Step 5**: Test with production credentials to ensure RLS works
- [ ] **Step 6**: Monitor application logs for any permission-related errors

#### Platform-Specific Secret Management

**Kubernetes:**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: fittrack-db-secret
type: Opaque
stringData:
  DATABASE_URL: "postgresql://fittrack_app:your_secure_password_here@your-db-host:5432/fittrack?sslmode=require"
```

**Heroku:**
```bash
heroku config:set DATABASE_URL="postgresql://fittrack_app:your_secure_password_here@your-db-host:5432/fittrack?sslmode=require"
```

**EC2/Docker:**
```bash
# In your environment file or systemd service
DATABASE_URL="postgresql://fittrack_app:your_secure_password_here@your-db-host:5432/fittrack?sslmode=require"
```

### Performance Considerations

- RLS adds filtering clauses to every query
- Proper indexes exist on `user_id` columns
- Typical overhead: ~10-200% depending on query complexity
- Connection pool isolation works efficiently

## Troubleshooting

### Common Issues

1. **No Data Returned**: Check if session variable is set correctly
2. **Performance Issues**: Verify indexes on `user_id` columns exist  
3. **Test Failures**: Ensure running as non-superuser role
4. **Access Denied**: Verify JWT token contains correct user ID

### TODO: Full Troubleshooting Guide

A comprehensive troubleshooting guide will be added if issues arise during development or deployment. Please report any RLS-related issues to help build this section.

## Migration Information

RLS policies are applied via migration `00005_add_rls_policies.sql`. See [`migrations/README.md`](../migrations/README.md) for detailed migration information.
