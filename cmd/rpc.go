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
	rpcAddr = cmdRPC.Flag.String("addr", ":8969", "tcp listen address")
	isTLS   = cmdRPC.Flag.Bool("tls", false, " use tls")
)

func init() {
	cmdRPC.Run = runRPC
}

func runRPC(args []string) bool {
	fmt.Printf("Start RPC service %s at addr %s\n", config.Version, *rpcAddr)
	s := rpc.NewServer(*rpcAddr, *isTLS)
	s.Serve()
	return true
}
