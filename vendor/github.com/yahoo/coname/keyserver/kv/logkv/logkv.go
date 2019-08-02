// Copyright 2014-2015 The Dename Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package logkv

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/yahoo/coname/keyserver/kv"
	"github.com/yahoo/coname/keyserver/kv/tracekv"
)

type traceLogger log.Logger

func toString(p tracekv.Update) string {
	if p.IsDeletion {
		return fmt.Sprintf("delete %q", p.Key)
	} else {
		return fmt.Sprintf("put %q = %q", p.Key, p.Value)
	}
}

func (l *traceLogger) put(p tracekv.Update) {
	(*log.Logger)(l).Print(toString(p))
}

func (l *traceLogger) batch(ps []tracekv.Update) {
	var ss []string
	for _, p := range ps {
		ss = append(ss, toString(p))
	}
	(*log.Logger)(l).Printf("batch{%s}", strings.Join(ss, "; "))
}

// WithDefaultLogging logs all Update-s and Write-s to db to os.Stdout with
// log.LstdFlags.
func WithDefaultLogging(db kv.DB) kv.DB {
	return WithLogging(db, log.New(os.Stdout, "", log.LstdFlags))
}

// WithLogging logs all Update-s and Write-s to db to l.
func WithLogging(db kv.DB, l *log.Logger) kv.DB {
	trace := (*traceLogger)(l)
	return tracekv.WithTracing(db, trace.put, trace.batch)
}
