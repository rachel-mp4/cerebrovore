#!/usr/bin/env bash
# shared helpers for the cerebrovore dev/prod scripts
# Sourced by ./d and every crumb script

[[ -n "${_CBV_LIB:-}" ]] && return
_CBV_LIB=1

# colors (stdio check)
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

VERBOSE="${VERBOSE:-false}"

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

STEPCOUNT="${STEPCOUNT:-0}"
log_step()  {
    STEPCOUNT=$((STEPCOUNT + 1))
    echo "";
    echo "${B}${BD} > ${RT}${DM}STEP $STEPCOUNT: $*${RT}"; }

input_yn() {
    local answer
    printf "${W}[${P}${BD}?${RT}${W}]${RT} $* ${R}Y${W}/${G}N${RT}: "
    read -r answer
    case "$answer" in
        [yY]) return 0 ;;
        *)    return 1 ;;
    esac
}

quiet() {
    if [[ "$VERBOSE" == true ]]; then
        "$@"
    else
        "$@" &>/dev/null
    fi
}

banner() {
    local mode="$1"
    echo "${B}"
    echo "      ⢀⣀ ⢀⡀ ⡀⣀ ⢀⡀ ⣇⡀ ⡀⣀ ⢀⡀ ⡀⢀ ⢀⡀ ⡀⣀ ⢀⡀"
    echo "      ⠣⠤ ⠣⠭ ⠏  ⠣⠭ ⠧⠜ ⠏  ⠣⠜ ⠱⠃ ⠣⠜ ⠏  ⠣⠭"
    if [[ "$mode" == prd ]]; then
        echo "      ⠤⠤⠤⠣${G} prd ${B}⠱⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤"
    else
        echo "      ⠤⠤⠤⠣${P} dev ${B}⠱⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤⠤"
    fi
    echo "${RT}"
}

load_env() {
    set -a
    source .env
    set +a
}

# resolve migrate bin: MIGRATE_BIN from .env, else PATH
# stale path falls back to PATH
require_migrate() {
    MIGRATE_BIN="${MIGRATE_BIN:-migrate}"
    if "$MIGRATE_BIN" -version &>/dev/null; then
        return 0
    fi
    if command -v migrate &>/dev/null; then
        MIGRATE_BIN="migrate"
        return 0
    fi
    log_fail "golang-migrate not found (MIGRATE_BIN='$MIGRATE_BIN'). run ./scripts/setup or set MIGRATE_BIN in .env"
}

# pg_dump database to ./backups (or $BACKUP_DIR)
# set SKIP_BACKUP=1 to tolerate failure (e.g. first deploy, nothing to dump yet)
# assumes you have postgres credentials sourced
do_backup() {
    log_step "database backup"
    local BACKUP_DIR="${BACKUP_DIR:-backups}"
    mkdir -p "$BACKUP_DIR"
    local STAMP=$(date +%Y%m%d-%H%M%S)
    local BACKUP_FILE="$BACKUP_DIR/${POSTGRES_DB}-${STAMP}.sql"

    # don't swallow stderr, so a failed dump tells you why
    if PGPASSWORD="$POSTGRES_PASSWORD" pg_dump -h 127.0.0.1 -p "$POSTGRES_PORT" -U "$POSTGRES_USER" "$POSTGRES_DB" > "$BACKUP_FILE"; then
        log_ok "backup: $BACKUP_FILE"
    else
        if [[ "${SKIP_BACKUP:-}" == 1 ]]; then
            log_warn "backup failed"
            if input_yn "SKIP_BACKUP=1 (is this your first deploy?) continue?"; then
                log_ok "backup skipped"
            else
                log_fail "aborting"
            fi
        else
            log_fail "aborting before migrations. if this is the first deploy, re-run with SKIP_BACKUP=1"
        fi
    fi
}

