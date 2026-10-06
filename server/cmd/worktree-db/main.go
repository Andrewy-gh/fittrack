// worktree-db is invoked by scripts/worktree.ts with the allocation in memory.
// It never installs, starts, upgrades, drops, or reconfigures PostgreSQL clusters.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type allocation struct {
	ID        string            `json:"id"`
	Root      string            `json:"root"`
	Databases map[string]string `json:"databases"`
	Postgres  struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"postgres"`
}

var ctx = context.Background()

func localURL(value, database string) (string, error) {
	u, err := url.Parse(value)
	if err != nil || u == nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Host != "127.0.0.1:55432" || u.Fragment != "" || u.User == nil || u.User.Username() == "" {
		return "", errors.New("local connection must explicitly name a role at 127.0.0.1:55432")
	}
	for key := range u.Query() {
		if key != "sslmode" && key != "connect_timeout" {
			return "", errors.New("connection override parameters are refused")
		}
	}
	u.Path = "/" + database
	return u.String(), nil
}

// Docker is opt-in and inspected only through this machine's local daemon.
// A context, DOCKER_HOST, or TCP daemon must never redirect this ownership check.
func dockerEndpoint() (map[string]bool, error) {
	container := os.Getenv("FITTRACK_LOCAL_POSTGRES_CONTAINER")
	if !regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`).MatchString(container) {
		return nil, errors.New("invalid local PostgreSQL container name")
	}
	endpoint := "unix:///var/run/docker.sock"
	if runtime.GOOS == "windows" {
		endpoint = "npipe:////./pipe/docker_engine"
	}
	deadline, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	command := exec.CommandContext(deadline, "docker", "--host", endpoint, "inspect", "--type", "container", container)
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(strings.ToUpper(entry), "DOCKER_") {
			command.Env = append(command.Env, entry)
		}
	}
	output, err := command.Output()
	if err != nil {
		return nil, errors.New("cannot inspect PostgreSQL container on the local Docker daemon")
	}
	var entries []struct {
		Config struct {
			Image  string
			Labels map[string]string
		}
		State           struct{ Running bool }
		NetworkSettings struct {
			Ports map[string][]struct {
				HostIp   string
				HostPort string
			}
			Networks map[string]struct{ IPAddress string }
		}
	}
	if json.Unmarshal(output, &entries) != nil || len(entries) != 1 {
		return nil, errors.New("invalid local Docker inspection")
	}
	c := entries[0]
	bindings := c.NetworkSettings.Ports["5432/tcp"]
	const image = "postgres:15.19-bookworm@sha256:539ceaaae49b3a7c8a04467cf00cc6788d8e3f1675df41860d86eebc4c40524f"
	if !c.State.Running || c.Config.Image != image || c.Config.Labels["fittrack.local-postgres"] != "15.19" || len(bindings) != 1 || bindings[0].HostIp != "127.0.0.1" || bindings[0].HostPort != "55432" {
		return nil, errors.New("refusing unowned, unpinned, or incorrectly published local PostgreSQL container")
	}
	addresses := map[string]bool{}
	for _, network := range c.NetworkSettings.Networks {
		if network.IPAddress != "" {
			addresses[network.IPAddress] = true
		}
	}
	return addresses, nil
}

func connect(connection, database string, major int) (*pgx.Conn, error) {
	var dockerAddresses map[string]bool
	var err error
	if os.Getenv("FITTRACK_LOCAL_POSTGRES_CONTAINER") != "" {
		dockerAddresses, err = dockerEndpoint()
		if err != nil {
			return nil, err
		}
	}
	cfg, err := pgx.ParseConfig(connection)
	if err != nil {
		return nil, errors.New("invalid local connection")
	}
	// Never let inherited libpq fallback settings redirect the endpoint.
	cfg.Host = "127.0.0.1"
	cfg.Port = 55432
	cfg.Database = database
	cfg.Fallbacks = nil
	cfg.ConnectTimeout = 10 * time.Second
	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return nil, errors.New("local database connection failed (credentials withheld)")
	}
	var host, name, role string
	var port, version int
	err = conn.QueryRow(ctx, "SELECT host(inet_server_addr()), inet_server_port(), current_database(), current_user, current_setting('server_version_num')::int").Scan(&host, &port, &name, &role, &version)
	endpointMatches := host == "127.0.0.1" && port == 55432
	if dockerAddresses != nil {
		endpointMatches = dockerAddresses[host] && port == 5432 && version == 150019
	}
	if err != nil || !endpointMatches || name != database || role != cfg.User || version/10000 != major {
		conn.Close(ctx)
		return nil, errors.New("server endpoint, database, role, or explicit PostgreSQL major does not match")
	}
	return conn, nil
}

