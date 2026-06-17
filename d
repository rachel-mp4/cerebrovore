#!/usr/bin/env bash
# sets up a dev env and runs cereebrovore locally
#   ./d            dev mode: setup (if needed) + db + migrate + vite & go
#   ./d -v         verbose, show docker/npm/vite output
#   ./d -reset-db  nuke and recreate the local database
#   ./d -i         pick a crumb script to run (interactive)
#   ./d -no-migrate same as no-arg but skips migrations (e.g. just poking templates)
#
# orchestration script, each can be run individually
# to accomplish this same task by hand, run the following scripts (./scripts)
#
# setup
# ensure-db
# mup
# dev
#
# if deploying prod, prod deploy script in scripts/prod (sudo ./scripts/prod)

# fail early fail friendly
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
SCRIPTS="$ROOT/scripts"

# library
source "$SCRIPTS/lib.sh"

# flags
VERBOSE=false
IDENTITY_FLAG=" -midp"
RESETDB=false
INTERACTIVE=false
NOMIGRATE=false
for arg in "$@"; do
    case "$arg" in
        -v)          VERBOSE=true ;;
        -s)          IDENTITY_FLAG="" ;;
        -reset-db)   RESETDB=true ;;
        -i)          INTERACTIVE=true ;;
        -no-migrate) NOMIGRATE=true ;;
        -h|--help)
            echo "usage: $0 [-v] [-s] [-reset-db] [-i] [-no-migrate] [-h|--help]"
            echo ""
            echo " (no arg)     dev mode, use for local building"
            echo " -v           verbose, show output from docker/npm/vite"
            echo " -s           use an external identity service provider listening on :9009"
            echo " -reset-db    nuke and recreate the local database"
            echo " -i           pick a crumb script to run (interactive)"
            echo " -no-migrate  dev mode but skip the migrate step"
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
    do_ensure_db
    do_mup
    do_dev "-db -dev$IDENTITY_FLAG"
}

# same thing, minus migrations (for when you just wanna poke around)
template_nomigrate() {
    do_setup
    load_env
    [[ "$RESETDB" == true ]] && do_reset_db
    do_ensure_db
    do_dev "-db -dev$IDENTITY_FLAG"
}

# -i menu: pick a template or crumb script to run
pick_script() {
    log_step "pick something to run"
    echo "  ${Y}${BD}run${RT}"
    echo "   ${Y}(1)${RT}  standard     full dev (setup, db, migrate, run)"
    echo "   ${Y}(2)${RT}  no-migrate   full dev but skips migrations"
    echo "  ${B}${BD}db${RT}"
    echo "   ${B}(3)${RT}  ensure-db    start postgres"
    echo "   ${B}(4)${RT}  mup          migrate up"
    echo "   ${B}(5)${RT}  mto          migrate to a version"
    echo "   ${B}(6)${RT}  psql         psql shell"
    echo "   ${B}(7)${RT}  backup       dump the db"
    echo "   ${B}(8)${RT}  backup prod  dump the production db"
    echo "   ${B}(9)${RT}  reset-db     nuke the db"
    echo "  ${G}${BD}dev${RT}"
    echo "   ${G}(10)${RT} setup        generate .env"
    echo "   ${G}(11)${RT} dev          run vite + go"
    echo "  ${P}${BD}prod${RT}"
    echo "   ${P}(12)${RT} prod         deploy to the server"
    echo "  ${R}${BD}meta${RT}"
    echo "   ${R}(Q)${RT}  quit         SO LONG"

    printf "   choose: "
    read -r choice

    case "$choice" in
        1) template_standard ;;
        2) template_nomigrate ;;
        3|4|5|7|8|9|11) load_env ;;&
        3|4|5|11) do_ensure_db ;;&
        4) do_mup ;;
        5) printf "   version: "; read -r version; do_mto "$version" ;;
        6) "$SCRIPTS/psql" ;;
        7) do_backup ;;
        8) BACKUP_DIR="/opt/cerebrovore/backups" do_backup ;;
        9) do_reset_db ;;
        10) do_setup ;;
        11) printf "   flags: "; read -r flags; do_dev "$flags" ;;
        12) printf "   args: "; read -r args; "$SCRIPTS/prod" "$args" ;;
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
cd "$ROOT"
log_info "project root: $ROOT"

# run one of the two templates
if [[ "$NOMIGRATE" == true ]]; then
    template_nomigrate
else
    template_standard
fi
