# Git Worktree Setup Guide

This project uses the `setup-worktree.sh` script to automate the creation and configuration of new git worktrees.

## What Gets Copied

When you create a new worktree, the following files and directories are automatically copied from your main worktree:

- `.claude/` - Claude Code configuration and commands
- `ai-dev-tasks/` - Development task workflows
- `client/.env` - Client environment variables
- `server/.env` - Server environment variables
- `server/setenv.sh` - Server environment export script
- All `CLAUDE.local.md` files (recursively, excluding node_modules)

## Database Options

### Shared Database (Default)

All worktrees share the same PostgreSQL database instance.

**Pros:**
- Single source of truth for data
- No port conflicts
- Easier to manage

**Cons:**
- Schema changes affect all worktrees
- Cannot test different database states simultaneously

### New Database

Each worktree has its own PostgreSQL database instance.

**Pros:**
- Isolated testing environments
- Can run different schema versions
- Safe for destructive operations

**Cons:**
- Requires different ports (must edit `server/.env`)
- Uses more disk space
- Requires managing multiple Docker containers

## Usage

### Basic Usage

Create a new worktree with a shared database:

```bash
./setup-worktree.sh feat-pagination
```

This creates a worktree at `C:/E/2025/fittrack-feat-pagination`

### Advanced Options

```bash
# Create worktree with a new database
./setup-worktree.sh feat-auth --new-db

# Create worktree but skip dependency installation
./setup-worktree.sh fix-bug --skip-install

# Create worktree but skip migrations
./setup-worktree.sh refactor --skip-migrations

# Combine options
./setup-worktree.sh feat-dashboard --new-db --skip-migrations
```

### Get Help

```bash
./setup-worktree.sh --help
```

## What Happens During Setup

1. **Creates Git Worktree** - Creates a new worktree with the specified branch name
2. **Copies Configuration** - Copies all specified files and directories
3. **Database Setup**:
   - **Shared**: Creates symlink to main worktree's `db-data`
   - **New**: Copies `db-data` or starts fresh
4. **Install Dependencies**:
   - Client: `bun install`
   - Server: `go mod download`
5. **Run Migrations** - Applies database migrations using `make migrate-up`

## After Setup

### For Shared Database

```bash
cd C:/E/2025/fittrack-feat-pagination

# Start client
cd client && bun run dev

# Start server (in another terminal)
cd server && make dev
```

### For New Database

```bash
cd C:/E/2025/fittrack-feat-pagination

# 1. Update server/.env with different port
# Change: DB_PORT=5432
# To:     DB_PORT=5433

# 2. Update DATABASE_URL in server/.env and server/setenv.sh
# Change: postgresql://user:password@localhost:5432/postgres
# To:     postgresql://user:password@localhost:5433/postgres

# 3. Start database
cd server && docker compose up -d

# 4. Run migrations
make migrate-up

# 5. Start client
cd ../client && bun run dev

# 6. Start server (in another terminal)
cd server && make dev
```

## Managing Worktrees

### List All Worktrees

```bash
git worktree list
```

### Remove a Worktree

```bash
# Remove the worktree directory
git worktree remove C:/E/2025/fittrack-feat-pagination

# Or if you've already deleted the directory
git worktree prune
```

### Switch Between Worktrees

Each worktree is just a directory, so you can have multiple terminal sessions open:

- Terminal 1: `C:/E/2025/fittrack` (main)
- Terminal 2: `C:/E/2025/fittrack-feat-pagination`
- Terminal 3: `C:/E/2025/fittrack-feat-auth`

Or use your IDE to open different worktrees in separate windows.

## Tips

1. **Database Management**: If using shared database, be careful with schema changes that might break other worktrees
2. **Port Conflicts**: If you get port conflicts, check your `.env` files and Docker containers
3. **Clean Up**: Regularly remove worktrees you're no longer using to save disk space
4. **Node Modules**: Each worktree has its own `node_modules`, so they can have different dependency versions
5. **Docker Containers**: When using new databases, remember to stop the containers when done:
   ```bash
   cd server && docker compose down
   ```

## Troubleshooting

### Migration Fails

Make sure the database is running:
```bash
docker ps | grep db
```

If not running:
```bash
cd server && docker compose up -d
```

### Port Already in Use

If you're using `--new-db`, update the `DB_PORT` in `server/.env` to a different port (e.g., 5433, 5434).

### Symlink Creation Failed (Windows)

If the symlink creation fails on Windows, you may need to:
1. Run Git Bash as Administrator, or
2. Enable Developer Mode in Windows Settings

### Dependencies Not Installing

Make sure you have the required tools installed:
- `bun` - for client dependencies
- `go` - for server dependencies
- `docker` - for database

## Files Not Tracked by Git

These files are typically in `.gitignore` but are important for local development, which is why the script copies them:

- `.env` files - contain sensitive credentials
- `CLAUDE.local.md` files - local documentation
- `db-data/` - database files (can be shared or copied)

## Customizing the Script

You can modify `setup-worktree.sh` to customize:

- Worktree location (change `WORKTREE_BASE_DIR`)
- Files to copy (add more copy commands)
- Setup steps (add custom initialization)
- Default behavior (change `DB_MODE` default)

Edit the script at: `C:/E/2025/fittrack/setup-worktree.sh`
