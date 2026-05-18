#!/usr/bin/env bash

# d (dev runner / deployer)
# ./d         local dev (watch + rebuild)
# ./d -eploy  production deploy (build + systemd)

# fail early fail friendly
set -euo pipefail

# colors
# stdio check
if [[ -t 1 ]]; then
    R=$(tput setaf 1)
    G=$(tput setaf 2)
    Y=$(tput setaf 3)
    B=$(tput setaf 4)
    P=$(tput setaf 5)
    W=$(tput setaf 7)
    GR=$(tput setaf 8)
    BD=$(tput bold)
    DM=$(tput dim)
    RT=$(tput sgr0)
else
    R="" G="" Y="" B="" P="" W="" GR="" BD="" DM="" RT=""
fi

# peek at mode early for banner (full parse happens later)
_DEPLOY=false
for _a in "$@"; do [[ "$_a" == "-eploy" ]] && _DEPLOY=true; done

echo "${B}"
echo "      ⢀⣀ ⢀⡀ ⡀⣀ ⢀⡀ ⣇⡀ ⡀⣀ ⢀⡀ ⡀⢀ ⢀⡀ ⡀⣀ ⢀⡀"
echo "      ⠣⠤ ⠣⠭ ⠏  ⠣⠭ ⠧⠜ ⠏  ⠣⠜ ⠱⠃ ⠣⠜ ⠏  ⠣⠭"
if [[ "$_DEPLOY" == true ]]; then
    echo "      ⠤⠤⠤⠣${G} prd ${B}⠱⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤"
else
    echo "      ⠤⠤⠤⠣${P} dev ${B}⠱⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤"
fi
echo "${RT}"

# help cmd
for arg in "$@"; do
    if [[ "$arg" == "-h" ]] || [[ "$arg" == "--help" ]]; then
        echo "usage: ./d [-eploy] [-v] [reset-db] [-h|--help]"
        echo ""
        echo " (no arg)   dev mode, use for local building"
        echo " -eploy     deploy mode (build + systemd, requires sudo)"
        echo " -v         verbose, show output from docker/npm/vite"
        echo " -reset-db  nuke and recreate the local database"
        echo " -h/--help  this text"
        exit 0
    fi
done

log_ok()    { echo "${G}[${BD}+${RT}${G}] $*${RT}"; }
log_warn()  { echo "${Y}[${BD}*${RT}${Y}] $*${RT}"; }
log_fail()  { echo "${R}[${BD}!${RT}${R}] $*${RT}"; exit 1; }
log_info()  { echo "[~] $*"; }

log_frfr()  {
    echo ""
    echo "${R}   !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
    echo "${BD}        PLEASE PAY ATTENTION${RT}"
    echo "${R}   !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
    echo ""
}

# step
STEPCOUNT=0
log_step()  { 
    STEPCOUNT=$((STEPCOUNT +1))
    echo ""; 
    echo "${B}${BD} > ${RT}${DM}STEP $STEPCOUNT: $*${RT}"; }

# input
input_yn() {
    local answer
    printf "${W}[${P}${BD}?${RT}${W}]${RT} $* ${R}Y${W}/${G}N${RT}: "
    read -r answer
    case "$answer" in
        [yY]) return 0 ;;
        *)    return 1 ;;
    esac
}

log_test() {
    log_step "starting log formatting test"
    log_info "some more info for you!"
    log_info "what if it was LONGER info it was info so long it wrapped in your term or something idk"
    log_warn "MAYBE IT WAS TOO LONG WARNING"
    log_step "step test"
    if ! input_yn "Do you actually wanna press these BUTTONS?"; then
        log_warn "ENTERED N"
    else
        log_info "ENTERED Y"
    fi

    log_frfr
    log_fail "OH NOOOOOO"
}
# log_test

# flags
DEPLOY=false
VERBOSE=false
RESETDB=false
for arg in "$@"; do
    case "$arg" in
        -eploy)    DEPLOY=true ;;
        -v)        VERBOSE=true ;;
        -reset-db) RESETDB=true ;;
    esac