# run the dev stack: vite (HMR) + the go backend
# press R to restart the backend, Ctrl-C to quit
# expects $ROOT to point at the repo root
do_dev() {
    # frontend
    log_step "checking node"

    if ! command -v node &>/dev/null; then
        log_fail "node not found, install node (>= 18) and try again"
    fi
    log_ok "node $(node --version)"

    # please don't change this
    local FRONTEND="$ROOT/frontend/cbv"

    # install deps if node_modules is missing or package.json is newer
    if [[ ! -d "$FRONTEND/node_modules" ]] || [[ "$FRONTEND/package.json" -nt "$FRONTEND/node_modules" ]]; then
        log_step "installing npm dependencies"
        quiet bash -c "cd '$FRONTEND' && npm install"
        log_ok "npm install done"
    else
        log_ok "frontend dependencies up to date"
    fi

    # track background pids for cleanup
    PIDS=()

    # dev: vite w/ hmr
    log_step "starting vite with HMR"
    # i feel like we shouldn't see vite output even in verbose 
    # because it's too aggressive but i'm not positive, 
    # that's why this is commented out.
    # stdin from /dev/null so vite doesn't fight the restart loop for the terminal
    if [[ "$VERBOSE" == true ]]; then
        # logLevel error because i'm bad at coding and there are lots of warnings
        # should learn about how to write npm scripts, because i think it's better
        # to do npm run dev vs npx vite dev, in case npm build system ever is more
        # than just running vite, but right now 6.17.2026 it should be identical
        (cd "$FRONTEND" && npx vite dev --logLevel error --clearScreen false) </dev/null &
    else
        (cd "$FRONTEND" && npm run dev) </dev/null &>/dev/null &
    fi
    PIDS+=($!)
    log_info "vite dev pid: $!"
    log_ok "vite started"

    # go setup
    log_step "checking go"

    if ! command -v go &>/dev/null; then
        log_fail "go not found, install go (>= 1.25) and try again"
    fi
    log_ok "go $(go version | awk '{print $3}')"

    local GO_FLAGS="$@"

    GO_PID=""
    GO_IN=""
    STTY_SAVED=""

    log_info "press ${BD}R${RT} to restart the backend; Ctrl-C to quit"
    start_go() {
        log_step "starting go backend (dev: $GO_FLAGS)"
        if [[ -n "$GO_IN" ]]; then
            go run ./cmd/main.go $GO_FLAGS <"$GO_IN" &
        else
            go run ./cmd/main.go $GO_FLAGS &
        fi
        GO_PID=$!
        log_info "go backend pid: $GO_PID"
        log_ok "http://localhost:8080/"
    }
    stop_go() {
        [[ -n "$GO_PID" ]] || return
        pkill -KILL -P "$GO_PID" 2>/dev/null || true
        kill -KILL "$GO_PID" 2>/dev/null || true
        wait "$GO_PID" 2>/dev/null || true
        GO_PID=""
    }

    CLEANED=false
    cleanup() {
        [[ "$CLEANED" == true ]] && return
        CLEANED=true
        set +e
        printf '\n'
        log_step "stopping..."
        stop_go
        for pid in "${PIDS[@]}"; do
            pkill -P "$pid" 2>/dev/null
            kill "$pid" 2>/dev/null
            log_info "stopped pid $pid"
        done
        exec 3>&- 2>/dev/null
        [[ -n "$GO_IN" ]] && rm -f "$GO_IN"
        [[ -n "$STTY_SAVED" ]] && stty "$STTY_SAVED" 2>/dev/null
        log_ok "so long..."
    }
    trap cleanup EXIT
    trap 'cleanup; exit 130' INT
    trap 'cleanup; exit 143' TERM

    if [[ -t 0 ]]; then
        # okay
        GO_IN="$(mktemp -u)"
        mkfifo "$GO_IN"

        STTY_SAVED="$(stty -g)"
        stty -icanon -echo min 1 time 0

        start_go
        exec 3>"$GO_IN"

        while true; do
            IFS= read -rsn1 ch || break
            case "$ch" in
                r|R)
                    printf '\n'
                    log_step "restarting backend"
                    stop_go
                    sleep 0.3
                    log_info "press ${BD}R${RT} to restart the backend; Ctrl-C to quit"
                    start_go
                    ;;
                '')
                    printf '\n' >&3 || true
                    printf '\n'
                    ;;
                *)
                    printf '%s' "$ch" >&3 || true
                    printf '%s' "$ch"
                    ;;
            esac
        done
    else
        start_go
        [[ -n "$GO_PID" ]] && wait "$GO_PID" 2>/dev/null || true
    fi
}

