package main

import (
	"fmt"
	"log/syslog"
	"math/rand"
	"runtime"
)

func SysLog(info string, path string) error {

	thisOS := runtime.GOOS
	if thisOS != "linux" {
		return nil
	}

	logger, err := syslog.Dial("tcp", "127.0.0.1:5000", syslog.LOG_ERR, path) // connection to a log daemon
	if err != nil {
		return err
	}
	defer logger.Close()
	syslogerr := logger.Info(info)
	if syslogerr != nil {
		panic("nope")
	}
	fmt.Printf("logged %s to %s\n", info, path)
	return nil
}

func main() {
	thisOS := runtime.GOOS
	fmt.Printf("%s\n", thisOS)
	for stage := 0; stage < 5; stage++ {
		n := rand.Intn(10000000)
		SysLog(fmt.Sprintf("TRACE-%d", n), fmt.Sprintf("wolk-trace%d", stage))
		SysLog(fmt.Sprintf("MINING-%d", n), fmt.Sprintf("wolk-mining%d", stage))
		SysLog(fmt.Sprintf("TX-%d", n), fmt.Sprintf("wolk-tx%d", stage))
		SysLog(fmt.Sprintf("CHUNK-%d", n), fmt.Sprintf("wolk-chunk%d", stage))
	}
}

