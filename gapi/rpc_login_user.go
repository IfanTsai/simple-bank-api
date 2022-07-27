package gapi

import (
	"context"
	"database/sql"

	db "github.com/ifantsai/simple-bank-api/db/sqlc"
	"github.com/ifantsai/simple-bank-api/pb"
	"github.com/ifantsai/simple-bank-api/util"
	"github.com/ifantsai/simple-bank-api/validator"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *GRPCServer) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	violations := validateLoginUserRequest(req)
	if len(violations) != 0 {
		return nil, invalidParameters(violations)
	}

	user, err := s.store.GetUser(ctx, req.GetUsername())
	if err != nil {
		errorCode := codes.Internal
		if errors.Is(errors.Cause(err), sql.ErrNoRows) {
			errorCode = codes.NotFound
		}

		return nil, status.Errorf(errorCode, "failed to get user, %s", err)
	}

	if err = util.CheckPassword(req.GetPassword(), user.HashedPassword); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to check password, %s", err)
	}

	// Note: only use username
	accessToken, accessTokenPayload, err := s.tokenMaker.CreateToken(0, user.Username, s.config.AccessTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create access token: %s", err)
	}

	refreshToken, refreshTokenPayload, err := s.tokenMaker.CreateToken(0, user.Username, s.config.RefreshTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create refresh token, %s", err)
	}

	metadata := s.extractMetadata(ctx)

	session, err := s.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshTokenPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    metadata.UserAgent,
		ClientIp:     metadata.ClientIP,
		IsBlocked:    false,
		ExpiresAt:    refreshTokenPayload.ExpiredAt,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create session, %s", err)
	}

	return &pb.LoginUserResponse{
		User:                  convertUser(&user),
		SessionId:             session.ID.String(),
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  timestamppb.New(accessTokenPayload.ExpiredAt),
		RefreshTokenExpiresAt: timestamppb.New(refreshTokenPayload.ExpiredAt),
	}, nil
}

func validateLoginUserRequest(req *pb.LoginUserRequest) []*BadRequestFieldViolation {
	var violations []*BadRequestFieldViolation

	if err := validator.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if err := validator.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}

	return violations
}