# make sure docker is up and pg accepts conn. 
# expects POSTGRES_* already in env (load_env)
do_ensure_db() {
    # docker (postgres)
    log_step "checking docker"

    if ! command -v docker &>/dev/null; then
        log_fail "docker not found, install docker and try again"
    fi

    if ! docker info &>/dev/null; then
        log_fail "docker daemon isn't running, start docker desktop (or dockerd) and try again"
    fi

    log_ok "docker available"

    log_step "starting postgres"
    quiet docker compose up -d --wait db

    # changelog: added bug
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
}

# migrate the db TO a specific version
# warn before rolling backwards

# ./scripts/mto 12 -> goto 12 (warns if it rolls back from a higher version)
do_mto() {
    local target="${1:-}"
    if [[ -z "$target" ]]; then
        log_fail "usage: mto <version>"
    fi
    if ! [[ "$target" =~ ^[0-9]+$ ]]; then
        log_fail "version must be a number // '$target'"
    fi
    require_migrate

    local DB_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable"

    # warn if this moves us backwards
    local current=$("$MIGRATE_BIN" -path migrations/ -database "$DB_URL" version 2>&1) || true
    if [[ "$current" =~ ^[0-9]+$ ]] && (( target < current )); then
        log_frfr
        log_warn "this rolls the database DOWN from $current to $target"
        input_yn "are you sure?" || { log_info "aborting"; return 0; }
    fi

    log_step "migrating to version $target"
    "$MIGRATE_BIN" -path migrations/ -database "$DB_URL" goto "$target"
    log_ok "now at version $target"
}

# migrate up, apply pending migrations to the local db
do_mup() {
    log_step "database migrations"

    require_migrate

    local DB_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable"

    # check current state before touching anything
    local MIGRATE_VERSION=$("$MIGRATE_BIN" -path migrations/ -database "$DB_URL" version 2>&1) || true

    if echo "$MIGRATE_VERSION" | grep -q "dirty"; then
        log_frfr
        log_fail "database is in dirty migration state, resolve manually with migrate -path migrations/ -database \"\$DB_URL\" force <version>"
    fi

    # run pending migrations (migrate exits non-zero on 'no change')
    local MIGRATE_OUTPUT
    if MIGRATE_OUTPUT=$("$MIGRATE_BIN" -path migrations/ -database "$DB_URL" up 2>&1); then
        log_ok "migrations applied"
    elif echo "$MIGRATE_OUTPUT" | grep -q "no change"; then
        log_ok "migrations up to date"
    else
        log_fail "migration failed: $MIGRATE_OUTPUT"
    fi
}

# nuke the local database volume and its data
# THIS ONE MEANS "BUSINESS"
do_reset_db() {
    log_frfr
    log_warn "this will DESTROY the local database and all its data"
    if input_yn "are you SURE you want to reset the database?"; then
        log_step "resetting database"
        quiet docker compose down -v
        log_ok "database volume destroyed"
    else
        log_info "skipping database reset"
    fi
}

