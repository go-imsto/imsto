package rpc

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"os"

	"github.com/getsentry/raven-go"
	"github.com/soheilhy/cmux"
	"golang.org/x/sync/errgroup"
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

// RPC ...
type RPC interface {
	Serve()
	Stop()
}

// server ...
type server struct {
	Addr string
	TLS  bool
	rs   *grpc.Server
}

// NewServer ...
func NewServer(addr string, isTLS bool) RPC {
	s := &server{
		Addr: addr,
		TLS:  isTLS,
	}

	var tags = map[string]string{"service": "imsto-rpc", "ver": config.Version}
	raven.SetTagsContext(tags)

	return s
}

func (s *server) Serve() {

	logger().Infow("listen", "addr", s.Addr)
	lis, err := net.Listen("tcp", s.Addr)
	if err != nil {
		logger().Fatalw("error listen", "addr", s.Addr)
	}
	s.serve(lis)
}

func (s *server) serve(lis net.Listener) error {
	cm := cmux.New(lis)
	grpcL := cm.Match(cmux.HTTP2())
	httpL := cm.Match(cmux.HTTP1Fast())

	g := new(errgroup.Group)
	g.Go(func() error { return s.grpcServe(grpcL) })
	g.Go(func() error { return s.httpServe(httpL) })
	g.Go(func() error { return cm.Serve() })

	logger().Infow("start cmux wait")
	return g.Wait()
}

func (s *server) grpcServe(l net.Listener) error {
	var opts []grpc.ServerOption
	opts = append(opts, grpc.MaxRecvMsgSize(int(defaultMaxMessageSize)))
	if s.TLS {
		opts = append(opts, grpc.Creds(tlsTc()))
	}

	s.rs = grpc.NewServer(opts...)

	pb.RegisterImageSvcServer(s.rs, &rpcImage{})
	healthgrpc.RegisterHealthServer(s.rs, health.NewServer())

	logger().Infow("start grpcServe", "grpc.ver", grpc.Version)
	return s.rs.Serve(l)
}

func (s *server) httpServe(l net.Listener) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	hs := &http.Server{Handler: mux}
	return hs.Serve(l)
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
