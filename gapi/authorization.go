package gapi

import (
	"context"
	"strings"

	"github.com/IfanTsai/go-lib/user/token"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

const (
	authorizationKey    = "authorization"
	authorizationBearer = "bearer"
)

func (s *GRPCServer) authorizeUser(ctx context.Context) (*token.Payload, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("no metadata")
	}

	auths, ok := md[authorizationKey]
	if !ok {
		return nil, errors.New("missing authorization")
	}

	if len(auths) == 0 {
		return nil, errors.New("authorization is empty")
	}

	if len(auths) > 1 {
		return nil, errors.New("multiple authorization")
	}

	authType, token, ok := strings.Cut(auths[0], " ")
	if !ok {
		return nil, errors.New("invalid authorization")
	}

	if strings.ToLower(authType) != authorizationBearer {
		return nil, errors.Errorf("not support authorization type: %s", authType)
	}

	payload, err := s.tokenMaker.VerifyToken(token)
	if err != nil {
		return nil, errors.Wrap(err, "invalid token")
	}

	return payload, nil
}
