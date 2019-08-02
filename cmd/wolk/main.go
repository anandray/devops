// Copyright 2014 Wolk Inc.
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

// plasma is the official command-line client for Ethereum.
package main

import (
	"fmt"
	"os"
	"runtime"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/fdlimit"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"github.com/wolkdb/cloudstore/cmd/utils"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/console"
	"github.com/wolkdb/cloudstore/crypto"
	"github.com/wolkdb/cloudstore/internal/debug"
	"github.com/wolkdb/cloudstore/log"
	"github.com/wolkdb/cloudstore/wolk"
	wolkcloud "github.com/wolkdb/cloudstore/wolk/cloud"
	cli "gopkg.in/urfave/cli.v1"
)

const (
	clientIdentifier = "cloud" // Client identifier to advertise over the network
)

var (
	// Git SHA1 commit hash of the release (set via linker flags)
	gitCommit = ""
	// Ethereum address of the plasma release oracle.
	relOracle = common.HexToAddress("0xfa7b9770ca4cb04296cac84f37736d4041251cdf")
	// The app that holds all commands and flags.
	app = utils.NewApp(gitCommit, "the go-ethereum command line interface")
	// flags that configure the node
	nodeFlags = []cli.Flag{
		utils.IdentityFlag,
		utils.UnlockedAccountFlag,
		utils.PasswordFileFlag,
		utils.BootnodesFlag,
		utils.BootnodesV4Flag,
		utils.BootnodesV5Flag,
		utils.DataDirFlag,
		utils.KeyStoreDirFlag,
		utils.NoUSBFlag,
		// utils.PlasmaDataDirFlag,
		utils.TxPoolNoLocalsFlag,
		utils.TxPoolJournalFlag,
		utils.TxPoolRejournalFlag,
		utils.TxPoolPriceLimitFlag,
		utils.TxPoolPriceBumpFlag,
		utils.TxPoolAccountSlotsFlag,
		utils.TxPoolGlobalSlotsFlag,
		utils.TxPoolAccountQueueFlag,
		utils.TxPoolGlobalQueueFlag,
		utils.TxPoolLifetimeFlag,
		utils.FastSyncFlag,
		utils.LightModeFlag,
		utils.SyncModeFlag,
		utils.GCModeFlag,
		utils.LightServFlag,
		utils.LightPeersFlag,
		utils.LightKDFFlag,
		utils.CacheFlag,
		utils.CacheDatabaseFlag,
		utils.CacheGCFlag,
		utils.TrieCacheGenFlag,
		utils.ListenPortFlag,
		utils.MaxPeersFlag,
		utils.MaxPendingPeersFlag,
		utils.TargetGasLimitFlag,
		utils.NATFlag,
		utils.NoDiscoverFlag,
		utils.DiscoveryV5Flag,
		utils.NetrestrictFlag,
		utils.NodeKeyFileFlag,
		utils.NodeKeyHexFlag,
		utils.DeveloperFlag,
		utils.DeveloperPeriodFlag,
		utils.TestnetFlag,
		utils.RinkebyFlag,
		utils.VMEnableDebugFlag,
		utils.NetworkIdFlag,
		utils.RPCCORSDomainFlag,
		utils.RPCVirtualHostsFlag,
		utils.MetricsEnabledFlag,
		utils.ExtraDataFlag,
		utils.PreemptiveState,
		configFileFlag,
	}

	rpcFlags = []cli.Flag{
		utils.RPCEnabledFlag,
		utils.RPCListenAddrFlag,
		utils.RPCPortFlag,
		utils.RPCApiFlag,
		utils.WolkLogFlag,
		utils.WolkPortFlag,
		utils.WolkProviderFlag,
		utils.WolkRegionFlag,
		utils.EmitCheckpointsFlag,
		utils.WSEnabledFlag,
		utils.WSListenAddrFlag,
		utils.WSPortFlag,
		utils.WSApiFlag,
		utils.WSAllowedOriginsFlag,

		utils.IPCDisabledFlag,
		utils.IPCPathFlag,
	}

	configFileFlag = cli.StringFlag{
		Name:  "config",
		Usage: "TOML configuration file",
	}
)