done

# wrap command stdout/stderr 
# for docker, npm, vite output suppression
quiet() {
    if [[ "$VERBOSE" == true ]]; then
        "$@"
    else
        "$@" &>/dev/null
    fi
}

INSTALL_DIR="/opt/cerebrovore"

if [[ "$DEPLOY" == true ]]; then
    log_info "mode: deploy"

    # must be root for systemd, service account, /opt writes
    if [[ "$EUID" -ne 0 ]]; then
        log_fail "deploy requires sudo: sudo ./d -eploy"
    fi
else
    log_info "mode: dev"
fi

# project root (e.g. run from ./scripts/d or something
# it's handled
ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"
log_info "project root: $ROOT"

# set up env
log_step "check .env"

# we need to generate a .env if it doesn't exist or if it's
# still identical to the example (nobody edited it).
NEEDS_SETUP=false
if [[ ! -f .env ]]; then
    NEEDS_SETUP=true
    log_warn ".env does not exist"
elif [[ -f .env.example ]] && diff -q .env .env.example &>/dev/null; then
    NEEDS_SETUP=true
    log_warn ".env is identical to .env.example (default secrets)"
fi

if [[ "$NEEDS_SETUP" == true ]]; then
    log_info "let's set up .env"
    echo ""

    # these just need to be random strings 
    # cerebrovore doesn't actually care if its default
    # .. but *I* do
    printf "${W}[${P}${BD}?${RT}${W}]${RT} secrets (SESSION_KEY, LRCD_SECRET, POSTGRES_PASSWORD)\n"
    printf "   ${G}(G)${RT}enerate random / ${R}(M)${RT}anual entry: "
    read -r SECRET_MODE

    # generate url safe passwords (standard hex is fine)
    case "$SECRET_MODE" in
        [gG])
            SESSION_KEY=$(openssl rand -hex 32)
            LRCD_SECRET=$(openssl rand -hex 32)
            POSTGRES_PASSWORD=$(openssl rand -hex 32)
            log_ok "generated random secrets"
            ;;
        [mM])
            printf "   SESSION_KEY: " && read -r SESSION_KEY
            printf "   LRCD_SECRET: " && read -r LRCD_SECRET
            printf "   POSTGRES_PASSWORD: " && read -r POSTGRES_PASSWORD
            if [[ "$POSTGRES_PASSWORD" =~ [^a-zA-Z0-9_.-] ]]; then
                log_warn "password contains URL-unsafe characters (/, +, @, etc.)"
                log_fail "use only alphanumeric, dash, dot, underscore"
            fi
            log_ok "manual secrets set"
            ;;
        *)
            log_fail "invalid choice, run ./d again"
            ;;
    esac

    echo ""
    log_info "this is completely optional in dev to be clear"
    printf "${W}[${P}${BD}?${RT}${W}]${RT} YOUTUBE_API_KEY (ww / yt metadata grabbing)\n"
    printf "   ${G}(M)${RT}anual entry / ${R}(S)${RT}kip: "
    read -r YT_MODE

    case "$YT_MODE" in
        [mM])
            printf "   YOUTUBE_API_KEY: " && read -r YOUTUBE_API_KEY
            log_ok "youtube api key set"
            ;;
        [sS]|"")
            YOUTUBE_API_KEY="NONE"
            log_ok "youtube api key skipped (set to NONE)"
            ;;
        *)
            log_fail "invalid choice, run ./d again"
            ;;
    esac

    # admin username
    echo ""
    printf "${W}[${P}${BD}?${RT}${W}]${RT} ADMIN_USERNAME (enter for 'admin'): "
    read -r ADMIN_USERNAME
    ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
    log_ok "admin username: $ADMIN_USERNAME"

    # write .env
    cat > .env <<EOF