# first-run setup for a new dev/environment
# generate .env if it's missing or the same as the example
# does nothing once .env exists, be careful with moving .env around
do_setup() {
    log_step "check .env"

    # need to generate a .env if it doesn't exist or if it's
    # still identical to the example (nobody edited it).
    local NEEDS_SETUP=false
    if [[ ! -f .env ]]; then
        NEEDS_SETUP=true
        log_warn ".env does not exist"
    elif [[ -f .env.example ]] && diff -q .env .env.example &>/dev/null; then
        NEEDS_SETUP=true
        log_warn ".env is identical to .env.example (default secrets)"
    fi

    if [[ "$NEEDS_SETUP" == false ]]; then
        log_ok ".env exists and is configured"
        return 0
    fi

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

    # discord link
    echo ""
    printf "${W}[${P}${BD}?${RT}${W}]${RT} DISCORD_LINK (optional): "
    read -r DISCORD_LINK
    log_ok "discord link: $DISCORD_LINK"

    # report delimiter
    echo ""
    printf "${W}[${P}${BD}?${RT}${W}]${RT} REPORT_DELIMITER (maybe pick a weird unicode character that users probably won't use): "
    read -r REPORT_DELIMITER
    log_ok "report delimiter: $REPORT_DELIMITER"
    
    manual_select_migrate() {
        # migrate tooling (golang-migrate), stored as MIGRATE_BIN in .env
        # no need for global PATH stuff
        echo ""
        printf "${W}[${P}${BD}?${RT}${W}]${RT} golang-migrate (db migration tool)\n"
        printf "   ${G}(1)${RT} install via go install ${B}(recommended)\n"
        printf "   ${G}(2)${RT} download from github + compile via go build into ./bin  ${DM}(1 fewer place to delete stuff from when you're done with this project)${RT}\n"
        printf "   ${G}(3)${RT} manual install to PATH \n"
        printf "   ${G}(4)${RT} manual entry of binary location \n"
        printf "   choose [1/2/3/4]: "
        read -r MIGRATE_MODE

        case "$MIGRATE_MODE" in
            1)
                command -v go &>/dev/null || log_fail "go not found, can't go-install migrate"
                log_info "installing migrate via go (may take a little while)..."
                go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1
                MIGRATE_BIN="$(go env GOPATH)/bin/migrate"
                [[ -x "$MIGRATE_BIN" ]] && log_ok "installed: $MIGRATE_BIN" \
                    || log_fail "go install finished but $MIGRATE_BIN missing, check go env GOPATH"
                ;;
            2)
                log_warn "may take a while, proceeding"
                command -v go  &>/dev/null || log_fail "go not found, can't compile migrate"
                command -v git &>/dev/null || log_fail "git not found, can't clone migrate"
                log_info "cloning + building golang-migrate into ./bin ..."
                mkdir -p bin
                DEST="$(pwd)/bin/migrate"
                TMP=$(mktemp -d)
                git clone --depth 1 https://github.com/golang-migrate/migrate "$TMP"
                (cd "$TMP" && git checkout v4.19.1 && go build -tags 'postgres' -o "$DEST" ./cmd/migrate)
                rm -rf "$TMP"
                MIGRATE_BIN="bin/migrate"
                [[ -x "$MIGRATE_BIN" ]] && log_ok "built: ./bin/migrate" \
                    || log_fail "build failed, ./bin/migrate not found"
                ;;
            3)
                log_info "come back when you have it installed!"
                exit 0
                ;;
            4)
                echo "   MIGRATE_BIN=" 
                read -r MIGRATE_BIN
                ;;
            *)
                log_fail "invalid choice, run ./d again"
                ;;
        esac
    }

    if command -v migrate &>/dev/null; then
        if input_yn "you have migrate ($(command -v migrate)) in your PATH, is it ok to use?"; then
            MIGRATE_BIN="migrate"
        else
            manual_select_migrate
        fi
    else
        manual_select_migrate
    fi

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
DISCORD_LINK=${DISCORD_LINK}
REPORT_DELIMITER=${REPORT_DELIMITER}
MIGRATE_BIN=${MIGRATE_BIN}
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
    echo "   DISCORD_LINK=${DM}${R}${DISCORD_LINK}${RT}"
    echo "   REPORT_DELIMITER=${DM}${R}${REPORT_DELIMITER}${RT}"
    echo "   MIGRATE_BIN=${DM}${R}${MIGRATE_BIN}${RT}"
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
}
