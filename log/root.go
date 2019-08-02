package log

import (
	"fmt"
	"log/syslog"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	root          = &logger{ctx: []interface{}{}, h: new(swapHandler)}
	StdoutHandler = StreamHandler(os.Stdout, LogfmtFormat())
	StderrHandler = StreamHandler(os.Stderr, LogfmtFormat())
	wlogroot      = &WolkLog{}
	wlogrootMu    = new(sync.RWMutex)
)

func init() {
	root.SetHandler(DiscardHandler())
}

type WolkLog struct {
	verbosity Lvl
	log       Logger
	syslog    Logger
	filters   []string
}

// New returns a new logger with the given context.
// New is a convenient alias for Root().New

func NewLogger(ctx ...interface{}) Logger {
	return root.New(ctx...)
}

// Root returns the root logger
func Root() Logger {
	return root
}

// Output is a convenient alias for write, allowing for the modification of
// the calldepth (number of stack frames to skip).
// calldepth influences the reported line number of the log message.
// A calldepth of zero reports the immediate caller of Output.
// Non-zero calldepth skips as many stack frames.
func Output(msg string, lvl Lvl, calldepth int, ctx ...interface{}) {
	root.write(msg, lvl, ctx, calldepth+skipLevel)
}

func New(verbosity Lvl, filters string, store string) {
	wlogrootMu.Lock()
	wlogroot.verbosity = verbosity
	wlogroot.filters = strings.Split(filters, ",")
	wlogroot.log = Root()
	wlogroot.log.SetHandler(CallerFileHandler(LvlFilterHandler(LvlTrace, StreamHandler(os.Stderr, TerminalFormat(true)))))

	useSyslog := true
	if strings.HasSuffix(os.Args[0], ".test") {
		useSyslog = false
	}

	if useSyslog {
		wlogroot.syslog = NewLogger()
		syslogNetHandler, err := SyslogNetHandler("tcp", "127.0.0.1:5000", syslog.LOG_ERR, store, TerminalFormat(true))
		if err == nil {
			wlogroot.syslog.SetHandler(syslogNetHandler)
		}
	}

	wlogrootMu.Unlock()
}

// Mining syslogs to wolk-mining
func Mining(stage int, info string, bn uint64, duration time.Duration) {
	hostname, _ := os.Hostname()
	ts := time.Now().Format("15:04:05.99999999")
	ev := fmt.Sprintf("%s|stage:%d|bn:%d|%s|%s|%s\n", hostname, stage, bn, ts, info, duration)
	SysLog(ev, MINING)
}

// Chunk syslogs to wolk-netstats
func Chunk(chunkID []byte, info string) {
	hostname, _ := os.Hostname()
	ts := time.Now().Format("15:04:05.99999999")
	ev := fmt.Sprintf("%x|%s|%s|%s\n", chunkID, hostname, ts, info)
	SysLog(ev, CHUNK)
}

func Info(msg string, ctx ...interface{}) {
	wlogroot.Log(msg, LvlInfo, ctx...)
}

func Debug(msg string, ctx ...interface{}) {
	wlogroot.Log(msg, LvlDebug, ctx...)
}

func Trace(msg string, ctx ...interface{}) {
	wlogroot.Log(msg, LvlTrace, ctx...)
}

func Warn(msg string, ctx ...interface{}) {
	wlogroot.Log(msg, LvlWarn, ctx...)
}

func Error(msg string, ctx ...interface{}) {
	wlogroot.Log(msg, LvlError, ctx...)
}
func Crit(msg string, ctx ...interface{}) {
	wlogroot.Log(msg, LvlCrit, ctx)
}

func (wl *WolkLog) filterMatch(msg string) bool {
	wlogrootMu.RLock()
	defer wlogrootMu.RUnlock()
	for _, f := range wl.filters {
		if len(f) > 0 && strings.Contains(msg, f) {
			return true
		}
	}
	return false
}

func (wl *WolkLog) Log(msg string, lv Lvl, ctx ...interface{}) {

	if lv > wl.verbosity {
		return
	}

	switch lv {
	case LvlCrit:
		wl.log.Crit(msg, ctx...)
		if wl.syslog != nil {
			wl.syslog.Crit(msg, ctx...)
		}
	case LvlError:
		wl.log.Error(msg, ctx...)
		if wl.syslog != nil {
			wl.syslog.Error(msg, ctx...)
		}
	case LvlWarn:
		wl.log.Warn(msg, ctx...)
		if wl.syslog != nil {
			wl.syslog.Warn(msg, ctx...)
		}
	case LvlInfo:
		wl.log.Info(msg, ctx...)
		if wl.syslog != nil {
			wl.syslog.Info(msg, ctx...)
		}
	case LvlDebug:
		if wl.filterMatch(msg) {
			wl.log.Debug(msg, ctx...)
			if wl.syslog != nil {
				wl.syslog.Debug(msg, ctx...)
			}
		}
	case LvlTrace:
		if wl.filterMatch(msg) {
			wl.log.Trace(msg, ctx...)
			if wl.syslog != nil {
				wl.syslog.Trace(msg, ctx...)
			}
		}
	}
}