SESSION_KEY=${SESSION_KEY}
LRCD_SECRET=${LRCD_SECRET}
POSTGRES_USER=cbv
POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
POSTGRES_DB=cbvdb
POSTGRES_PORT=9001
YOUTUBE_API_KEY=${YOUTUBE_API_KEY}
ADMIN_USERNAME=${ADMIN_USERNAME}
EOF
    echo ""
    log_ok ".env written"

    echo ""
    log_info "generated .env:"
    echo "   SESSION_KEY=${DM}${R}${SESSION_KEY}${RT}"
    echo "   LRCD_SECRET=${DM}${R}${LRCD_SECRET}${RT}"
    echo "   POSTGRES_USER=${DM}${R}cbv${RT}"
    echo "   POSTGRES_PASSWORD=${DM}${R}${POSTGRES_PASSWORD}${RT}"
    echo "   POSTGRES_DB=${DM}${R}cbvdb${RT}"
    echo "   POSTGRES_PORT=${DM}${R}9001${RT}"
    echo "   YOUTUBE_API_KEY=${DM}${R}${YOUTUBE_API_KEY}${RT}"
    echo "   ADMIN_USERNAME=${DM}${R}${ADMIN_USERNAME}${RT}"
    echo ""

    log_info "${R}${BD}MAKE SURE TO SET UP ADMIN ACCOUNT IN THE WEBAPP${RT}"
    log_info "${R}${BD}SIGN UP WITH THIS ACCOUNT NAME:${G} ${ADMIN_USERNAME}${RT}"
    if ! input_yn "Will you make the admin account please"; then
        log_warn "I'M NOT GONNA SELL YOU THESE BASS STRINGS"
        log_warn "I, I- I'M /NEVER/ GONNA SELL YOU THESE BASS STRINGS"
        log_warn "b-BUSTER!"
    else
        log_info "thank you"
    fi
else
    log_ok ".env exists and is configured"
fi

# load .env vars into this shell for postgres build
set -a
source .env
set +a

# docker (postgres)
log_step "checking docker"

if ! command -v docker &>/dev/null; then
    log_fail "docker not found, install docker and try again"
fi

if ! docker info &>/dev/null; then
    log_fail "docker daemon isn't running, start docker desktop (or dockerd) and try again"
fi

log_ok "docker available"

# are we nuking the db
if [[ "$RESETDB" == true ]]; then
    log_frfr
    log_warn "this will DESTROY the local database and all its data"
    if input_yn "are you SURE you want to reset the database?"; then
        log_step "resetting database"
        quiet docker compose down -v
        log_ok "database volume destroyed"
    else
        log_info "skipping database reset"
    fi
fi

log_step "starting postgres"
quiet docker compose up -d --wait db

# -reset-db bug
# generating .env -> nuke db -> sometimes pg not ready
log_info "waiting for postgres to accept connections..."
for i in $(seq 1 15); do
    if docker compose exec -T db pg_isready -U "$POSTGRES_USER" &>/dev/null; then
        break
    fi
    sleep 1
done
if ! docker compose exec -T db pg_isready -U "$POSTGRES_USER" &>/dev/null; then
    log_fail "postgres didn't become ready in 15s, check docker logs"
fi
log_ok "postgres is ready"

# migrations
log_step "database migrations"

if ! command -v migrate &>/dev/null; then
    log_fail "migrate not found, install golang-migrate"
fi

DB_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable"

# check current state before touching anything
MIGRATE_VERSION=$(migrate -path migrations/ -database "$DB_URL" version 2>&1) || true

if echo "$MIGRATE_VERSION" | grep -q "dirty"; then
    log_frfr
    log_fail "database is in dirty migration state, resolve manually with migrate -path migrations/ -database \"\$DB_URL\" force <version>"
fi

# run pending migrations (if any)
# migrate exits non-zero for "no change" (nothing to do),
MIGRATE_OUTPUT=$(migrate -path migrations/ -database "$DB_URL" up 2>&1) && MIGRATE_RC=0 || MIGRATE_RC=$?

if echo "$MIGRATE_OUTPUT" | grep -q "no change"; then
    log_ok "migrations up to date"
elif [[ $MIGRATE_RC -eq 0 ]]; then
    log_ok "migrations applied"
