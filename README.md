```
 ⢀⣀ ⢀⡀ ⡀⣀ ⢀⡀ ⣇⡀ ⡀⣀ ⢀⡀ ⡀⢀ ⢀⡀ ⡀⣀ ⢀⡀
 ⠣⠤ ⠣⠭ ⠏  ⠣⠭ ⠧⠜ ⠏  ⠣⠜ ⠱⠃ ⠣⠜ ⠏  ⠣⠭
```

## auto setup

### mac/linux
1. run `./d`
2. follow the instructions
3. hopefully everything works! if not, let us know!

### windows
1. install git, it should come with git bash, which you should use as your
   terminal emulator (or you can pick your favorite terminal emulator that runs
bash) 
2. in git bash: `./d`
2. follow the instructions
3. maybe everything works! if not, please let us know, not testing on windows
   rn, but we want it to work!!

## manual setup (deprecated)

### required steps
1. install go
2. install npm
3. install docker (required for database, which is currently required since
   mock is no good)
4. install golang-migrate
5. copy over environment variables from example, changing stuff & generating
   secrets as desired
6. `docker compose up -d`
7. `./scripts/mup`
8. `cd frontend/cbv`
9. `npm install

and then you pick one of two paths to run the code:

### manual dev-style (w/ HMR) build
1. `cd frontend/cbv`
2. `npm run dev`
3. open another terminal
4. `go run ./cmd -db -midp -dev`

### manual prod-style (w/o HMR) build
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

## KNOWN BUGS AND BAD BEHAVIOR THATS ANNOYING 
mock db is completely useless, need to upgrade it to an in memory db at least
lol

this is correct behavior that's fine, but note that since rendered posts in
threads get an id according to their encoding as base 36 alphanumeric number
every possible that's just numbers and letters unless it starts with a 0 is NOT
ok for use as an ID, because html ID should be unique, and this has the
potential for a collision! for this reason, prefer html IDs with a dash

frontend parser for message text does not render links, it only renders
mentions, hashtags, styling, and diffs

i believe the bug in wormwatch where a video appears to be unavailable is a
result of adblockers in firefox. refreshing the page a few times seems to fix
it, but this is of course not ideal

## TODO
- like wormwatch entries
- #hug thru parser, and hug button thru not parser
- probably blocking usernames is needed at some point

## code style guide (or perhaps a style dictionary)

i'm a bit crazy idk it's probs typical in go, a lot of abbreviations. here are
some common variable names:

- id uint32 - a post/thread id
- pid uint32 - a post id
- tid uint32 - a thread id
- nid string - a base 36 (niftimal) post/thread id
- npid string - a base 36 (niftimal) post id
- ntid string - a base 36 (niftimal) thread id
- idtoa - convert id to alphanumeric (base 36)
- atoid - convert alphanumeric (base 36) to id
- cid string - content id, a hash of a file used as its location
- tmap - thread map
- wwd - wormwatchdata

etc... you're smart you'll probably get it all, the main weird thing is that we
use n to denote if an id is in base 36. i think in some places i use p at the
end of a function to say it accepts a pointer, in another place i use it to
mean we panic if it fails. f at the end of a function means we force (ignore
error)

## environment variables
- `SESSION_KEY` is for encrypting the cookie store
- `LRCD_SECRET` is for generating nonces on lrcd inits for several purposes,
but mostly to ensure that you don't try and post someone else's message
- `POSTGRES_BLAH` are self explanatory, used in our postgres container
- `YOUTUBE_API_KEY` is needed to get the duration & metadata of youtube videos
in wormwatch
- `ADMIN_USERNAME` is the username of the admin, which is like a moderator
except they can create more moderators in the app through the /administrate
endpoint. they don't need to be set as a moderator in the moderators table to
have moderator powers
- `DISCORD_LINK` is a url for a discord invite to a discord server that is
associated with your site
- `REPORT_DELIMITER` is a delimiter used to add the current in-progress status
of reported messages to report reasons, it's just nice to not expose to users
so they don't try and file misleading reports
- `MIGRATE_BIN` is not used by the application, but it is used by our scripts
to migrate
