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

# crumb scripts
source "$SCRIPTS/lib.sh"
source "$SCRIPTS/setup"
source "$SCRIPTS/ensure-db"
source "$SCRIPTS/reset-db"
source "$SCRIPTS/mup"
source "$SCRIPTS/dev"

# flags
VERBOSE=false
RESETDB=false
INTERACTIVE=false
NOMIGRATE=false
for arg in "$@"; do
    case "$arg" in
        -v)          VERBOSE=true ;;
        -reset-db)   RESETDB=true ;;
        -i)          INTERACTIVE=true ;;
        -no-migrate) NOMIGRATE=true ;;
        -h|--help)
            echo "usage: ./d [-v] [-reset-db] [-i] [-no-migrate] [-h|--help]"
            echo ""
            echo " (no arg)     dev mode, use for local building"
            echo " -v           verbose, show output from docker/npm/vite"
            echo " -reset-db    nuke and recreate the local database"
            echo " -i           pick a crumb script to run (interactive)"
            echo " -no-migrate  dev mode but skip the migrate step"
            echo " -h/--help    this text"
            echo ""
            echo " prod deploy moved to: sudo ./scripts/prod"
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
    do_dev
}

# same thing, minus migrations (for when you just wanna poke around)
template_nomigrate() {
    do_setup
    load_env
    [[ "$RESETDB" == true ]] && do_reset_db
    do_ensure_db
    do_dev
}

# -i menu: pick a template or crumb script to run
pick_script() {
    log_step "pick something to run"
    echo "  ${Y}${BD}run${RT}"
    echo "   ${Y}(1)${RT}  standard    full dev (setup, db, migrate, run)"
    echo "   ${Y}(2)${RT}  no-migrate  full dev but skips migrations"
    echo "  ${B}${BD}db${RT}"
    echo "   ${B}(3)${RT}  ensure-db   start postgres"
    echo "   ${B}(4)${RT}  mup         migrate up"
    echo "   ${B}(5)${RT}  mto         migrate to a version"
    echo "   ${B}(6)${RT}  psql        psql shell"
    echo "   ${B}(7)${RT}  backup      dump the db"
    echo "   ${B}(8)${RT}  reset-db    nuke the db"
    echo "  ${G}${BD}dev${RT}"
    echo "   ${G}(9)${RT}  setup       generate .env"
    echo "   ${G}(10)${RT} dev         run vite + go"
    echo "  ${P}${BD}prod${RT}"
    echo "   ${P}(11)${RT} prod        deploy to the server"
    echo "  ${R}${BD}meta${RT}"
    echo "   ${R}(Q)${RT}  quit        SO LONG"

    printf "   choose: "
    read -r choice

    case "$choice" in
        1) template_standard; return ;;
        2) template_nomigrate; return ;;
        3) picked=ensure-db ;;
        4) picked=mup ;;
        5) picked=mto ;;
        6) picked=psql ;;
        7) picked=backup ;;
        8) picked=reset-db ;;
        9) picked=setup ;;
        10) picked=dev ;;
        11) picked=prod ;;
        q) picked=exit ;;
        Q) picked=exit ;;
        *) log_fail "invalid choice" ;;
    esac

    if [[ "$picked" == "exit" ]]; then
        exit 0
    fi

    printf "   args (optional): "
    read -r extra
    log_info "running ./scripts/$picked $extra"
    "$SCRIPTS/$picked" $extra
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
