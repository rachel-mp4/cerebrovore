## 1 dev environment 

prefer `npm run dev` + `go run ./cmd` in two separate terminals while
developing frontend; this way you get vite hot module replacement!

## 2 prod environment

use `npm run build` to compile svelte, then `go run ./cmd -prod` in just one
terminal

## 3 adding svelte "islands" 

ideally we just use svelte for a few isolated things that require additional
client-side state, and leave the rest to go templating. if you need to add a
new island

1. copy `beep.ts` to `new-island.ts` 
   1. if you need this island to be on the same page as an existing island, you
      must rename the id that it mounts to!
2. copy `Beep.svelte` to `New-Island.svelte`
3. add to roll-up in vite.config.ts. you can give it whatever name you want for
   the key, but pick New-Island. the value of course must be the path
`src/beep.ts` 
4. in `handler/handler.go`, add a field to Prod struct to represent the path
   that our compiled svelte will end up at, and a second field to represent the
paths that our compiled css will end up at. ideally, we won't use any css in
svelte components, because this makes it harder for users to write their own
css, however, it is probably faster to write component level css initially, and
then we want to copy it over, and if you don't make a field for the array of
css paths, then there will be differences in appearence between dev + prod
5. in `cmd/main.go`, extend the json schema for `Manifest` accordingly, and
   after unmarshaling, copy the path into the new field in the `handler.Prod`
struct, and copy the array of css paths into the handler struct
6. in our template that has this new island, we need both a div with the id
   specified in `new-island.ts`, and a `script type="module"` with a link to
our code, and in prod, we need to range over all css paths and add links to
them so look at `tmpl/beep.html`; but in the case where we are in prod, we want
to access the field of prod for the compiled code, and in the case where we are
in dev, we can just link "straight to it" and vite handles the rest

# included islands

## beep

just a basic svelte thing for reference

## chat

lrc chat implementation. see [lrcproto](https://github.com/rachel-mp4/lrcproto)
and [lrcd](https://github.com/rachel-mp4/lrcd) to get a sense for the protocol

## watcher 

live thread watcher. bumps are recieved as json; the final bump is gonna include 
field bumplimit value true. new threads have the field new as true

## worm

wormwatch. wormwatch events are sent as json, the initial connection should
include the server time (type "timeS"), and if there is currently a wormwatch
queue, the state of the queue (type "queue"). if the queue is not currently
paused, it will also send an event with the server timestamp that the video
started (or will start) playing at (type "start"). if the queue is ever paused,
it will send an event (type "pause") to all clients. when the final video in
queue finishes, server sends an event letting clients know (type "clear")

# websocket stuff

if you wanna grab something from a websocket, we buffer any events set on the 
initial load. since an island might run after the websocket inital load, we
want to grab stuff the from the buffer before we handle any events. this is
the pattern that we use. btw `cbvWSBuffer` is the object that holds the buffers
for each scope, defined globally in static/js/websockets.js


```javascript
//ex: for wormwatch

// @ts-ignore
const wws = cbvWSBuffer?.wormwatch //the name of the key will be the type sent
if (wws !== undefined) {
   wws.forEach(handleWormwatchEvent) // key is an array of the data, which gets
}                                    // sent as detail of CustomEvent that we
                                     // wanna listen for
document.addEventListener("cbv:wormwatch", (e) => {
   const ev = e as CustomEvent
   const wwe = ev.detail
   handleWormwatchEvent(wwe)
})
```
