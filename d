#!/usr/bin/env bash
# sets up a dev env and runs cereebrovore locally
#   ./d            dev mode: setup (if needed) + db + migrate + vite & go
#   ./d -v         verbose, show docker/npm/vite output
#   ./d -reset-db  nuke and recreate the local database
#   ./d -i         pick a crumb script to run (interactive)
#
# orchestration script, each can be run individually
# to accomplish this same task by hand, run the following scripts (./scripts)
#
# setup
# mup
# dev
#
# if deploying prod, prod deploy script in scripts/prod (sudo ./scripts/prod)

# fail early fail friendly
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
cd $ROOT
SCRIPTS="$ROOT/scripts"

# library
source "$SCRIPTS/backup"
source "$SCRIPTS/dev"
source "$SCRIPTS/lib.sh"
source "$SCRIPTS/mto"
source "$SCRIPTS/mup"
source "$SCRIPTS/reset-db"
source "$SCRIPTS/setup"

# flags
VERBOSE=false
IDENTITY_FLAG=" -midp"
FIRST_TIME_FLAG=""
RESETDB=false
INTERACTIVE=false
for arg in "$@"; do
    case "$arg" in
        -v)         VERBOSE=true ;;
        -s)         IDENTITY_FLAG="" ;;
        -first)     FIRST_TIME_FLAG=" -first" ;;
        -reset-db)  RESETDB=true; FIRST_TIME_FLAG=" -first" ;;
        -i)         INTERACTIVE=true ;;
        -h|--help)
            echo "usage: $0 [-v] [-s] [-reset-db] [-i] [-no-migrate] [-h|--help]"
            echo ""
            echo " (no arg)     dev mode, use for local building"
            echo " -v           verbose, show output from docker/npm/vite"
            echo " -s           use an external identity service provider listening on :9009"
            echo " -reset-db    nuke and recreate the local database"
            echo " -i           pick a crumb script to run (interactive)"
            echo " -h/--help    this text"
            echo ""
            exit 0 ;;
    esac
done

# the standard template: setup, db, migrate, run
template_standard() {
    do_setup
    load_env
    [[ "$RESETDB" == true ]] && do_reset_db
    do_mup
    do_dev "-db -dev$IDENTITY_FLAG$FIRST_TIME_FLAG"
}

# -i menu: pick a template or crumb script to run
pick_script() {
    log_step "pick something to run"
    echo "  ${Y}${BD}run${RT}"
    echo "   ${Y}(1)${RT}  standard     full dev (setup, db, migrate, run)"
    echo "  ${B}${BD}db${RT}"
    echo "   ${B}(2)${RT}  mup          migrate up"
    echo "   ${B}(3)${RT}  mto          migrate to a version"
    echo "   ${B}(4)${RT}  psql         psql shell"
    echo "   ${B}(5)${RT}  backup       dump the db"
    echo "   ${B}(6)${RT}  backup prod  dump the production db"
    echo "   ${B}(7)${RT}  reset-db     nuke the db"
    echo "   ${B}(8)${RT}  load backup  loads a backup"
    echo "  ${G}${BD}dev${RT}"
    echo "   ${G}(9)${RT}  setup        generate .env"
    echo "   ${G}(10)${RT} dev          run vite + go"
    echo "  ${P}${BD}prod${RT}"
    echo "   ${P}(11)${RT} prod         deploy to the server"
    echo "  ${R}${BD}meta${RT}"
    echo "   ${R}(Q)${RT}  quit         SO LONG"

    printf "   choose: "
    read -r choice

    case "$choice" in
        1) template_standard ;;
        2|3|5|6|7|10) load_env ;;&
        2) do_mup ;;
        3) printf "   version: "; read -r version; do_mto "$version" ;;
        4) "$SCRIPTS/psql" ;;
        5) do_backup ;;
        6) BACKUP_DIR="/opt/cerebrovore/backups" do_backup ;;
        7) do_reset_db ;;
        8) printf "   path or flags: "; read -r args; "$SCRIPTS/load" "$args" ;;
        9) do_setup ;;
        10) printf "   flags: "; read -r flags; do_dev "$flags" ;;
        11) printf "   args: "; read -r args; "$SCRIPTS/prod" "$args" ;;
        q|Q) return ;;
        *) log_fail "invalid choice" ;;
    esac
}

banner dev

# -i: pick a crumb script and run it
if [[ "$INTERACTIVE" == true ]]; then
    pick_script
    exit 0
fi

log_info "mode: dev"
log_info "project root: $ROOT"

template_standard
