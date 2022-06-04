package gapi

import (
	"context"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	grpcGatewayUserAgentKey = "grpcgateway-user-agent"
	userAgentKey            = "user-agent"
	xForwardedForKey        = "x-forwarded-for"
)

type Metadata struct {
	UserAgent string
	ClientIP  string
}

func (s *GRPCServer) extractMetadata(ctx context.Context) *Metadata {
	mtdt := &Metadata{}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if ua := md.Get(grpcGatewayUserAgentKey); len(ua) > 0 {
			mtdt.UserAgent = ua[0]
		}

		if ua := md.Get(userAgentKey); len(ua) > 0 {
			mtdt.UserAgent = ua[0]
		}

		if ip := md.Get(xForwardedForKey); len(ip) > 0 {
			mtdt.ClientIP = ip[0]
		}
	}

	if p, ok := peer.FromContext(ctx); ok {
		if ip := p.Addr.String(); len(ip) > 0 {
			mtdt.ClientIP = ip
		}
	}

	return mtdt
}
