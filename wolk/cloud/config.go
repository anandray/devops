package cloud

import (
	"bufio"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/naoina/toml"
	"github.com/wolkdb/cloudstore/crypto"
	"os"
	"reflect"
	"unicode"
)

func LoadConfig(file string, cfg *Config) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	err = tomlSettings.NewDecoder(bufio.NewReader(f)).Decode(cfg)
	// Add file name to errors that have a line number.
	if _, ok := err.(*toml.LineError); ok {
		err = errors.New(file + ", " + err.Error())
	}
	return err
}

// DefaultConfig contains default settings for use on the Ethereum main net.
var DefaultConfig = Config{
	DataDir:           "/usr/local/wolk",
	HTTPPort:          80,
	GenesisFile:       "/root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/genesis.json",
	KeyStoreDir:       "/root/go/src/github.com/wolkdb/cloudstore/wolk/cloud/credentials/keystore",
	ConsensusIdx:       -1,
	Provider:           "cephif",
	ConsensusAlgorithm: "algorand",
	NodeType:           "storage",
	SSLCertFile:        "/etc/ssl/certs/wildcard.wolk.com/www.wolk.com.crt",
	SSLKeyFile:         "/etc/ssl/certs/wildcard.wolk.com/www.wolk.com.key",
	Preemptive:         false,
	CephConfig:         "/etc/ceph/ceph.conf",
	CephCluster:        "z",
}

type Config struct {
	NodeType           string
	ConsensusIdx       int
	ConsensusAlgorithm string
	Preemptive         bool
	HTTPPort int

	Provider         string // leveldb, amazon_dynamo, google_bigtable, google_datastore,
	Address          common.Address
	OperatorKey      *crypto.PrivateKey
	OperatorECDSAKey *ecdsa.PrivateKey
	KeyStoreDir      string
	Password         string

	// genesis file
	GenesisFile string

	DataDir  string

	NetworkID     uint64
	StartBlock    int64

	TrustedNode string
	Identity    string
	SSLCertFile string
	SSLKeyFile  string

	StaticNodes  []*enode.Node
	TrustedNodes []*enode.Node

	CephConfig  string
	CephCluster string
}

const DefaultConfigWolkFile = "/root/go/src/github.com/wolkdb/cloudstore/wolk.toml"

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
