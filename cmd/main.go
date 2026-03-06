package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/rachel-mp4/cerebrovore/db"
	"github.com/rachel-mp4/cerebrovore/handler"
	"github.com/rachel-mp4/cerebrovore/model"
)

func main() {
	fmt.Println("*eats ur brain*")
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
	cold := flag.Bool("cold", false, "disables hot module replacement")
	port := flag.Int("port", 8080, "port to listen on")
	dontmock := flag.Bool("db", false, "doesn't mock the database")
	idp := flag.Bool("idp", false, "doesn't mock the id provider")
	flag.Parse()
	var ca *handler.CompiledAssets
	// if not doing hot module replacement, we must read the manifest to figure out
	// where the required scripts are, so that way we can link them accordingly when
	// we template our html
	if *cold {
		manifest, err := os.ReadFile("./frontend/dist/.vite/manifest.json")
		if err != nil {
			panic(err)
		}
		var ms Manifest
		err = json.Unmarshal(manifest, &ms)
		if err != nil {
			panic(err)
		}
		ca = &handler.CompiledAssets{
			ChatPath:    ms.Chat.File,
			ChatCss:     ms.Chat.CSS,
			BeepPath:    ms.Beep.File,
			BeepCss:     ms.Beep.CSS,
			WatcherPath: ms.Watcher.File,
			WatcherCss:  ms.Watcher.CSS,
		}
	}
	var store db.Storer
	if *dontmock && !*idp {
		fmt.Println("WARNING WARNING WARNING NOT MOCKING DB AND MOCKING IDP")
		fmt.Println("IF THIS IS PROD, USERS CAN JUST SET THEIR SESSION ID")
		fmt.Println("TO WHATEVER THEY WANT")
		fmt.Println("press enter to confirm")
		fmt.Scanln()
	}
	if *dontmock {
		realstore, err := db.Init()
		if err != nil {
			panic(err)
		}
		store = realstore
	} else {
		mockstore, err := db.MockInit()
		if err != nil {
			panic(err)
		}
		store = mockstore
	}

	// in order to initialize our model of the threads, we need to get
	// all threads. in truth this should be get all threads that haven't
	// hit post limit, but the post limit does not yet exist
	threads, err := store.GetAllThreads(context.Background())
	if err != nil {
		panic(err)
	}
	// we also need the max id in order to allocate post ids properly
	// cross site. i'm not sure if this is exactly ideal, because it
	// couples all the threads, but it seems cool + you have to make
	// dumb decisions to learn
	mid, err := store.GetMaxPostId(context.Background())
	if err != nil {
		panic(err)
	}
	m := model.NewModel(threads, mid)
	h := handler.NewHandler(ca, m, store, *idp)
	http.ListenAndServe(fmt.Sprintf(":%d", *port), h.Serve())
}

// Manifest is a json file generated when we compile our frontend
// that maps out where the scripts and css ended up
type Manifest struct {
	Chat struct {
		File string   `json:"file"`
		CSS  []string `json:"css,omitempty"`
	} `json:"src/chat.ts"`
	Beep struct {
		File string   `json:"file"`
		CSS  []string `json:"css,omitempty"`
	} `json:"src/beep.ts"`
	Watcher struct {
		File string   `json:"file"`
		CSS  []string `json:"css,omitempty"`
	} `json:"src/watcher.ts"`
}
