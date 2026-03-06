```
 ⢀⣀ ⢀⡀ ⡀⣀ ⢀⡀ ⣇⡀ ⡀⣀ ⢀⡀ ⡀⢀ ⢀⡀ ⡀⣀ ⢀⡀
 ⠣⠤ ⠣⠭ ⠏  ⠣⠭ ⠧⠜ ⠏  ⠣⠜ ⠱⠃ ⠣⠜ ⠏  ⠣⠭
```

initial setup
1. install go
2. install npm
3. install docker (required for database, which is currently required since
   mock is no good)
4. install golang-migrate
5. copy over environment variables from example, changing stuff & generating
   secrets as desired
6. `docker compose up -d`
7. `./migrateup`
8. `cd frontend/cbv`
9. `npm install`

hopefully it's not too confusing...

dev build
1. `cd frontend/cbv`
2. `npm run dev`
3. open another terminal
4. `go run ./cmd -db`

at some point or maybe never i'll make it so you can run a lousy version
without db for dev purposes to make it easier if you just wanna do frontend

prod build
1. `cd frontend/cbv`
2. `npm run build`
3. `cd -`
4. `go run ./cmd -cold -db -idp`

the flags are explained a bit if you run `go run ./cmd -h`

but the basic idea is that in production ideally there is some identity
provider service (not yet implemented) that we direct unauthorized users to,
where they authorize, and then call a callback endpoint at which point we add
to database. if you want a more traditional imageboard experience without
accounts, or you're just doing development, it's totally fine to use db flag
without idp

### KNOWN BUGS AND BAD BEHAVIOR THATS ANNOYING 
when you click on a thumbnail, swapping it out for the full-size image, it's
bad need to store the size (aspect ratio dimensions size), broadcast over lrc
etc...

mock db is completely useless, need to upgrade it to an in memory db at least
lol

this is correct behavior that's fine, but note that since rendered posts in
threads get an id according to their encoding as base 36 alphanumeric number
every possible that's just numbers and letters unless it starts with a 0 is NOT
ok for use as an ID, because html ID should be unique, and this has the
potential for a collision!
