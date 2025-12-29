#!/usr/bin/env bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
MAIN_WORKTREE="C:/E/2025/fittrack"
WORKTREE_BASE_DIR="C:/E/2025"
PROJECT_NAME="fittrack"

# Print colored message
print_info() {
    echo -e "${BLUE}ℹ ${NC}$1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_section() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

# Usage
usage() {
    cat << EOF
Usage: $0 <branch-name> [options]

Creates a new git worktree with automated setup.

Arguments:
    branch-name         Name of the branch (will create worktree at ${WORKTREE_BASE_DIR}/${PROJECT_NAME}-<branch-name>)

Options:
    --shared-db         Use the same PostgreSQL database (default)
    --new-db            Create a new PostgreSQL database instance
    --skip-install      Skip dependency installation
    --skip-migrations   Skip database migrations
    -h, --help          Show this help message

Examples:
    $0 feat-pagination
    $0 fix-auth-bug --new-db
    $0 refactor-api --skip-migrations

EOF
    exit 1
}

# Parse arguments
BRANCH_NAME=""
DB_MODE="shared"  # default to shared database
SKIP_INSTALL=false
SKIP_MIGRATIONS=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            ;;
        --shared-db)
            DB_MODE="shared"
            shift
            ;;
        --new-db)
            DB_MODE="new"
            shift
            ;;
        --skip-install)
            SKIP_INSTALL=true
            shift
            ;;
        --skip-migrations)
            SKIP_MIGRATIONS=true
            shift
            ;;
        -*)
            print_error "Unknown option: $1"
            usage
            ;;
        *)
            if [[ -z "$BRANCH_NAME" ]]; then
                BRANCH_NAME="$1"
            else
                print_error "Multiple branch names provided"
                usage
            fi
            shift
            ;;
    esac
done

# Validate branch name
if [[ -z "$BRANCH_NAME" ]]; then
    print_error "Branch name is required"
    usage
fi

# Set worktree path
WORKTREE_PATH="${WORKTREE_BASE_DIR}/${PROJECT_NAME}-${BRANCH_NAME}"

# Check if worktree already exists
if [[ -d "$WORKTREE_PATH" ]]; then
    print_error "Worktree already exists at: $WORKTREE_PATH"
    exit 1
fi

# Prompt for database mode if not specified
if [[ "$DB_MODE" == "shared" ]]; then
    print_section "Database Setup"
    print_info "Database mode: SHARED (using existing database)"
    print_warning "All worktrees will share the same database data"
    echo ""
else
    print_section "Database Setup"
    print_info "Database mode: NEW (creating separate database)"
    print_warning "This worktree will have its own database instance"
    echo ""
fi

print_section "Creating Worktree"

# Navigate to main worktree
cd "$MAIN_WORKTREE"

# Check if branch exists remotely
if git show-ref --verify --quiet "refs/remotes/origin/$BRANCH_NAME"; then
    print_info "Branch '$BRANCH_NAME' exists remotely, checking out..."
    git worktree add "$WORKTREE_PATH" "$BRANCH_NAME"
else
    print_info "Creating new branch '$BRANCH_NAME'..."
    git worktree add -b "$BRANCH_NAME" "$WORKTREE_PATH"
fi

print_success "Worktree created at: $WORKTREE_PATH"

print_section "Copying Configuration Files"

# Copy .claude directory
if [[ -d "$MAIN_WORKTREE/.claude" ]]; then
    cp -r "$MAIN_WORKTREE/.claude" "$WORKTREE_PATH/.claude"
    print_success "Copied .claude/ directory"
else
    print_warning ".claude/ directory not found in main worktree"
fi

# Copy ai-dev-tasks directory
if [[ -d "$MAIN_WORKTREE/ai-dev-tasks" ]]; then
    cp -r "$MAIN_WORKTREE/ai-dev-tasks" "$WORKTREE_PATH/ai-dev-tasks"
    print_success "Copied ai-dev-tasks/ directory"
else
    print_warning "ai-dev-tasks/ directory not found in main worktree"
fi

# Copy client .env
if [[ -f "$MAIN_WORKTREE/client/.env" ]]; then
    mkdir -p "$WORKTREE_PATH/client"
    cp "$MAIN_WORKTREE/client/.env" "$WORKTREE_PATH/client/.env"
    print_success "Copied client/.env"
else
    print_warning "client/.env not found in main worktree"
fi

# Copy server .env
if [[ -f "$MAIN_WORKTREE/server/.env" ]]; then
    mkdir -p "$WORKTREE_PATH/server"
    cp "$MAIN_WORKTREE/server/.env" "$WORKTREE_PATH/server/.env"
    print_success "Copied server/.env"
else
    print_warning "server/.env not found in main worktree"
fi

# Copy server setenv.sh
if [[ -f "$MAIN_WORKTREE/server/setenv.sh" ]]; then
    mkdir -p "$WORKTREE_PATH/server"
    cp "$MAIN_WORKTREE/server/setenv.sh" "$WORKTREE_PATH/server/setenv.sh"
    print_success "Copied server/setenv.sh"
else
    print_warning "server/setenv.sh not found in main worktree"
fi

