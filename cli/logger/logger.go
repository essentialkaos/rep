package logger

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2026 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"os"
	"time"

	"github.com/essentialkaos/ek/v13/log"
	"github.com/essentialkaos/ek/v13/path"
	"github.com/essentialkaos/ek/v13/system"
)

// ////////////////////////////////////////////////////////////////////////////////// //

type Logger struct {
	loggers map[string]*SubLogger
	dir     string
	perms   os.FileMode
}

type SubLogger struct {
	lg *log.Logger
}

// ////////////////////////////////////////////////////////////////////////////////// //

// usernameCache is cached current user name
var usernameCache string

// ////////////////////////////////////////////////////////////////////////////////// //

var getUserFunc = system.CurrentUser

// ////////////////////////////////////////////////////////////////////////////////// //

// New creates new logger for CLI
func New(dir string, perms os.FileMode) *Logger {
	return &Logger{
		loggers: map[string]*SubLogger{},
		dir:     dir,
		perms:   perms,
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Add creates new sub-logger for given name
func (l *Logger) Add(name string) error {
	logFile := path.Join(l.dir, name+".log")
	lg, err := log.New(logFile, l.perms)

	if err != nil {
		return err
	}

	lg.EnableBufIO(time.Second)

	l.loggers[name] = &SubLogger{lg}

	return nil
}

// Get returns sub-logger with given name
func (l *Logger) Get(name string) *SubLogger {
	return l.loggers[name]
}

// Flush writes buffered data to file
func (l *Logger) Flush() {
	for _, sl := range l.loggers {
		sl.lg.Flush()
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Print writes message to log file
func (l *SubLogger) Print(f string, a ...interface{}) {
	if l == nil || l.lg == nil {
		return
	}

	l.lg.Info("("+getUserName()+") "+f, a...)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// getUserName returns current user real name
func getUserName() string {
	if usernameCache != "" {
		return usernameCache
	}

	curUser, err := getUserFunc()

	if err != nil {
		return "unknown"
	}

	usernameCache = curUser.RealName

	return usernameCache
}
