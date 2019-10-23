package rpc

import (
	"github.com/getsentry/raven-go"
)

var (
	packagePrefixes = []string{"github.com/go-imsto"}
)

func reportError(err error, tags map[string]string) {
	var packet *raven.Packet
	packet = raven.NewPacket(err.Error(),
		raven.NewException(err, raven.NewStacktrace(1, 3, packagePrefixes)))

	raven.Capture(packet, tags)
}
