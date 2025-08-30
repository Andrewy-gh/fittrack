$ make test
go test -v ./...
# github.com/Andrewy-gh/fittrack/server/internal/workout
internal\workout\handler_test.go:569:2: expected declaration, found '}'
FAIL    github.com/Andrewy-gh/fittrack/server/internal/workout [setup failed]
?       github.com/Andrewy-gh/fittrack/server/cmd/api   [no test files]
?       github.com/Andrewy-gh/fittrack/server/cmd/test-jwk      [no test files]
?       github.com/Andrewy-gh/fittrack/server/docs      [no test files]
=== RUN   TestAuthenticator_Middleware
=== RUN   TestAuthenticator_Middleware/bypass_non-API_paths
=== RUN   TestAuthenticator_Middleware/missing_access_token
=== RUN   TestAuthenticator_Middleware/invalid_access_token
=== RUN   TestAuthenticator_Middleware/user_service_failure
=== RUN   TestAuthenticator_Middleware/successful_authentication_with_existing_user
--- PASS: TestAuthenticator_Middleware (0.00s)
    --- PASS: TestAuthenticator_Middleware/bypass_non-API_paths (0.00s)
    --- PASS: TestAuthenticator_Middleware/missing_access_token (0.00s)
    --- PASS: TestAuthenticator_Middleware/invalid_access_token (0.00s)
    --- PASS: TestAuthenticator_Middleware/user_service_failure (0.00s)
    --- PASS: TestAuthenticator_Middleware/successful_authentication_with_existing_user (0.00s)        
=== RUN   TestAuthenticator_Middleware_SessionUserID
=== RUN   TestAuthenticator_Middleware_SessionUserID/successfully_set_session_user_ID
=== RUN   TestAuthenticator_Middleware_SessionUserID/error_setting_session_user_ID
--- PASS: TestAuthenticator_Middleware_SessionUserID (0.00s)
    --- PASS: TestAuthenticator_Middleware_SessionUserID/successfully_set_session_user_ID (0.00s)      
    --- PASS: TestAuthenticator_Middleware_SessionUserID/error_setting_session_user_ID (0.00s)
=== RUN   TestAuthenticator_Middleware_NilDBPool
--- PASS: TestAuthenticator_Middleware_NilDBPool (0.00s)
=== RUN   TestJWKSCache_GetUserIDFromToken
--- PASS: TestJWKSCache_GetUserIDFromToken (0.00s)
=== RUN   TestJWKSCache_GetUserIDFromToken_ErrorCases
=== RUN   TestJWKSCache_GetUserIDFromToken_ErrorCases/malformed_token
=== RUN   TestJWKSCache_GetUserIDFromToken_ErrorCases/empty_token
--- PASS: TestJWKSCache_GetUserIDFromToken_ErrorCases (0.00s)
    --- PASS: TestJWKSCache_GetUserIDFromToken_ErrorCases/malformed_token (0.00s)
    --- PASS: TestJWKSCache_GetUserIDFromToken_ErrorCases/empty_token (0.00s)
