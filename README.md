```
 ⢀⣀ ⢀⡀ ⡀⣀ ⢀⡀ ⣇⡀ ⡀⣀ ⢀⡀ ⡀⢀ ⢀⡀ ⡀⣀ ⢀⡀
 ⠣⠤ ⠣⠭ ⠏  ⠣⠭ ⠧⠜ ⠏  ⠣⠜ ⠱⠃ ⠣⠜ ⠏  ⠣⠭
```

auto setup
1. `./d`
2. follow the instructions
3. hopefully everything works! if not, let me know!

manual setup
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

manual dev build
1. `cd frontend/cbv`
2. `npm run dev`
3. open another terminal
4. `go run ./cmd -db -midp`

at some point or maybe never i'll make it so you can run a lousy version
without db for dev purposes to make it easier if you just wanna do frontend

manual prod build
1. `cd frontend/cbv`
2. `npm run build`
3. `cd -`
4. `go run ./cmd -cold -db -midp`

the flags are explained a bit if you run `go run ./cmd -h`

the one strange thing is that we have an identity provider, which at the moment
is either an in-memory store of username:hashedPassword that gets backed up to
a file (`-midp`), or an external service that we communicate with through http
api (`-sidp {port}`) (not included in this repo, it's private). cerebrovore
sends the credentials to that service and it responds accordingly

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
potential for a collision! for this reason, prefer html IDs with a dash

hashtag parser in linkifyjs doesn't like it when the hashtag is just numbers.
want to create a custom parser here, but be careful because performance here
matters, and it's nice to maintain flexibility. likely we don't care for any
commands, and in general as length of post id grows, probability that a post id
is just numbers approaches 0, so this is very low priority