# Copy all CLAUDE.local.md files
print_info "Looking for CLAUDE.local.md files..."
CLAUDE_FILES_FOUND=false
while IFS= read -r -d '' file; do
    # Get relative path from main worktree
    REL_PATH="${file#$MAIN_WORKTREE/}"
    TARGET_FILE="$WORKTREE_PATH/$REL_PATH"
    TARGET_DIR=$(dirname "$TARGET_FILE")

    # Create directory if it doesn't exist
    mkdir -p "$TARGET_DIR"

    # Copy the file
    cp "$file" "$TARGET_FILE"
    print_success "Copied $REL_PATH"
    CLAUDE_FILES_FOUND=true
done < <(find "$MAIN_WORKTREE" -type f -name "CLAUDE.local.md" -not -path "*/node_modules/*" -not -path "*/.git/*" -print0)

if [[ "$CLAUDE_FILES_FOUND" == false ]]; then
    print_info "No CLAUDE.local.md files found"
fi

# Handle database setup
print_section "Database Configuration"

if [[ "$DB_MODE" == "new" ]]; then
    print_info "Setting up new database instance..."

    # Copy db-data if it exists and has data
    if [[ -d "$MAIN_WORKTREE/server/db-data" ]] && [[ ! -z "$(ls -A "$MAIN_WORKTREE/server/db-data" 2>/dev/null)" ]]; then
        print_info "Copying database data from main worktree..."
        mkdir -p "$WORKTREE_PATH/server/db-data"
        cp -r "$MAIN_WORKTREE/server/db-data"/* "$WORKTREE_PATH/server/db-data/"
        print_success "Database data copied"
    else
        print_info "No existing database data found - will start fresh"
        mkdir -p "$WORKTREE_PATH/server/db-data"
    fi

    print_warning "NOTE: You'll need to use a different DB_PORT in server/.env to avoid conflicts"
    print_warning "      Example: Change DB_PORT=5432 to DB_PORT=5433"
else
    print_info "Using shared database - symlinking db-data directory..."

    # Create symlink to main worktree's db-data
    if [[ -d "$MAIN_WORKTREE/server/db-data" ]]; then
        mkdir -p "$WORKTREE_PATH/server"
        # Remove if exists (in case it was copied)
        rm -rf "$WORKTREE_PATH/server/db-data"
        # Create symlink (Windows-compatible)
        cmd //c mklink //D "$(cygpath -w "$WORKTREE_PATH/server/db-data")" "$(cygpath -w "$MAIN_WORKTREE/server/db-data")" > /dev/null 2>&1 || \
        ln -s "$MAIN_WORKTREE/server/db-data" "$WORKTREE_PATH/server/db-data"
        print_success "Symlinked db-data to main worktree"
    else
        print_warning "db-data directory not found in main worktree"
    fi
fi

# Install dependencies
if [[ "$SKIP_INSTALL" == false ]]; then
    print_section "Installing Dependencies"

    # Install client dependencies
    if [[ -f "$WORKTREE_PATH/client/package.json" ]]; then
        print_info "Installing client dependencies with bun..."
        cd "$WORKTREE_PATH/client"
        bun install
        print_success "Client dependencies installed"
    else
        print_warning "client/package.json not found"
    fi

    # Install server dependencies
    if [[ -f "$WORKTREE_PATH/server/go.mod" ]]; then
        print_info "Installing server dependencies with go mod..."
        cd "$WORKTREE_PATH/server"
        go mod download
        print_success "Server dependencies installed"
    else
        print_warning "server/go.mod not found"
    fi
else
    print_section "Skipping Dependencies"
    print_warning "Dependency installation skipped"
fi

# Run migrations
if [[ "$SKIP_MIGRATIONS" == false ]]; then
    print_section "Running Database Migrations"

    if [[ -f "$WORKTREE_PATH/server/Makefile" ]]; then
        print_info "Running database migrations..."
        cd "$WORKTREE_PATH/server"

        # Check if database is running
        if docker ps | grep -q "db"; then
            make migrate-up
            print_success "Migrations completed"
        else
            print_warning "Database container not running"
            print_info "Start the database with: cd $WORKTREE_PATH/server && docker compose up -d"
            print_info "Then run migrations with: cd $WORKTREE_PATH/server && make migrate-up"
        fi
    else
        print_warning "server/Makefile not found"
    fi
else
    print_section "Skipping Migrations"
    print_warning "Database migrations skipped"
fi

# Summary
print_section "Setup Complete!"

echo -e "${GREEN}Your new worktree is ready at:${NC}"
echo -e "  ${BLUE}$WORKTREE_PATH${NC}"
echo ""
echo -e "${GREEN}Next steps:${NC}"
echo -e "  1. ${BLUE}cd $WORKTREE_PATH${NC}"

if [[ "$DB_MODE" == "new" ]]; then
    echo -e "  2. Update ${BLUE}server/.env${NC} with a different DB_PORT (e.g., 5433)"
    echo -e "  3. Start database: ${BLUE}cd server && docker compose up -d${NC}"
    echo -e "  4. Run migrations: ${BLUE}make migrate-up${NC}"
    echo -e "  5. Start dev servers:"
else
    echo -e "  2. Start dev servers:"
fi

echo -e "     - Client: ${BLUE}cd client && bun run dev${NC}"
echo -e "     - Server: ${BLUE}cd server && make dev${NC}"
echo ""

print_success "Happy coding! 🚀"
