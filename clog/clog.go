package clog

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"
)

// parity with ./d script for display
var (
	r  string // red
	g  string // green
	y  string // yellow
	b  string // blue
	p  string // purple
	w  string // white
	bd string // bold
	dm string // dim
	rt string // reset

	// Dev controls logging behavior
	// true - log to file (./cerebrovore.log)
	// false - log to stdout (for journald), debug suppressed
	Dev bool

	logFile *os.File
	logMu   sync.Mutex
)

// grab log.Println output
type clogWriter struct{}

func (clogWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimSpace(string(p))
	Info("%s", msg)
	return len(p), nil
}

func init() {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		r = "\033[31m"
		g = "\033[32m"
		y = "\033[33m"
		b = "\033[34m"
		p = "\033[35m"
		w = "\033[37m"
		bd = "\033[1m"
		dm = "\033[2m"
		rt = "\033[0m"
	}

	// override the default logger
	log.SetFlags(0)
	log.SetOutput(io.Writer(clogWriter{}))
}

// Init sets up logging based on Dev mode.
// dev:  logs to the given file path with timestamps
// prod: logs to stdout as plain text (for journald/systemd)
func Init(logPath string) error {
	if Dev {
		f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		logFile = f
	}
	return nil
}

func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

func writeLog(level string, msg string) {
	logMu.Lock()
	defer logMu.Unlock()
	if Dev && logFile != nil {
		ts := time.Now().Format(time.RFC3339) // insane, whatever
		fmt.Fprintf(logFile, "%s [%s] %s\n", ts, level, msg)
	} else if !Dev {
		ts := time.Now().Format(time.RFC3339)
		fmt.Printf("%s [%s] %s\n", ts, level, msg)
	}
}

func Okay(msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	fmt.Printf("%s[%s+%s%s] %s%s\n", g, bd, rt, g, formatted, rt)
	writeLog("OKAY", formatted)
}

func Warn(msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	fmt.Printf("%s[%s*%s%s] %s%s\n", y, bd, rt, y, formatted, rt)
	writeLog("WARN", formatted)
}

func Fail(msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	fmt.Printf("%s[%s!%s%s] %s%s\n", r, bd, rt, r, formatted, rt)
	writeLog("FAIL", formatted)
}

func Info(msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	fmt.Printf("[~] %s\n", formatted)
	writeLog("INFO", formatted)
}

func Tmpl(err error) {
	if err == nil {
		return
	}
	formatted := fmt.Sprintf("tmpl error: %s", err.Error())
	fmt.Printf("[T] %s\n", formatted)
	writeLog("TMPL", formatted)
}

func LogE(err error) {
	if err == nil {
		return
	}
	formatted := fmt.Sprintf("error: %s", err.Error())
	fmt.Printf("[E] %s\n", formatted)
	writeLog("LOGE", formatted)
}

// use this for meta messages, just for you!
func Dbug(msg string, args ...any) {
	if !Dev {
		return
	}
	formatted := fmt.Sprintf(msg, args...)
	fmt.Printf("%s[&] %s%s\n", dm, formatted, rt)
}

func InputYN(prompt string) bool {
	fmt.Printf("%s[%s%s?%s%s]%s %s %sY%s/%sN%s: ", w, p, bd, rt, w, rt, prompt, r, w, g, rt)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(answer)
	return strings.EqualFold(answer, "y")
}
