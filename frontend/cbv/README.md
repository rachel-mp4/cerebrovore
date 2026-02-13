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