else
    log_fail "migration failed: $MIGRATE_OUTPUT"
fi

# frontend
log_step "checking node"

if ! command -v node &>/dev/null; then
    log_fail "node not found, install node (>= 18) and try again"
fi
log_ok "node $(node --version)"

# please don't change this
FRONTEND="$ROOT/frontend/cbv"

# install deps if node_modules is missing or package.json
# is newer (new dependency added / pushed).
if [[ ! -d "$FRONTEND/node_modules" ]] || [[ "$FRONTEND/package.json" -nt "$FRONTEND/node_modules" ]]; then
    log_step "installing npm dependencies"
    quiet bash -c "cd '$FRONTEND' && npm install"
    log_ok "npm install done"
else
    log_ok "frontend dependencies up to date"
fi

# track background pids for cleanup
PIDS=()

if [[ "$DEPLOY" == false ]]; then
    # dev: vite w/ hmr
    log_step "starting vite with HMR"
    if [[ "$VERBOSE" == true ]]; then
        (cd "$FRONTEND" && npx vite dev) &
    else
        (cd "$FRONTEND" && npx vite dev) &>/dev/null &
    fi
    PIDS+=($!)
    log_info "vite dev pid: $!"
    log_ok "vite started"
else
    # deploy: one-shot vite build
    log_step "building frontend (production)"
    quiet bash -c "cd '$FRONTEND' && npx vite build"
    log_ok "frontend built"
fi

# go setup
log_step "checking go"

if ! command -v go &>/dev/null; then
    log_fail "go not found, install go (>= 1.25) and try again"
fi
log_ok "go $(go version | awk '{print $3}')"

if [[ "$DEPLOY" == false ]]; then
    GO_FLAGS="-db -midp -dev"
    log_step "starting go backend (dev: $GO_FLAGS)"

    CLEANED=false
    cleanup() {
        [[ "$CLEANED" == true ]] && return
        CLEANED=true
        set +e
        log_step "stopping..."
        for pid in "${PIDS[@]}"; do
            pkill -P "$pid" 2>/dev/null
            kill "$pid" 2>/dev/null
            log_info "stopped pid $pid"
        done
        log_ok "so long..."
    }
    trap cleanup EXIT INT TERM

    log_ok "http://localhost:8080/"
    go run ./cmd/main.go $GO_FLAGS || true
