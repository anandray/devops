package wolk

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"reflect"
	"unicode"

	"github.com/naoina/toml"
)

type Policy struct {
	DeltaPosTarget      uint64
	DeltaNegTarget      uint64
	StorageBetaTarget   uint64
	BandwidthBetaTarget uint64
	QTarget             uint64
	GammaTarget         uint64
}

const (
	DefaultPolicyWolkFile = "/root/go/src/github.com/wolkdb/cloudstore/policy.toml"
	MaxDeltaPosDiff       = 2000
	MaxDeltaNegDiff       = 2000
	MaxStorageBetaDiff    = 2000
	MaxBandwidthBetaDiff  = 4000
	MaxGammaDiff          = 100
)

var tomlSettings = toml.Config{
	NormFieldName: func(rt reflect.Type, key string) string {
		return key
	},
	FieldToKey: func(rt reflect.Type, field string) string {
		return field
	},
	MissingField: func(rt reflect.Type, field string) error {
		link := ""
		if unicode.IsUpper(rune(rt.Name()[0])) && rt.PkgPath() != "main" {
			link = fmt.Sprintf(", see https://godoc.org/%s#%s for available fields", rt.PkgPath(), rt.Name())
		}
		return fmt.Errorf("field '%s' is not defined in %s%s", field, rt.String(), link)
	},
}

func LoadPolicy(file string, policy *Policy) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	err = tomlSettings.NewDecoder(bufio.NewReader(f)).Decode(policy)
	// Add file name to errors that have a line number.
	if _, ok := err.(*toml.LineError); ok {
		err = errors.New(file + ", " + err.Error())
	}
	return err
}

func (self *Policy) AdjustUint64(attr string, curval uint64) uint64 {
	var target uint64
	var maxdiff int
	switch attr {
	case "deltaPos":
		target = self.DeltaPosTarget
		maxdiff = MaxDeltaPosDiff
		break
	case "deltaNeg":
		target = self.DeltaNegTarget
		maxdiff = MaxDeltaNegDiff
		break
	case "storageBeta":
		target = self.StorageBetaTarget
		maxdiff = MaxStorageBetaDiff
		break
	case "bandwidthBeta":
		target = self.BandwidthBetaTarget
		maxdiff = MaxBandwidthBetaDiff
		break
	case "gamma":
		target = self.GammaTarget
		maxdiff = MaxGammaDiff
		break
	}
	diff := curval - target
	if maxdiff > 0 { // abs(diff) < maxdiff
		if diff > 0 {

		} else if diff < 0 {

		}
	}
	return curval
}
