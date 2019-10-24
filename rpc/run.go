package rpc

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/getsentry/raven-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"

	pb "github.com/go-imsto/imsto-client/impb"
	"github.com/go-imsto/imsto/config"
)

const (
	defaultMaxMessageSize = 1 << 19 // 512K
)

func tlsTc() credentials.TransportCredentials {
	var (
		caCrt, serverCrt, serverKey string
	)
	caCrt = envOr("IMSTO_GRPC_CA", "certs/ca.crt")
	serverCrt = envOr("IMSTO_GRPC_SERVER_CRT", "certs/server.crt")
	serverKey = envOr("IMSTO_GRPC_SERVER_KEY", "certs/server.key")

	cert, err := tls.LoadX509KeyPair(serverCrt, serverKey)
	if err != nil {
		logger().Fatalw("tls.LoadX509KeyPair", "err", err)
	}

	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caCrt)
	if err != nil {
		logger().Fatalw("ioutil.ReadFile", "err", err)
	}

	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		logger().Fatalw("certPool.AppendCertsFromPEM err")
	}

	c := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	})
	return c
}

func decRPCServer(s *grpc.Server) *grpc.Server {
	pb.RegisterImageSvcServer(s, &rpcImage{})
	healthgrpc.RegisterHealthServer(s, health.NewServer())
	return s
}

// server ...
type server struct {
	Addr string
	TLS  bool
	rs   *grpc.Server
}

// NewServer ...
func NewServer(addr string, isTLS bool) *server {
	s := &server{
		Addr: addr,
		TLS:  isTLS,
	}

	var tags = map[string]string{"service": "imsto-rpc", "ver": config.Version}
	raven.SetTagsContext(tags)

	var opts []grpc.ServerOption
	opts = append(opts, grpc.MaxRecvMsgSize(int(defaultMaxMessageSize)))
	if s.TLS {
		opts = append(opts, grpc.Creds(tlsTc()))
	}

	s.rs = decRPCServer(grpc.NewServer(opts...))
	return s
}

func (s *server) Serve() {

	logger().Infow("listen", "addr", s.Addr)
	lis, err := net.Listen("tcp", s.Addr)
	if err != nil {
		logger().Fatalw("error listen", "addr", s.Addr)
	}
	if err := http.Serve(lis, s); err != nil {
		logger().Fatalw("error in Serve", "err", err)
	}

}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
		s.rs.ServeHTTP(w, r)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}

}

func (s *server) Stop() {
	if s.rs != nil {
		logger().Infow("stopping rpc")
		s.rs.GracefulStop()
	}
}

func envOr(key, dft string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return dft
}
