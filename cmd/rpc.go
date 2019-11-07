package cmd

import (
	"fmt"

	"github.com/go-imsto/imsto/config"
	"github.com/go-imsto/imsto/rpc"
)

var cmdRPC = &Command{
	UsageLine: "rpc [-addr :8969] [-tls]",
	Short:     "serve RPC service",
	Long: `
serve RPC service
`,
}

var (
	rpcAddr string
	isTLS   bool
)

func init() {
	cmdRPC.Run = runRPC
	cmdRPC.Flag.StringVar(&rpcAddr, "addr", config.Current.RPCListen, "tcp listen address")
	cmdRPC.Flag.BoolVar(&isTLS, "tls", false, " use tls")
}

func runRPC(args []string) bool {
	fmt.Printf("Start RPC service %s at addr %s\n", config.Version, rpcAddr)
	s := rpc.NewServer(rpcAddr, isTLS)
	s.Serve()
	return true
}
