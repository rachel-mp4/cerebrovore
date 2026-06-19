package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rachel-mp4/cerebrovore/clog"
	"github.com/rachel-mp4/cerebrovore/db"
	"github.com/rachel-mp4/cerebrovore/handler"
	"github.com/rachel-mp4/cerebrovore/id"
	"github.com/rachel-mp4/cerebrovore/model"
)

func main() {
	cold := flag.Bool("cold", false, "disables hot module replacement")
	port := flag.Int("port", 8080, "port to listen on")
	sidp := flag.Int("sidp", 9009, "uses a service id provider listening on port")
	dontmock := flag.Bool("db", false, "doesn't mock the database")
	midp := flag.Bool("midp", false, "uses an in memory id provider instead of service id provider")
	dev := flag.Bool("dev", false, "run in dev mode (file logging, debug output)")
	flag.Parse()

	clog.Dev = *dev
	if err := clog.Init("cerebrovore.log"); err != nil {
		clog.Warn("could not open log file: %s", err)
	}
	defer clog.Close()

	clog.Info("*eats ur brain*")

	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
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
			ChatPath:     ms.Chat.File,
			ChatCss:      ms.Chat.CSS,
			BeepPath:     ms.Beep.File,
			BeepCss:      ms.Beep.CSS,
			WatcherPath:  ms.Watcher.File,
			WatcherCss:   ms.Watcher.CSS,
			WormPath:     ms.Worm.File,
			WormCss:      ms.Worm.CSS,
			SettingsPath: ms.Settings.File,
			SettingsCss:  ms.Settings.CSS,
		}
	}
	var store db.Storer
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

	var reqcode bool
	var idp id.Provider
	if *midp {
		idp = id.NewMemoryProvider()
		reqcode = false
	} else {
		idp = id.NewServiceProvider(*sidp)
		reqcode = true
	}

	// in order to initialize our model of the threads, we need to get
	// all threads. in truth this should be get all threads that haven't
	// hit post limit, but the post limit does not yet exist
	threads, err := store.GetAllThreads(context.Background())
	first := false
	if err != nil {
		if !clog.InputYN("is this your first time running on this database?") {
			panic(err)
		}
		clog.Okay("good luck!")
		first = true
	}
	clog.Info("clearing old selfbans")
	nrows, err := store.ClearOldSelfBans(context.Background())
	if err != nil && !first {
		if !clog.InputYN("is this your first time running on this database?") {
			panic(err)
		}
		clog.Okay("good luck!")
		first = true
	}
	clog.Info("cleared %d selfbans", nrows)
	// we also need the max id in order to allocate post ids properly
	// cross site. i'm not sure if this is exactly ideal, because it
	// couples all the threads, but it seems cool + you have to make
	// dumb decisions to learn
	mid, err := store.GetMaxPostId(context.Background())
	if err != nil && !first {
		if !clog.InputYN("is this your first time running on this database?") {
			panic(err)
		}
		clog.Okay("good luck!")
		first = true
		mid = 0
	}
	m := model.NewModel(threads, mid)
	h := handler.NewHandler(ca, m, store, idp, reqcode)
	// catch sigint and sigterm (ctrl+c)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		// i add this just in case for whatever stupid reason clog blocks & oh no!
		// the program hangs + doesn't respond to SIGINT / SIGTERM
		go func() {
			<-sig
			fmt.Println("")
			os.Exit(1)
		}()

		m.SystemMessage(`cerebrovore shutdown sequence initiated!
you have 15 seconds to get your affairs in order
good luck!`)
		fmt.Println("")
		clog.Okay("shutdown squence initiated... or ctrl-c again to force")
		time.Sleep(15 * time.Second)
		clog.Okay("my brain has been eaten, GOOD BYE")
		clog.Close()
		os.Exit(0)
	}()

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
	Worm struct {
		File string   `json:"file"`
		CSS  []string `json:"css,omitempty"`
	} `json:"src/worm.ts"`
	Settings struct {
		File string   `json:"file"`
		CSS  []string `json:"css,omitempty"`
	} `json:"src/settings.ts"`
}