=== RUN   TestErrorResponse
--- PASS: TestErrorResponse (0.00s)
PASS
ok      github.com/Andrewy-gh/fittrack/server/internal/auth     (cached)
=== RUN   TestConnectionPoolIsolation
    connection_pool_test.go:89: Set user context for test-user-1, got back: test-user-1
    connection_pool_test.go:95: RLS enabled on workout table: true
    connection_pool_test.go:102: Current user: user, Is superuser: true
    connection_pool_test.go:115: Policies on workout table: [workout_update_policy workout_delete_policy workout_select_policy workout_insert_policy]
    connection_pool_test.go:140: ✅ SUPERUSER BYPASS: User test-user-1 accessed workout 1 (RLS policies don't apply to superusers)
    connection_pool_test.go:89: Set user context for test-user-2, got back: test-user-2
    connection_pool_test.go:89: Set user context for test-user-1, got back: test-user-1
    connection_pool_test.go:89: Set user context for test-user-2, got back: test-user-2
    connection_pool_test.go:95: RLS enabled on workout table: true
    connection_pool_test.go:95: RLS enabled on workout table: true
    connection_pool_test.go:95: RLS enabled on workout table: true
    connection_pool_test.go:102: Current user: user, Is superuser: true
    connection_pool_test.go:102: Current user: user, Is superuser: true
    connection_pool_test.go:102: Current user: user, Is superuser: true
    connection_pool_test.go:115: Policies on workout table: [workout_update_policy workout_delete_policy workout_select_policy workout_insert_policy]
    connection_pool_test.go:115: Policies on workout table: [workout_update_policy workout_delete_policy workout_select_policy workout_insert_policy]
    connection_pool_test.go:115: Policies on workout table: [workout_update_policy workout_delete_policy workout_select_policy workout_insert_policy]
    connection_pool_test.go:140: ✅ SUPERUSER BYPASS: User test-user-1 accessed workout 2 (RLS policies don't apply to superusers)
    connection_pool_test.go:140: ✅ SUPERUSER BYPASS: User test-user-2 accessed workout 2 (RLS policies don't apply to superusers)
    connection_pool_test.go:140: ✅ SUPERUSER BYPASS: User test-user-2 accessed workout 1 (RLS policies don't apply to superusers)
--- PASS: TestConnectionPoolIsolation (0.42s)
=== RUN   TestIsUniqueConstraintError
=== RUN   TestIsUniqueConstraintError/postgres_duplicate_key_error
=== RUN   TestIsUniqueConstraintError/postgres_unique_constraint_error
=== RUN   TestIsUniqueConstraintError/sqlite_unique_constraint_error
=== RUN   TestIsUniqueConstraintError/other_error
=== RUN   TestIsUniqueConstraintError/nil_error
--- PASS: TestIsUniqueConstraintError (0.00s)
    --- PASS: TestIsUniqueConstraintError/postgres_duplicate_key_error (0.00s)
    --- PASS: TestIsUniqueConstraintError/postgres_unique_constraint_error (0.00s)
    --- PASS: TestIsUniqueConstraintError/sqlite_unique_constraint_error (0.00s)
    --- PASS: TestIsUniqueConstraintError/other_error (0.00s)
    --- PASS: TestIsUniqueConstraintError/nil_error (0.00s)
=== RUN   TestIsForeignKeyConstraintError
=== RUN   TestIsForeignKeyConstraintError/postgres_foreign_key_error
=== RUN   TestIsForeignKeyConstraintError/sqlite_foreign_key_error
=== RUN   TestIsForeignKeyConstraintError/other_error
=== RUN   TestIsForeignKeyConstraintError/nil_error
--- PASS: TestIsForeignKeyConstraintError (0.00s)
    --- PASS: TestIsForeignKeyConstraintError/postgres_foreign_key_error (0.00s)
    --- PASS: TestIsForeignKeyConstraintError/sqlite_foreign_key_error (0.00s)
    --- PASS: TestIsForeignKeyConstraintError/other_error (0.00s)
    --- PASS: TestIsForeignKeyConstraintError/nil_error (0.00s)
=== RUN   TestIsRowLevelSecurityError
=== RUN   TestIsRowLevelSecurityError/RLS_error
=== RUN   TestIsRowLevelSecurityError/wrapped_RLS_error
=== RUN   TestIsRowLevelSecurityError/postgres_permission_denied_error
=== RUN   TestIsRowLevelSecurityError/insufficient_privilege_error
=== RUN   TestIsRowLevelSecurityError/other_error
=== RUN   TestIsRowLevelSecurityError/nil_error
--- PASS: TestIsRowLevelSecurityError (0.00s)
    --- PASS: TestIsRowLevelSecurityError/RLS_error (0.00s)
    --- PASS: TestIsRowLevelSecurityError/wrapped_RLS_error (0.00s)
    --- PASS: TestIsRowLevelSecurityError/postgres_permission_denied_error (0.00s)
    --- PASS: TestIsRowLevelSecurityError/insufficient_privilege_error (0.00s)
    --- PASS: TestIsRowLevelSecurityError/other_error (0.00s)
    --- PASS: TestIsRowLevelSecurityError/nil_error (0.00s)
=== RUN   TestIsRLSContextError
=== RUN   TestIsRLSContextError/RLS_context_error
=== RUN   TestIsRLSContextError/set_config_error
=== RUN   TestIsRLSContextError/session_variable_error
=== RUN   TestIsRLSContextError/other_error
=== RUN   TestIsRLSContextError/nil_error
--- PASS: TestIsRLSContextError (0.00s)
    --- PASS: TestIsRLSContextError/RLS_context_error (0.00s)
    --- PASS: TestIsRLSContextError/set_config_error (0.00s)
    --- PASS: TestIsRLSContextError/session_variable_error (0.00s)
    --- PASS: TestIsRLSContextError/other_error (0.00s)
    --- PASS: TestIsRLSContextError/nil_error (0.00s)
=== RUN   TestRLSConnectionPoolIsolation
    rls_security_test.go:62: Skipping performance tests - running as superuser (RLS policies are bypassed)
--- SKIP: TestRLSConnectionPoolIsolation (0.17s)
=== RUN   TestRLSPerformanceImpact
    rls_security_test.go:148: Skipping performance tests - running as superuser (RLS policies are bypassed)
--- SKIP: TestRLSPerformanceImpact (0.14s)
=== RUN   TestRLSPolicyBypassPrevention
    rls_security_test.go:249: Skipping bypass prevention tests - running as superuser (RLS policies are bypassed)
--- SKIP: TestRLSPolicyBypassPrevention (0.14s)
=== RUN   TestRLSSessionVariableEdgeCases
    rls_security_test.go:414: Skipping edge case tests - running as superuser (RLS policies are bypassed)
--- SKIP: TestRLSSessionVariableEdgeCases (0.13s)
PASS
ok      github.com/Andrewy-gh/fittrack/server/internal/database (cached)
=== RUN   TestExerciseHandler_ListExercises
=== RUN   TestExerciseHandler_ListExercises/successful_fetch
=== RUN   TestExerciseHandler_ListExercises/internal_server_error
=== RUN   TestExerciseHandler_ListExercises/unauthenticated_user
--- PASS: TestExerciseHandler_ListExercises (0.00s)
    --- PASS: TestExerciseHandler_ListExercises/successful_fetch (0.00s)
    --- PASS: TestExerciseHandler_ListExercises/internal_server_error (0.00s)
    --- PASS: TestExerciseHandler_ListExercises/unauthenticated_user (0.00s)
=== RUN   TestExerciseHandler_GetExerciseWithSets
=== RUN   TestExerciseHandler_GetExerciseWithSets/successful_fetch
=== RUN   TestExerciseHandler_GetExerciseWithSets/invalid_exercise_ID
=== RUN   TestExerciseHandler_GetExerciseWithSets/service_error
=== RUN   TestExerciseHandler_GetExerciseWithSets/empty_results_-_returns_200_with_empty_array
=== RUN   TestExerciseHandler_GetExerciseWithSets/unauthenticated_user
--- PASS: TestExerciseHandler_GetExerciseWithSets (0.00s)
    --- PASS: TestExerciseHandler_GetExerciseWithSets/successful_fetch (0.00s)
    --- PASS: TestExerciseHandler_GetExerciseWithSets/invalid_exercise_ID (0.00s)
    --- PASS: TestExerciseHandler_GetExerciseWithSets/service_error (0.00s)
    --- PASS: TestExerciseHandler_GetExerciseWithSets/empty_results_-_returns_200_with_empty_array (0.00s)
    --- PASS: TestExerciseHandler_GetExerciseWithSets/unauthenticated_user (0.00s)
=== RUN   TestExerciseHandler_GetRecentSetsForExercise
=== RUN   TestExerciseHandler_GetRecentSetsForExercise/successful_fetch
=== RUN   TestExerciseHandler_GetRecentSetsForExercise/invalid_exercise_ID
=== RUN   TestExerciseHandler_GetRecentSetsForExercise/service_error
=== RUN   TestExerciseHandler_GetRecentSetsForExercise/unauthenticated_user
--- PASS: TestExerciseHandler_GetRecentSetsForExercise (0.00s)
    --- PASS: TestExerciseHandler_GetRecentSetsForExercise/successful_fetch (0.00s)
    --- PASS: TestExerciseHandler_GetRecentSetsForExercise/invalid_exercise_ID (0.00s)
    --- PASS: TestExerciseHandler_GetRecentSetsForExercise/service_error (0.00s)
    --- PASS: TestExerciseHandler_GetRecentSetsForExercise/unauthenticated_user (0.00s)
=== RUN   TestExerciseHandler_GetOrCreateExercise
=== RUN   TestExerciseHandler_GetOrCreateExercise/successful_creation
=== RUN   TestExerciseHandler_GetOrCreateExercise/invalid_JSON
=== RUN   TestExerciseHandler_GetOrCreateExercise/validation_error
=== RUN   TestExerciseHandler_GetOrCreateExercise/service_error
=== RUN   TestExerciseHandler_GetOrCreateExercise/unauthenticated_user
--- PASS: TestExerciseHandler_GetOrCreateExercise (0.00s)
    --- PASS: TestExerciseHandler_GetOrCreateExercise/successful_creation (0.00s)
    --- PASS: TestExerciseHandler_GetOrCreateExercise/invalid_JSON (0.00s)
    --- PASS: TestExerciseHandler_GetOrCreateExercise/validation_error (0.00s)
    --- PASS: TestExerciseHandler_GetOrCreateExercise/service_error (0.00s)
    --- PASS: TestExerciseHandler_GetOrCreateExercise/unauthenticated_user (0.00s)
=== RUN   TestExerciseHandlerRLSIntegration
=== RUN   TestExerciseHandlerRLSIntegration/Scenario1_UserA_CanRetrieveOwnExercises
=== RUN   TestExerciseHandlerRLSIntegration/Scenario2_UserB_CannotRetrieveUserAExercises
=== RUN   TestExerciseHandlerRLSIntegration/Scenario3_AnonymousUser_CannotAccessExercises
=== RUN   TestExerciseHandlerRLSIntegration/Scenario4_GetSpecificExercise_UserB_CannotAccessUserAExercise
=== RUN   TestExerciseHandlerRLSIntegration/Scenario5_CreateExercise_UserIsolation
=== RUN   TestExerciseHandlerRLSIntegration/Scenario6_ConcurrentRequests_ProperIsolation
--- PASS: TestExerciseHandlerRLSIntegration (0.45s)
    --- PASS: TestExerciseHandlerRLSIntegration/Scenario1_UserA_CanRetrieveOwnExercises (0.00s)        
    --- PASS: TestExerciseHandlerRLSIntegration/Scenario2_UserB_CannotRetrieveUserAExercises (0.00s)   
    --- PASS: TestExerciseHandlerRLSIntegration/Scenario3_AnonymousUser_CannotAccessExercises (0.00s)  
    --- PASS: TestExerciseHandlerRLSIntegration/Scenario4_GetSpecificExercise_UserB_CannotAccessUserAExercise (0.02s)
    --- PASS: TestExerciseHandlerRLSIntegration/Scenario5_CreateExercise_UserIsolation (0.01s)
    --- PASS: TestExerciseHandlerRLSIntegration/Scenario6_ConcurrentRequests_ProperIsolation (0.06s)   
=== RUN   TestExerciseRepository_GetRecentSetsForExercise
--- PASS: TestExerciseRepository_GetRecentSetsForExercise (0.00s)
=== RUN   TestExerciseService_GetRecentSetsForExercise
--- PASS: TestExerciseService_GetRecentSetsForExercise (0.00s)
PASS
ok      github.com/Andrewy-gh/fittrack/server/internal/exercise (cached)
=== RUN   TestSanitizeErrorMessage
=== RUN   TestSanitizeErrorMessage/PostgreSQL_error_code_23505
=== RUN   TestSanitizeErrorMessage/PostgreSQL_error_code_23503
=== RUN   TestSanitizeErrorMessage/pgx_connection_error
=== RUN   TestSanitizeErrorMessage/JWT_parse_error
=== RUN   TestSanitizeErrorMessage/JWT_claims_error
=== RUN   TestSanitizeErrorMessage/authorization_context_error
=== RUN   TestSanitizeErrorMessage/clean_validation_error
=== RUN   TestSanitizeErrorMessage/validation_error_occurred_message
=== RUN   TestSanitizeErrorMessage/failed_to_decode_request_body
=== RUN   TestSanitizeErrorMessage/nil_error
=== RUN   TestSanitizeErrorMessage/unauthorized_database_error
=== RUN   TestSanitizeErrorMessage/auth_database_error
--- PASS: TestSanitizeErrorMessage (0.00s)
    --- PASS: TestSanitizeErrorMessage/PostgreSQL_error_code_23505 (0.00s)
    --- PASS: TestSanitizeErrorMessage/PostgreSQL_error_code_23503 (0.00s)
    --- PASS: TestSanitizeErrorMessage/pgx_connection_error (0.00s)
    --- PASS: TestSanitizeErrorMessage/JWT_parse_error (0.00s)
    --- PASS: TestSanitizeErrorMessage/JWT_claims_error (0.00s)
    --- PASS: TestSanitizeErrorMessage/authorization_context_error (0.00s)
    --- PASS: TestSanitizeErrorMessage/clean_validation_error (0.00s)
    --- PASS: TestSanitizeErrorMessage/validation_error_occurred_message (0.00s)
    --- PASS: TestSanitizeErrorMessage/failed_to_decode_request_body (0.00s)
    --- PASS: TestSanitizeErrorMessage/nil_error (0.00s)
    --- PASS: TestSanitizeErrorMessage/unauthorized_database_error (0.00s)
    --- PASS: TestSanitizeErrorMessage/auth_database_error (0.00s)
=== RUN   TestContainsDatabaseError
=== RUN   TestContainsDatabaseError/PostgreSQL_error_code
=== RUN   TestContainsDatabaseError/pgx_error
=== RUN   TestContainsDatabaseError/Constraint_error
=== RUN   TestContainsDatabaseError/Table_reference
=== RUN   TestContainsDatabaseError/Database_connection
=== RUN   TestContainsDatabaseError/Clean_error
=== RUN   TestContainsDatabaseError/Go-playground_validator
=== RUN   TestContainsDatabaseError/HTTP_error
=== RUN   TestContainsDatabaseError/Empty_string
--- PASS: TestContainsDatabaseError (0.00s)
    --- PASS: TestContainsDatabaseError/PostgreSQL_error_code (0.00s)
    --- PASS: TestContainsDatabaseError/pgx_error (0.00s)
    --- PASS: TestContainsDatabaseError/Constraint_error (0.00s)
    --- PASS: TestContainsDatabaseError/Table_reference (0.00s)
    --- PASS: TestContainsDatabaseError/Database_connection (0.00s)
    --- PASS: TestContainsDatabaseError/Clean_error (0.00s)
    --- PASS: TestContainsDatabaseError/Go-playground_validator (0.00s)
    --- PASS: TestContainsDatabaseError/HTTP_error (0.00s)
    --- PASS: TestContainsDatabaseError/Empty_string (0.00s)
=== RUN   TestIsValidationError
=== RUN   TestIsValidationError/Validation_error_message
=== RUN   TestIsValidationError/Missing_field
=== RUN   TestIsValidationError/Decode_error
=== RUN   TestIsValidationError/Go-playground_validator
=== RUN   TestIsValidationError/Validation_with_database_error
=== RUN   TestIsValidationError/Validation_with_JWT_error
=== RUN   TestIsValidationError/Database_error
=== RUN   TestIsValidationError/JWT_error
=== RUN   TestIsValidationError/Generic_error
--- PASS: TestIsValidationError (0.00s)
    --- PASS: TestIsValidationError/Validation_error_message (0.00s)
    --- PASS: TestIsValidationError/Missing_field (0.00s)
    --- PASS: TestIsValidationError/Decode_error (0.00s)
    --- PASS: TestIsValidationError/Go-playground_validator (0.00s)
    --- PASS: TestIsValidationError/Validation_with_database_error (0.00s)
    --- PASS: TestIsValidationError/Validation_with_JWT_error (0.00s)
    --- PASS: TestIsValidationError/Database_error (0.00s)
    --- PASS: TestIsValidationError/JWT_error (0.00s)
    --- PASS: TestIsValidationError/Generic_error (0.00s)
=== RUN   TestContainsJWTError
=== RUN   TestContainsJWTError/JWT_parse_error
=== RUN   TestContainsJWTError/Token_validation
=== RUN   TestContainsJWTError/JWKS_error
=== RUN   TestContainsJWTError/Claims_error
=== RUN   TestContainsJWTError/Algorithm_error
=== RUN   TestContainsJWTError/Clean_error
=== RUN   TestContainsJWTError/HTTP_error
=== RUN   TestContainsJWTError/Empty_string
--- PASS: TestContainsJWTError (0.01s)
    --- PASS: TestContainsJWTError/JWT_parse_error (0.00s)
    --- PASS: TestContainsJWTError/Token_validation (0.00s)
    --- PASS: TestContainsJWTError/JWKS_error (0.00s)
    --- PASS: TestContainsJWTError/Claims_error (0.00s)
    --- PASS: TestContainsJWTError/Algorithm_error (0.00s)
    --- PASS: TestContainsJWTError/Clean_error (0.00s)
    --- PASS: TestContainsJWTError/HTTP_error (0.00s)
    --- PASS: TestContainsJWTError/Empty_string (0.00s)
=== RUN   TestErrorJSON_NoSensitiveDataInResponse
=== RUN   TestErrorJSON_NoSensitiveDataInResponse/PostgreSQL_constraint_violation
time=2025-08-28T23:13:15.133-04:00 level=ERROR msg="failed to create user" error="pq: duplicate key value violates unique constraint \"users_pkey\" SQLSTATE 23505" path=/api/test method=POST status=500     
time=2025-08-28T23:13:15.135-04:00 level=DEBUG msg="raw error details" error="pq: duplicate key value violates unique constraint \"users_pkey\" SQLSTATE 23505" error_type=*errors.errorString path=/api/test 
=== RUN   TestErrorJSON_NoSensitiveDataInResponse/pgx_connection_error
time=2025-08-28T23:13:15.136-04:00 level=ERROR msg="database unavailable" error="pgx: failed to connect to database pool" path=/api/test method=POST status=500
time=2025-08-28T23:13:15.136-04:00 level=DEBUG msg="raw error details" error="pgx: failed to connect to database pool" error_type=*errors.errorString path=/api/test
=== RUN   TestErrorJSON_NoSensitiveDataInResponse/JWT_token_error
time=2025-08-28T23:13:15.137-04:00 level=ERROR msg="authentication failed" error="failed to parse token: invalid signature algorithm RS256" path=/api/test method=POST status=500
time=2025-08-28T23:13:15.137-04:00 level=DEBUG msg="raw error details" error="failed to parse token: invalid signature algorithm RS256" error_type=*errors.errorString path=/api/test
--- PASS: TestErrorJSON_NoSensitiveDataInResponse (0.00s)
    --- PASS: TestErrorJSON_NoSensitiveDataInResponse/PostgreSQL_constraint_violation (0.00s)
    --- PASS: TestErrorJSON_NoSensitiveDataInResponse/pgx_connection_error (0.00s)
    --- PASS: TestErrorJSON_NoSensitiveDataInResponse/JWT_token_error (0.00s)
=== RUN   TestNoPostgreSQLErrorCodesInResponse
=== RUN   TestNoPostgreSQLErrorCodesInResponse/ErrorCode_23505
time=2025-08-28T23:13:15.137-04:00 level=ERROR msg="operation failed" error="Database error with code: 23505 and additional context" path=/api/test method=GET status=500
time=2025-08-28T23:13:15.137-04:00 level=DEBUG msg="raw error details" error="Database error with code: 23505 and additional context" error_type=*errors.errorString path=/api/test
=== RUN   TestNoPostgreSQLErrorCodesInResponse/ErrorCode_23503
time=2025-08-28T23:13:15.137-04:00 level=ERROR msg="operation failed" error="Database error with code: 23503 and additional context" path=/api/test method=GET status=500
time=2025-08-28T23:13:15.137-04:00 level=DEBUG msg="raw error details" error="Database error with code: 23503 and additional context" error_type=*errors.errorString path=/api/test
=== RUN   TestNoPostgreSQLErrorCodesInResponse/ErrorCode_42501
time=2025-08-28T23:13:15.137-04:00 level=ERROR msg="operation failed" error="Database error with code: 42501 and additional context" path=/api/test method=GET status=500
time=2025-08-28T23:13:15.137-04:00 level=DEBUG msg="raw error details" error="Database error with code: 42501 and additional context" error_type=*errors.errorString path=/api/test
=== RUN   TestNoPostgreSQLErrorCodesInResponse/ErrorCode_42P01
time=2025-08-28T23:13:15.137-04:00 level=ERROR msg="operation failed" error="Database error with code: 42P01 and additional context" path=/api/test method=GET status=500
time=2025-08-28T23:13:15.137-04:00 level=DEBUG msg="raw error details" error="Database error with code: 42P01 and additional context" error_type=*errors.errorString path=/api/test
=== RUN   TestNoPostgreSQLErrorCodesInResponse/ErrorCode_42703
time=2025-08-28T23:13:15.137-04:00 level=ERROR msg="operation failed" error="Database error with code: 42703 and additional context" path=/api/test method=GET status=500
time=2025-08-28T23:13:15.137-04:00 level=DEBUG msg="raw error details" error="Database error with code: 42703 and additional context" error_type=*errors.errorString path=/api/test
=== RUN   TestNoPostgreSQLErrorCodesInResponse/ErrorCode_08006
time=2025-08-28T23:13:15.137-04:00 level=ERROR msg="operation failed" error="Database error with code: 08006 and additional context" path=/api/test method=GET status=500
time=2025-08-28T23:13:15.137-04:00 level=DEBUG msg="raw error details" error="Database error with code: 08006 and additional context" error_type=*errors.errorString path=/api/test
=== RUN   TestNoPostgreSQLErrorCodesInResponse/ErrorCode_57P01
time=2025-08-28T23:13:15.137-04:00 level=ERROR msg="operation failed" error="Database error with code: 57P01 and additional context" path=/api/test method=GET status=500
time=2025-08-28T23:13:15.138-04:00 level=DEBUG msg="raw error details" error="Database error with code: 57P01 and additional context" error_type=*errors.errorString path=/api/test
=== RUN   TestNoPostgreSQLErrorCodesInResponse/ErrorCode_pgx:_connection
time=2025-08-28T23:13:15.138-04:00 level=ERROR msg="operation failed" error="Database error with code: pgx: connection and additional context" path=/api/test method=GET status=500
time=2025-08-28T23:13:15.138-04:00 level=DEBUG msg="raw error details" error="Database error with code: pgx: connection and additional context" error_type=*errors.errorString path=/api/test
=== RUN   TestNoPostgreSQLErrorCodesInResponse/ErrorCode_pq:_duplicate_key
time=2025-08-28T23:13:15.138-04:00 level=ERROR msg="operation failed" error="Database error with code: pq: duplicate key and additional context" path=/api/test method=GET status=500
time=2025-08-28T23:13:15.138-04:00 level=DEBUG msg="raw error details" error="Database error with code: pq: duplicate key and additional context" error_type=*errors.errorString path=/api/test
=== RUN   TestNoPostgreSQLErrorCodesInResponse/ErrorCode_SQLSTATE
time=2025-08-28T23:13:15.138-04:00 level=ERROR msg="operation failed" error="Database error with code: SQLSTATE and additional context" path=/api/test method=GET status=500
time=2025-08-28T23:13:15.138-04:00 level=DEBUG msg="raw error details" error="Database error with code: SQLSTATE and additional context" error_type=*errors.errorString path=/api/test
=== RUN   TestNoPostgreSQLErrorCodesInResponse/ErrorCode_constraint_violation
time=2025-08-28T23:13:15.138-04:00 level=ERROR msg="operation failed" error="Database error with code: constraint violation and additional context" path=/api/test method=GET status=500
time=2025-08-28T23:13:15.138-04:00 level=DEBUG msg="raw error details" error="Database error with code: constraint violation and additional context" error_type=*errors.errorString path=/api/test
=== RUN   TestNoPostgreSQLErrorCodesInResponse/ErrorCode_relation_does_not_exist
time=2025-08-28T23:13:15.139-04:00 level=ERROR msg="operation failed" error="Database error with code: relation does not exist and additional context" path=/api/test method=GET status=500
time=2025-08-28T23:13:15.139-04:00 level=DEBUG msg="raw error details" error="Database error with code: relation does not exist and additional context" error_type=*errors.errorString path=/api/test
--- PASS: TestNoPostgreSQLErrorCodesInResponse (0.00s)
    --- PASS: TestNoPostgreSQLErrorCodesInResponse/ErrorCode_23505 (0.00s)
    --- PASS: TestNoPostgreSQLErrorCodesInResponse/ErrorCode_23503 (0.00s)
    --- PASS: TestNoPostgreSQLErrorCodesInResponse/ErrorCode_42501 (0.00s)
    --- PASS: TestNoPostgreSQLErrorCodesInResponse/ErrorCode_42P01 (0.00s)
    --- PASS: TestNoPostgreSQLErrorCodesInResponse/ErrorCode_42703 (0.00s)
    --- PASS: TestNoPostgreSQLErrorCodesInResponse/ErrorCode_08006 (0.00s)
    --- PASS: TestNoPostgreSQLErrorCodesInResponse/ErrorCode_57P01 (0.00s)
    --- PASS: TestNoPostgreSQLErrorCodesInResponse/ErrorCode_pgx:_connection (0.00s)
    --- PASS: TestNoPostgreSQLErrorCodesInResponse/ErrorCode_pq:_duplicate_key (0.00s)
    --- PASS: TestNoPostgreSQLErrorCodesInResponse/ErrorCode_SQLSTATE (0.00s)
    --- PASS: TestNoPostgreSQLErrorCodesInResponse/ErrorCode_constraint_violation (0.00s)
    --- PASS: TestNoPostgreSQLErrorCodesInResponse/ErrorCode_relation_does_not_exist (0.00s)
PASS
ok      github.com/Andrewy-gh/fittrack/server/internal/response (cached)
testing: warning: no tests to run
PASS
ok      github.com/Andrewy-gh/fittrack/server/internal/testutils        (cached) [no tests to run]     
?       github.com/Andrewy-gh/fittrack/server/internal/user     [no test files]
FAIL
make: *** [Makefile:79: test] Error 1