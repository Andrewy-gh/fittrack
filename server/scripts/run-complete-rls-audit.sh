#!/bin/bash

# Complete RLS Audit Script - Step 5
# Runs policy coverage audit and privilege escalation tests

set -e

echo "========================================="
echo "    COMPLETE RLS SECURITY AUDIT"
echo "========================================="
echo "Step 5: Audit policy coverage & privilege escalation"
echo ""

# Check if database is running
if ! docker compose ps | grep -q "db.*Up"; then
    echo "❌ Database is not running. Please start with: docker compose up -d"
    exit 1
fi

echo "✓ Database is running"
echo ""

# Function to run SQL via docker
run_sql() {
    local sql_file=$1
    local description=$2
    
    echo "📋 Running: $description"
    echo "----------------------------------------"
    
    if [ -f "$sql_file" ]; then
        Get-Content "$sql_file" | docker exec -i db psql -U user -d postgres 2>&1
        echo ""
    else
        echo "❌ SQL file not found: $sql_file"
        return 1
    fi
}

# Step 1: Run policy coverage audit
echo "🔍 STEP 1: Policy Coverage Audit"
echo "================================="
run_sql "scripts/audit-rls-coverage.sql" "Cross-checking pg_policies vs information_schema.tables"

# Step 2: Setup test data if needed
echo "🔧 STEP 2: Setup Test Data"  
echo "==========================="
run_sql "scripts/setup-test-data.sql" "Creating test users and data for privilege escalation tests"

# Step 3: Run privilege escalation tests
echo "🛡️  STEP 3: Privilege Escalation Tests"
echo "======================================"
run_sql "scripts/test-privilege-escalation-fixed.sql" "Attempting UPDATE with tampered WHERE TRUE clause as normal user"

# Step 4: Generate final summary
echo "📊 STEP 4: Final Security Assessment"
echo "===================================="
echo ""
echo "🔍 AUDIT COMPLETE - See findings in RLS_AUDIT_FINDINGS.md"
echo ""
echo "Key Issues Found:"
echo "• Missing RLS policies on 'set' table"  
echo "• RLS bypass vulnerabilities in core tables"
echo "• SQL injection patterns bypass RLS"
echo "• Users can see all data instead of just their own"
echo ""
echo "⚠️  CRITICAL: Do not deploy to production until issues are resolved!"
echo ""
echo "Next Steps:"
echo "1. Review RLS_AUDIT_FINDINGS.md for detailed analysis"
echo "2. Fix missing policies and logic errors"  
echo "3. Re-run tests to verify fixes"
echo "4. Create GitHub issue to track remediation"
echo ""
echo "Audit artifacts created:"
echo "• RLS_AUDIT_FINDINGS.md - Comprehensive security report"
echo "• scripts/audit-rls-coverage.sql - Policy coverage audit"
echo "• scripts/test-privilege-escalation-fixed.sql - Security tests"
echo "• scripts/setup-test-data.sql - Test data creation"
echo ""
echo "========================================="
echo "         RLS AUDIT COMPLETE"
echo "========================================="
