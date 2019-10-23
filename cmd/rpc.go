package cmd

import (
	"github.com/go-imsto/imsto/rpc"
)

var cmdRPC = &Command{
	UsageLine: "rpc -addr :8969 -whiteList=\"127.0.0.1\"",
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
	s := rpc.NewServer(*rpcAddr, *isTLS)
	s.Serve()
	return true
}
