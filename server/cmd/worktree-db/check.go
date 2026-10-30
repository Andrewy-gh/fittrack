package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// check never creates roles/databases, takes setup locks, or runs migrations.
// It shares connect's local endpoint, role, Docker ownership and version guards.
func check(mode string) string {
	deadline, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	ctx = deadline
	var cfg allocation
	if json.Unmarshal([]byte(os.Getenv("FITTRACK_DATABASE_CONFIG")), &cfg) != nil ||
		!regexp.MustCompile(`^[a-z0-9_]+_[a-f0-9]{12}$`).MatchString(cfg.ID) || cfg.Postgres.Host != "127.0.0.1" || cfg.Postgres.Port != 55432 {
		return "invalid_configuration"
	}
	major, err := strconv.Atoi(os.Getenv("FITTRACK_POSTGRES_MAJOR"))
	if err != nil || major != 15 {
		return "invalid_configuration"
	}
	if mode != "bootstrap" && mode != "dev" && mode != "test" && mode != "rls" && mode != "e2e" {
		return "invalid_configuration"
	}
	ownerURL, err := localURL(os.Getenv("FITTRACK_LOCAL_DATABASE_URL"), "postgres")
	if err != nil {
		return "invalid_configuration"
	}
	runtimeURL, err := localURL(os.Getenv("FITTRACK_LOCAL_RUNTIME_DATABASE_URL"), "postgres")
	if err != nil {
		return "invalid_configuration"
	}
	owner, _ := url.Parse(ownerURL)
	runtime, _ := url.Parse(runtimeURL)
	if owner.User.Username() == "fittrack_app" || runtime.User.Username() != "fittrack_app" {
		return "invalid_configuration"
	}
	admin, err := connect(ownerURL, "postgres", major)
	if err != nil {
		return "cluster_unavailable"
	}
	defer admin.Close(ctx)
	var createDB, createRole, roleExists bool
	if admin.QueryRow(ctx, `SELECT rolcreatedb OR rolsuper, rolcreaterole OR rolsuper,
 EXISTS(SELECT 1 FROM pg_roles WHERE rolname='fittrack_app') FROM pg_roles WHERE rolname=current_user`).Scan(&createDB, &createRole, &roleExists) != nil || !createDB || (!roleExists && !createRole) {
		return "owner_permissions"
	}
	if !roleExists {
		password, supplied := runtime.User.Password()
		if mode == "bootstrap" && supplied && password != "" {
			return "ready" // Setup will create the role; credentials are not yet authenticated.
		}
		return "runtime_auth"
	}
	var safe bool
	if admin.QueryRow(ctx, `SELECT NOT EXISTS(SELECT 1 FROM pg_roles WHERE rolname='fittrack_app' AND
 (rolsuper OR rolcreatedb OR rolcreaterole OR rolinherit OR rolreplication OR rolbypassrls OR NOT rolcanlogin))
 AND NOT EXISTS(SELECT 1 FROM pg_auth_members WHERE member=(SELECT oid FROM pg_roles WHERE rolname='fittrack_app'))`).Scan(&safe) != nil || !safe {
		return "unsafe_runtime_role"
	}
	role, err := connect(runtimeURL, "postgres", major)
	if err != nil {
		return "runtime_auth"
	}
	role.Close(ctx)
	if mode == "bootstrap" {
		return "ready"
	}
	name := cfg.Databases[mode]
	if name != "ft_"+cfg.ID+"_"+mode || len(name) > 63 {
		return "invalid_configuration"
	}
	var actualOwner, marker string
	if admin.QueryRow(ctx, "SELECT pg_get_userbyid(datdba), coalesce(shobj_description(oid,'pg_database'),'') FROM pg_database WHERE datname=$1", name).Scan(&actualOwner, &marker) != nil {
		return "database_unprepared"
	}
	if actualOwner != owner.User.Username() || marker != "fittrack-worktree:"+cfg.ID {
		return "database_ownership"
	}
	connection, err := connect(ownerURL, name, major)
	if err != nil {
		return "database_unprepared"
	}
	defer connection.Close(ctx)
	rows, err := connection.Query(ctx, "SELECT version_id, is_applied FROM goose_db_version ORDER BY id")
	if err != nil {
		return "migrations_pending"
	}
	versions := map[int64]bool{}
	for rows.Next() {
		var version int64
		var applied bool
		if rows.Scan(&version, &applied) != nil {
			rows.Close()
			return "migrations_pending"
		}
		versions[version] = applied
	}
	rows.Close()
	if rows.Err() != nil {
		return "migrations_pending"
	}
	files, err := filepath.Glob(filepath.Join(cfg.Root, "server", "migrations", "*.sql"))
	if err != nil || len(files) == 0 {
		return "invalid_configuration"
	}
	for _, file := range files {
		version, err := strconv.ParseInt(strings.SplitN(filepath.Base(file), "_", 2)[0], 10, 64)
		if err != nil || !versions[version] {
			return "migrations_pending"
		}
	}
	if mode == "rls" {
		runtimeDB, err := connect(runtimeURL, name, major)
		if err != nil {
			return "runtime_auth"
		}
		runtimeDB.Close(ctx)
	}
	return "ready"
}

func reportCheck(mode string) {
	code := check(mode)
	// Only fixed codes cross this boundary; driver/DSN diagnostics remain private.
	output, _ := json.Marshal(struct {
		Ready bool   `json:"ready"`
		Code  string `json:"code"`
	}{code == "ready", code})
	fmt.Println(string(output))
}
