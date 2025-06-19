package tklog

import (
	"fmt"
	"log"
	"os"
)

func Debug(msg string, args ...any) {
	ensureLogger().Output(2, fmt.Sprintf("DBG: "+msg+"\r", args...))
}

func Info(msg string, args ...any) {
	ensureLogger().Output(2, fmt.Sprintf("INF: "+msg+"\r", args...))
}

func Warn(msg string, args ...any) {
	ensureLogger().Output(2, fmt.Sprintf("WRN: "+msg+"\r", args...))
}

func Error(msg string, args ...any) {
	ensureLogger().Output(2, fmt.Sprintf("ERR: "+msg+"\r", args...))
}

func Fatal(msg string, args ...any) {
	ensureLogger().Output(2, fmt.Sprintf("FTL: "+msg+"\r", args...))
	os.Exit(1)
}

func Panic(msg string, args ...any) {
	s := fmt.Sprintf("PAN: "+msg+"\r", args...)
	ensureLogger().Output(2, s)
	panic(s)
}

var logger *log.Logger

func ensureLogger() *log.Logger {
	if logger == nil {
		f, err := os.Create("tked.log")
		if err != nil {
			panic(err)
		}

		logger = log.New(f, "TKED ", log.Lshortfile|log.LstdFlags)
	}

	return logger
}
