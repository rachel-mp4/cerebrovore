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

type Manifest struct {
	Chat struct {
		File string   `json:"file"`
		CSS  []string `json:"css,omitempty"`
	} `json:"src/chat.ts"`
	Beep struct {
		File string   `json:"file"`
		CSS  []string `json:"css,omitempty"`
	} `json:"src/beep.ts"`
}

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
			ChatPath: ms.Chat.File,
			ChatCss:  ms.Chat.CSS,
			BeepPath: ms.Beep.File,
			BeepCss:  ms.Beep.CSS,
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

	threads, err := store.GetAllThreads(context.Background())
	if err != nil {
		panic(err)
	}
	mid, err := store.GetMaxPostId(context.Background())
	if err != nil {
		panic(err)
	}
	m := model.NewModel(threads, mid)
	h := handler.NewHandler(ca, m, store, *idp)
	http.ListenAndServe(fmt.Sprintf(":%d", *port), h.Serve())
}