func init() {
	// Initialize the CLI app and start plasma
	app.Action = cloud
	app.HideVersion = true // we have a command to print the version
	app.Copyright = "Copyright 2018 Wolk Inc"
	app.Commands = []cli.Command{
		consoleCommand,
		attachCommand,
		javascriptCommand,
	}
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Flags = append(app.Flags, nodeFlags...)
	app.Flags = append(app.Flags, debug.Flags...)
	app.Flags = append(app.Flags, rpcFlags...)

	app.Before = func(ctx *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		if err := debug.Setup(ctx); err != nil {
			return err
		}

		// Start system runtime metrics collection
		go metrics.CollectProcessMetrics(3 * time.Second)

		utils.SetupNetwork(ctx)
		return nil
	}

	app.After = func(ctx *cli.Context) error {
		debug.Exit()

		console.Stdin.Close() // Resets terminal mode.
		return nil
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func defaultNodeConfig() *node.Config {
	cfg := node.DefaultConfig
	cfg.Name = clientIdentifier
	cfg.Version = params.VersionWithCommit(gitCommit)
	cfg.IPCPath = "wolk.ipc"
	return &cfg
}

func makeFullNode(ctx *cli.Context) *node.Node {
	fdlimit.Raise(4096)
	// config is DefaultConfig
	fdlimit.Raise(8192)
	cfg := &(wolkcloud.DefaultConfig)
	var logfilters string
	if ctx.GlobalIsSet(utils.WolkLogFlag.Name) {
		logfilters = ctx.GlobalString(utils.WolkLogFlag.Name)
	}
	if ctx.GlobalIsSet(utils.WolkPortFlag.Name) {
		cfg.HTTPPort = ctx.GlobalInt(utils.WolkPortFlag.Name)
	}

	if ctx.GlobalBool(utils.PreemptiveState.Name) {
		cfg.Preemptive = true
	}

	offset := cfg.HTTPPort - 80
	if offset > 9 {
		offset = 9
	}
	log.New(log.LvlTrace, logfilters, fmt.Sprintf("wolk-trace%1d", offset))

	// Make node with node config
	node_cfg := defaultNodeConfig()
	utils.SetNodeConfig(ctx, node_cfg)

	// Wolk Config "/root/go/src/github.com/wolkdb/cloudstore/wolk.toml" is specified on command line --config
	cfg_file := wolkcloud.DefaultConfigWolkFile
	if file := ctx.GlobalString(configFileFlag.Name); file != "" {
		cfg_file = file
	}
	if _, err := os.Stat(cfg_file); !os.IsNotExist(err) {
		if err := wolkcloud.LoadConfig(cfg_file, cfg); err != nil {
			log.Error("CONFIG ERROR", "err", err)
			os.Exit(0)
		}
	}
	log.Info("Wolk Config LOADED", "Provider", cfg.Provider, "ConsensusIdx", cfg.ConsensusIdx)

	if identity := ctx.GlobalString(utils.IdentityFlag.Name); len(identity) == 0 {
		userIdentity := fmt.Sprintf("%v%d", cfg.NodeType, cfg.ConsensusIdx)
		node_cfg.UserIdent = userIdentity
		node_cfg.Name = "wolk"
	}

	k_str := fmt.Sprintf("%x", wolkcommon.Computehash([]byte(fmt.Sprintf("%d", cfg.ConsensusIdx))))
	epKey, _ := ethcrypto.HexToECDSA(k_str)

	//previously, ConsensusIdx >= 0 is consensus node
	// if cfg.ConsensusIdx >= 0 {
	// 	node_cfg.P2P.PrivateKey = epKey
	// }

	if cfg.NodeType == "consensus" {
		node_cfg.P2P.PrivateKey = epKey
	}
	log.Info("makeFullNode", "P2P.PrivateKey", node_cfg.P2P.PrivateKey)
	privateKey, err := crypto.HexToPrivateKey(k_str)
	if err != nil {
	} else {
		log.Info("TESTNET PrivateKey", "privatekey", k_str)
		cfg.OperatorKey = privateKey

		alt := make(map[int]string)
		alt[0] = `{"kty":"EC","crv":"P-256","d":"yFToI9sZjyxt6gjPGOTuCTmNGOGBdmP39gdNKpk3pBo","x":"vk2WAZwjks828V4ls4h_ucehfCX2DlAr6xehp7o8ShE","y":"Pe4WdjQfJtKqVBU9NzsBE15kiOPNEnz4BcQIZtKWYVc","key_ops":["sign"],"ext":true}`
		alt[1] = `{"kty":"EC","crv":"P-256","d":"l1Ie64hwxf8Y5YBIfoDqcnCT3dZSjEMFt22jajspwTo","x":"ernbGAGxqIX0Y5gH2jicT3HASW6uDkbeDn98BZRoWVU","y":"X2mz5KtorHTmLyJ_MHkx_CpdMW1GvDZE2CclccEKI8Y","key_ops":["sign"],"ext":true}`
		alt[2] = `{"kty":"EC","crv":"P-256","d":"d-EaLm5MtclJhP89QKYB_SWyxsddPly75r-O5vGJb7Y","x":"XD3vELP2lYimgDGbR1GQDjaf0f4hMPUjq_NjCdCBEYI","y":"JYwcOY_DFyrofusvyP9k7LYRm-DArvHwREl5OEdKZYQ","key_ops":["sign"],"ext":true}`
		alt[3] = `{"kty":"EC","crv":"P-256","d":"KyaJiZ1SZ73sEyWVSW4qXjZD2-rtpcaJBctLlx90Si0","x":"Kkjomu20LRmWxlaZasGKRJg5GFge2th8SvaSI4oWwAY","y":"3Cxv-XdTfLnCsw5myB34qCe0QQsEOdS6QVjjg2yCzeA","key_ops":["sign"],"ext":true}`
		alt[4] = `{"kty":"EC","crv":"P-256","d":"2JfpcA9A9bGqaJ1sp2I9GYcMzXzQ70Tvqxx5lu9JGUE","x":"iv9d0JFPfYzmqCqemWbmfrp4OfhV71L200BqwNuFsAc","y":"TOBt6vWwdXK_8SaQo2aAFQqHC-7-t_J2vOP2fRtSlMY","key_ops":["sign"],"ext":true}`
		alt[5] = `{"kty":"EC","crv":"P-256","d":"OC7Udt5lBIugK1Td8pWpUntG1pfpL9-JVq5NnChoQqc","x":"0_Xk2kWluLih4rFmCuimmWI1l2z-bihkWhRVMygLMeU","y":"Z4j71vW936T838W55Qneq8XGzMrS-nY6cK3X_TinM10","key_ops":["sign"],"ext":true}`
		alt[6] = `{"kty":"EC","crv":"P-256","d":"gaajWY7fGUP9ukUqC8Z__4YYt4Om-eqndPfrLExliFA","x":"te-xWItYSF5f8k3QbXjflRwkYEWscGsKjfe1TFU7gTU","y":"bLSHs8tiEFRp74qXaiw5Oj2QgWe6HQFs87xsrcwR4xM","key_ops":["sign"],"ext":true}`
		alt[7] = `{"kty":"EC","crv":"P-256","d":"HkJujHtL7R1ZJafOwL4qWNkYEZK_rYhV5G3vG1DVIl8","x":"dQg8EKJwOc8iNgSUHgFXrYEB9dwZgMZ9_nRnAOlAgOg","y":"X6Eg2dnbLxDrRcnkJ3r6lly01ax1k5y2QunmfhEA9hs","key_ops":["sign"],"ext":true}`
		cfg.OperatorECDSAKey, err = crypto.JWKToECDSA(alt[0])
		if err != nil {
			log.Error("JWKToECDSA", "err", err)
		}
	}
	genesisConfig, err := wolk.LoadGenesisFile(cfg.GenesisFile)

	if ctx.GlobalIsSet(utils.WolkProviderFlag.Name) {
		cfg.Provider = ctx.GlobalString(utils.WolkProviderFlag.Name)
	}


	if ctx.GlobalIsSet(utils.WolkPortFlag.Name) {
		cfg.HTTPPort = ctx.GlobalInt(utils.WolkPortFlag.Name)
	}
	if ctx.GlobalIsSet(utils.DataDirFlag.Name) {
		cfg.DataDir = ctx.GlobalString(utils.DataDirFlag.Name)
	}

	node_cfg.P2P.ListenAddr = fmt.Sprintf(":%d", 30303+1000*offset)
	if cfg.NodeType == "consensus" {
		node_cfg.P2P.StaticNodes, _ = genesisConfig.GetStaticNodes(offset)
		cfg.StaticNodes = node_cfg.P2P.StaticNodes
		cfg.TrustedNodes, _ = genesisConfig.GetStorageNode(offset, cfg.ConsensusIdx)
		if len(cfg.TrustedNodes) > 0 {
			node_cfg.P2P.StaticNodes = append(node_cfg.P2P.StaticNodes, cfg.TrustedNodes[0])
		}
	} else {
		/*
			if len(cfg.TrustedNode) > 0 {
				node_cfg.P2P.TrustedNodes, _ = genesisConfig.GetTrustedNodes(offset, cfg.TrustedNode)
				cfg.TrustedNodes = node_cfg.P2P.TrustedNodes
			}
		*/
		node_cfg.P2P.StaticNodes, _ = genesisConfig.GetStaticNode(offset, cfg.ConsensusIdx)
		cfg.StaticNodes = node_cfg.P2P.StaticNodes
		node_cfg.P2P.BootstrapNodes, _ = genesisConfig.GetStaticNodes(offset)
		cfg.TrustedNodes, _ = genesisConfig.GetStaticNode(offset, cfg.ConsensusIdx)
	}

	log.Info("main", "wolk config", cfg)
	log.Info("main", "node_cfg", node_cfg)

	stack, err := node.New(node_cfg)
	if err != nil {
		utils.Fatalf("Failed to create the protocol stack: %v", err)
	}


	err = stack.Register(func(ctx *node.ServiceContext) (svc node.Service, err error) {
		//wcs, err := wolk.NewWolk(ctx, cfg, stack.Server())
		wcs, err := wolk.NewWolk(ctx, cfg)
		if err != nil {
			log.Error("[main:makeFullNode] NewWolk", "err", err)
			return wcs, err
		}
		log.Info("[main:makeFullNode] NewWolk SUCCESS")
		return wcs, err
	})
	if err != nil {
		log.Error("[main:makeFullNode] stack.Register", "err", err)
		os.Exit(0)
	}
	return stack
}

// cloudstore is the main entry point into the system if no special subcommand is ran.
// It creates a default node based on the command line arguments and runs it in
// blocking mode, waiting for it to be shut down.
func cloud(ctx *cli.Context) error {
	node := makeFullNode(ctx)
	startNode(ctx, node)
	node.Wait()
	return nil
}

// startNode boots up the system node and all registered protocols, after which
// it unlocks any requested accounts, and starts the RPC/IPC interfaces and the
// miner.
func startNode(ctx *cli.Context, stack *node.Node) {
	// Start up the node itself
	log.Debug("startNode")
	utils.StartNode(stack)
	cfg := &(wolkcloud.DefaultConfig)
	cfg_file := wolkcloud.DefaultConfigWolkFile
	if file := ctx.GlobalString(configFileFlag.Name); file != "" {
		cfg_file = file
	}
	if _, err := os.Stat(cfg_file); !os.IsNotExist(err) {
		if err := wolkcloud.LoadConfig(cfg_file, cfg); err != nil {
			log.Error("CONFIG ERROR", "err", err)
		}
	}

	log.Info("stack.Service", "stack", stack, "stack.Service", stack.Service)
	var wolk *wolk.WolkStore
	if err := stack.Service(&wolk); err != nil {
		utils.Fatalf("Wolk service not running: %v", err)
	}
	log.Info("wolk", "wolk", wolk)
}