func execute(conn *pgx.Conn, sql string, args ...any) error {
	_, err := conn.Exec(ctx, sql, args...)
	if err != nil {
		return errors.New("local SQL operation failed; server details withheld to protect credentials")
	}
	return nil
}

func setup() error {
	var cfg allocation
	if err := json.Unmarshal([]byte(os.Getenv("FITTRACK_DATABASE_CONFIG")), &cfg); err != nil {
		return errors.New("invoke through node scripts/worktree.ts db setup")
	}
	if !regexp.MustCompile(`^[a-z0-9_]+_[a-f0-9]{12}$`).MatchString(cfg.ID) || cfg.Postgres.Host != "127.0.0.1" || cfg.Postgres.Port != 55432 {
		return errors.New("invalid worktree allocation")
	}
	major, err := strconv.Atoi(os.Getenv("FITTRACK_POSTGRES_MAJOR"))
	if err != nil || major < 10 || major > 99 {
		return errors.New("an explicit verified PostgreSQL major is required")
	}
	modes := []string{"dev", "test", "rls", "e2e"}
	for _, mode := range modes {
		if cfg.Databases[mode] != "ft_"+cfg.ID+"_"+mode || len(cfg.Databases[mode]) > 63 {
			return errors.New("database name does not match allocation")
		}
	}
	ownerURL, err := localURL(os.Getenv("FITTRACK_LOCAL_DATABASE_URL"), "postgres")
	if err != nil {
		return err
	}
	owner, _ := url.Parse(ownerURL)
	if owner.User.Username() == "fittrack_app" {
		return errors.New("database setup needs an owning role, not fittrack_app")
	}
	runtimeURL, err := localURL(os.Getenv("FITTRACK_LOCAL_RUNTIME_DATABASE_URL"), "postgres")
	if err != nil {
		return errors.New("set FITTRACK_LOCAL_RUNTIME_DATABASE_URL for the stable local fittrack_app login")
	}
	runtime, _ := url.Parse(runtimeURL)
	if runtime.User.Username() != "fittrack_app" {
		return errors.New("runtime connection must use fittrack_app")
	}
	goose := os.Getenv("FITTRACK_GOOSE")
	if goose == "" {
		goose = "goose"
	}
	version, err := exec.Command(goose, "-version").Output()
	if err != nil || strings.TrimSpace(string(version)) != "goose version: v3.24.3" {
		return errors.New("Goose v3.24.3 is required; set FITTRACK_GOOSE to its executable")
	}
	admin, err := connect(ownerURL, "postgres", major)
	if err != nil {
		return err
	}
	defer admin.Close(ctx)
	// The postgres-database advisory lock coordinates catalog/role setup across clones too.
	var locked bool
	if err = admin.QueryRow(ctx, "SELECT pg_try_advisory_lock(714624891036)").Scan(&locked); err != nil || !locked {
		return errors.New("another local database setup holds the cluster lock; retry when it finishes")
	}
	defer admin.Exec(ctx, "SELECT pg_advisory_unlock(714624891036)")
	var canCreate bool
	if err = admin.QueryRow(ctx, "SELECT rolcreatedb OR rolsuper FROM pg_roles WHERE rolname=current_user").Scan(&canCreate); err != nil || !canCreate {
		return errors.New("local owner requires CREATEDB")
	}
	// Preflight every existing database before changing any of them.
	marker := "fittrack-worktree:" + cfg.ID
	existing := map[string]bool{}
	for _, mode := range modes {
		name := cfg.Databases[mode]
		var role, comment string
		err = admin.QueryRow(ctx, "SELECT pg_get_userbyid(datdba), coalesce(shobj_description(oid,'pg_database'),'') FROM pg_database WHERE datname=$1", name).Scan(&role, &comment)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil || role != owner.User.Username() || comment != marker {
			return fmt.Errorf("refusing unowned database %s (owner/marker mismatch)", name)
		}
		existing[name] = true
	}
	var roleExists, safe bool
	err = admin.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_roles WHERE rolname='fittrack_app'),
 NOT EXISTS(SELECT 1 FROM pg_roles WHERE rolname='fittrack_app' AND
 (rolsuper OR rolcreatedb OR rolcreaterole OR rolinherit OR rolreplication OR rolbypassrls OR NOT rolcanlogin))
 AND NOT EXISTS(SELECT 1 FROM pg_auth_members WHERE member=(SELECT oid FROM pg_roles WHERE rolname='fittrack_app'))`).Scan(&roleExists, &safe)
	if err != nil || !safe {
		return errors.New("existing fittrack_app is unsafe; refusing shared-role changes")
	}
	if !roleExists {
		password, supplied := runtime.User.Password()
		if !supplied || password == "" {
			return errors.New("first setup requires a local runtime password; subsequent setup never resets it")
		}
		var literal string
		if err = admin.QueryRow(ctx, "SELECT quote_literal($1::text)", password).Scan(&literal); err != nil {
			return errors.New("cannot prepare local role")
		}
		if err = execute(admin, "CREATE ROLE fittrack_app LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS PASSWORD "+literal); err != nil {
			return err
		}
	}
	runtimeConnection, err := connect(runtimeURL, "postgres", major)
	if err != nil {
		return errors.New("fittrack_app authentication failed; no existing password was changed")
	}
	runtimeConnection.Close(ctx)
	grants, err := os.ReadFile(filepath.Join(cfg.Root, "server", "scripts", "provision-runtime-role.sql"))
	if err != nil {
		return errors.New("cannot read runtime grants")
	}
	for _, mode := range modes {
		name := cfg.Databases[mode]
		identifier := pgx.Identifier{name}.Sanitize()
		if !existing[name] {
			if err = execute(admin, "CREATE DATABASE "+identifier+" OWNER "+pgx.Identifier{owner.User.Username()}.Sanitize()); err != nil {
				return fmt.Errorf("create %s: %v", name, err)
			}
			// An interruption before marking deliberately requires manual ownership recovery.
			if err = execute(admin, "COMMENT ON DATABASE "+identifier+" IS '"+marker+"'"); err != nil {
				return err
			}
		}
		connection, _ := localURL(ownerURL, name)
		db, err := connect(connection, name, major)
		if err != nil {
			return err
		}
		db.Close(ctx)
		migration := exec.Command(goose, "-dir", filepath.Join(cfg.Root, "server", "migrations"), "up")
		migration.Env = append(os.Environ(), "GOOSE_DRIVER=postgres", "GOOSE_DBSTRING="+connection)
		// Goose errors can contain a DSN. Never stream them or put URLs in argv.
		if err = migration.Run(); err != nil {
			return fmt.Errorf("migrate %s failed (%v); data preserved, connection details withheld", name, err)
		}
		db, err = connect(connection, name, major)
		if err != nil {
			return err
		}
		if err = execute(db, "SET fittrack.preserve_runtime_role = 'on'"); err == nil {
			err = execute(db, string(grants))
		}
		db.Close(ctx)
		if err != nil {
			return fmt.Errorf("grants %s: %v", name, err)
		}
		fmt.Printf("Ready: %s (PostgreSQL %d)\n", name, major)
	}
	return nil
}

func main() {
	if err := setup(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
