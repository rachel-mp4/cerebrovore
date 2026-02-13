package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

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
	prod := flag.Bool("prod", false, "runs prod")
	port := flag.Int("port", 8080, "port to listen on")
	flag.Parse()
	var hp *handler.Prod
	if *prod {
		manifest, err := os.ReadFile("./frontend/dist/.vite/manifest.json")
		if err != nil {
			panic(err)
		}
		var ms Manifest
		err = json.Unmarshal(manifest, &ms)
		if err != nil {
			panic(err)
		}
		hp = &handler.Prod{
			ChatPath: ms.Chat.File,
			ChatCss:  ms.Chat.CSS,
			BeepPath: ms.Beep.File,
			BeepCss:  ms.Beep.CSS,
		}
	}

	m := model.NewModel()
	h := handler.NewHandler(hp, m)
	http.ListenAndServe(fmt.Sprintf(":%d", *port), h.Serve())
}
