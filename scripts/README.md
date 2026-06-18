# scripts!

## backup 

this backs up your database by dumping it to a file. hopefully you don't ever
need to use your backup, but if you do, you can load the backup from a clean
docker image of postgres using `\i path/to/dump.sql` inside `psql` utility

## dev

this is similar to just running the `d` script without any arguments, however
it does not have sensible defaults for cerebrovore's flags, and any flags you
pass to it get `dev` get forwarded to cerebrovore

## ensure-db

this ensures that you have docker installed and postgres running in an image,
it shouldn't be necessary to call directly in most instances

## mto

this migrates the database to a specific migration number (migrations are in
the top level migrations directory in this repository, and their numbers are
at the beginning of their name) passed as the first and only argument to this

the main reason you'd want to use this is if you're adding a new feature to
the database, and you need to revert some changes while protoyping. make sure
to backup before calling this!

## mup

this migrates the database all the way up to the most recent migration

## prod

this deploys cerebrovore to production, registering a new systemd service.

## psql

this runs psql, a tui for poking at postgres databases. can do a lot with it

## setup

this sets up your environment variables, if you haven't set them up already,
and in the process helps you with installing migrate, the tool we use to apply
database migrations
