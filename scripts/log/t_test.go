package wolklog

import (
	"fmt"
	"log/syslog"
	"runtime"
	"testing"
	"time"
)

const DEBUG = "wolk-debug"
const CLOUD = "wolk-cloud"

func TestLog(t *testing.T) {
	str := fmt.Sprintf("hi:niyogi:%s", time.Now())
	nentries := 1000
	for i := 0; i < nentries; i++ {
		SyslogSimple(str, CLOUD)
		// Syslog(str, CLOUD)
	}
	fmt.Printf("DONE: %s", str)
}

func SyslogSimple(info string, path string) error {
	thisOS := runtime.GOOS
	if thisOS != "linux" {
		return nil
	}
	//path = wolk-cloud for timing lines
	//path = wolk-debug for debug lines
	logger, err := syslog.Dial("udp", "127.0.0.1:514", syslog.LOG_INFO, path) // connection to a log daemon
	if err != nil {
		fmt.Printf("ERR %v\n", err)
		return err
	}
	err = logger.Info(info)
	logger.Close()
	return err
}