else
    # deploy, build binary + systemd
    # check git status
    log_step "checking git"
    git fetch --quiet origin
    LOCAL=$(git rev-parse HEAD)
    REMOTE=$(git rev-parse origin/main)
    if [[ "$LOCAL" != "$REMOTE" ]]; then
        log_warn "local is behind origin/main"
        if input_yn "abort and update first?"; then
            log_info "run: git pull origin main"
            log_info "then re-run: ./d -eploy"
            exit 0
        else
            log_warn "deploying from older version"
        fi
    else
        log_ok "up to date with origin/main"
    fi

    # service account
    log_step "service account"
    SVC_USER="cerebrovore"
    if id "$SVC_USER" &>/dev/null; then
        log_ok "user '$SVC_USER' exists"
    else
        log_info "creating system user '$SVC_USER'"
        useradd -r -s /usr/sbin/nologin "$SVC_USER"
        log_ok "user '$SVC_USER' created"
    fi

    # build binary
    log_step "building go binary"
    quiet go build -o cerebrovore ./cmd/main.go
    log_ok "binary built"

    # stop service if running
    log_step "deploying to $INSTALL_DIR"
    if systemctl is-active cerebrovore &>/dev/null; then
        log_info "stopping cerebrovore service"
        systemctl stop cerebrovore
        log_ok "service stopped"
    fi

    # set up install dir
    mkdir -p "$INSTALL_DIR"

    # wipe old build artifacts (NOT uploads, .env, .fileStore)
    rm -f "$INSTALL_DIR/cerebrovore"
    rm -rf "$INSTALL_DIR/frontend/dist"
    rm -rf "$INSTALL_DIR/static"
    rm -rf "$INSTALL_DIR/tmpl"
    rm -rf "$INSTALL_DIR/migrations"
    rm -f "$INSTALL_DIR/docker-compose.yml"

    # copy fresh build
    cp cerebrovore "$INSTALL_DIR/"
    mkdir -p "$INSTALL_DIR/frontend"
    cp -r frontend/dist "$INSTALL_DIR/frontend/"
    cp -r static "$INSTALL_DIR/"
    cp -r tmpl "$INSTALL_DIR/"
    cp -r migrations "$INSTALL_DIR/"
    cp docker-compose.yml "$INSTALL_DIR/"

    # make sure .env exists in install dir
    if [[ ! -f "$INSTALL_DIR/.env" ]]; then
        log_warn "no .env in $INSTALL_DIR"
        log_info "copying from repo, you should review it"
        cp .env "$INSTALL_DIR/.env"
    fi

    # make sure uploads dir exists
    mkdir -p "$INSTALL_DIR/uploads"

    # set ownership
    chown -R "$SVC_USER":"$SVC_USER" "$INSTALL_DIR"
    log_ok "files deployed to $INSTALL_DIR"

    # prod database
    log_step "prod db"

    # source prod .env for credentials
    set -a
    source "$INSTALL_DIR/.env"
    set +a

    # check if postgres is already running
    if docker compose -f "$INSTALL_DIR/docker-compose.yml" ps db --status running 2>/dev/null | grep -q db; then
        log_ok "postgres already running"
    else
        log_info "starting postgres"
        (cd "$INSTALL_DIR" && docker compose up -d --wait db)
        log_ok "postgres started"
    fi

    # back up database before migrations
    log_step "database backup"
    BACKUP_DIR="$INSTALL_DIR/backups"
    mkdir -p "$BACKUP_DIR"
    BACKUP_FILE="$BACKUP_DIR/cbvdb-$(date +%Y%m%d-%H%M%S).sql"
    PGPASSWORD="$POSTGRES_PASSWORD" pg_dump -h 127.0.0.1 -p "$POSTGRES_PORT" -U "$POSTGRES_USER" "$POSTGRES_DB" > "$BACKUP_FILE" 2>/dev/null && \
        log_ok "backup: $BACKUP_FILE" || \
        log_warn "backup failed (first deploy? continuing...)"

    # migrations
    log_step "production migrations"
    PROD_DB_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable"

    MIGRATE_VERSION=$(migrate -path "$INSTALL_DIR/migrations/" -database "$PROD_DB_URL" version 2>&1) || true
    if echo "$MIGRATE_VERSION" | grep -q "dirty"; then
        log_frfr
        log_fail "production database is in dirty migration state, resolve manually"
    fi

    MIGRATE_OUTPUT=$(migrate -path "$INSTALL_DIR/migrations/" -database "$PROD_DB_URL" up 2>&1) && MIGRATE_RC=0 || MIGRATE_RC=$?
    if echo "$MIGRATE_OUTPUT" | grep -q "no change"; then
        log_ok "migrations up to date"
    elif [[ $MIGRATE_RC -eq 0 ]]; then
        log_ok "migrations applied"
    else
        log_fail "migration failed: $MIGRATE_OUTPUT"
    fi

    # systemd service
    log_step "systemd service"
    cat > /etc/systemd/system/cerebrovore.service <<EOF
[Unit]
Description=cerebrovore
After=network.target docker.service
Requires=docker.service

[Service]
Type=simple
User=${SVC_USER}
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/cerebrovore -cold -db -port 9000
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable cerebrovore
    systemctl start cerebrovore
    log_ok "cerebrovore service started"

    # verify
    sleep 1
    if systemctl is-active cerebrovore &>/dev/null; then
        log_ok "service is running"
    else
        log_fail "service failed to start, check: journalctl -u cerebrovore"
    fi

    log_step "deploy complete"
    log_info "logs: journalctl -u cerebrovore -f"
    log_info "stop: sudo systemctl stop cerebrovore"
    log_info "status: sudo systemctl status cerebrovore"
fi
