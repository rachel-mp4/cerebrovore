package clog

import (
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

func writeLog(level string, fmsg string) {
	logMu.Lock()
	defer logMu.Unlock()
	if Dev {
		switch level {
		case okay:
			okayP(fmsg)
		case warn:
			warnP(fmsg)
		case fail:
			failP(fmsg)
		case info:
			infoP(fmsg)
		case tmpl:
			tmplP(fmsg)
		case loge:
			logeP(fmsg)
		case dbug:
			dbugP(fmsg)
		}
		if logFile != nil {
			ts := time.Now().Format(time.RFC3339) // insane, whatever
			fmt.Fprintf(logFile, "%s [%s] %s\n", ts, level, fmsg)
		}
	} else {
		ts := time.Now().Format(time.RFC3339)
		fmt.Printf("%s [%s] %s\n", ts, level, fmsg)
	}
}

const (
	okay = "OKAY"
	warn = "WARN"
	fail = "FAIL"
	info = "INFO"
	tmpl = "TMPL"
	loge = "LOGE"
	dbug = "DBUG"
)

func okayP(formatted string) {
	fmt.Printf("%s[%s+%s%s] %s%s\n", g, bd, rt, g, formatted, rt)
}

func Okay(msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	writeLog(okay, formatted)
}

func warnP(formatted string) {
	fmt.Printf("%s[%s*%s%s] %s%s\n", y, bd, rt, y, formatted, rt)
}

func Warn(msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	writeLog(warn, formatted)
}

func failP(formatted string) {
	fmt.Printf("%s[%s!%s%s] %s%s\n", r, bd, rt, r, formatted, rt)
}

func Fail(msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	writeLog(fail, formatted)
}

func infoP(fmsg string) {
	fmt.Printf("[~] %s\n", fmsg)
}

func Info(msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	writeLog(info, formatted)
}

func tmplP(fmsg string) {
	fmt.Printf("[T] %s\n", fmsg)
}

func Tmpl(err error) {
	if err == nil {
		return
	}
	formatted := fmt.Sprintf("tmpl error: %s", err.Error())
	writeLog(tmpl, formatted)
}

func logeP(fmsg string) {
	fmt.Printf("[E] %s\n", fmsg)
}

func LogE(err error, in string) {
	if err == nil {
		return
	}
	formatted := fmt.Sprintf("error in %s: %s", in, err.Error())
	writeLog(loge, formatted)
}

func dbugP(fmsg string) {
	fmt.Printf("%s[&] %s%s\n", dm, fmsg, rt)
}

func Dbug(msg string, args ...any) {
	if !Dev {
		return
	}
	formatted := fmt.Sprintf(msg, args...)
	writeLog(dbug, formatted)
}

func InputYN(prompt string) bool {
	fmt.Printf("%s[%s%s?%s%s]%s %s %sY%s/%sN%s: ", w, p, bd, rt, w, rt, prompt, r, w, g, rt)
	var answer string
	_, err := fmt.Scanln(&answer)
	if err != nil {
		LogE(err, "inputYN")
		return false
	}
	return strings.EqualFold(answer, "y")
}
