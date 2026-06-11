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
