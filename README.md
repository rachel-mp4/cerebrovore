```
 ⢀⣀ ⢀⡀ ⡀⣀ ⢀⡀ ⣇⡀ ⡀⣀ ⢀⡀ ⡀⢀ ⢀⡀ ⡀⣀ ⢀⡀
 ⠣⠤ ⠣⠭ ⠏  ⠣⠭ ⠧⠜ ⠏  ⠣⠜ ⠱⠃ ⠣⠜ ⠏  ⠣⠭
```

initial setup
1. install go
2. install npm
3. `cd frontend/cbv`
4. `npm install`

dev build
1. `cd frontend/cbv`
2. `npm run dev`
3. open another terminal
4. `go run ./cmd`

prod build
1. `cd frontend/cbv`
2. `npm run build`
3. `cd -`
4. `go run ./cmd -prod`

### KNOWN BUGS AND BAD BEHAVIOR THATS ANNOYING 
when you click on a thumbnail, swapping it out for the full-size image, it's
bad need to store the size (aspect ratio dimensions size), broadcast over lrc
etc...

mock db is completely useless, need to upgrade it to an in memory db at least
lol
