// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (
	"fmt"
)

func tprint(in string, args ...interface{}) {
	if in == "\n" {
		fmt.Println()
	} else {
		fmt.Printf("[test] "+in+"\n", args...)
	}
}

type Expected struct {
	Rows             []interface{}
	AffectedRowCount int
}
