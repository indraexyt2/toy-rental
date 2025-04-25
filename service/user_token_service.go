package service

import (
	"context"
	"final-project/entity"
	"final-project/repository"
	"final-project/utils/helpers"
)

type ITokenService interface {
	IBaseService[entity.UserToken]
	FindByAccessToken(ctx context.Context, accessToken string) (entity.UserToken, error)
	DeleteByAccessToken(ctx context.Context, accessToken string) error
	RefreshToken(ctx context.Context, accessToken string, claimsToken helpers.ClaimsToken) (entity.UserToken, error)
}

type TokenService struct {
	BaseService[entity.UserToken]
	userTokenRepository repository.IUserTokenRepository
	jwtHelper           helpers.JWTHelper
}

func NewTokenService(
	tokenRepo repository.IUserTokenRepository,
	jwtHelper helpers.JWTHelper,
) ITokenService {
	return &TokenService{
		BaseService:         BaseService[entity.UserToken]{repository: tokenRepo},
		userTokenRepository: tokenRepo,
		jwtHelper:           jwtHelper,
	}
}

func (s *TokenService) FindByAccessToken(ctx context.Context, accessToken string) (entity.UserToken, error) {
	return s.FindByAccessToken(ctx, accessToken)
}

func (s *TokenService) DeleteByAccessToken(ctx context.Context, accessToken string) error {
	return s.userTokenRepository.DeleteByAccessToken(ctx, accessToken)
}

func (s *TokenService) RefreshToken(ctx context.Context, accessToken string, claimsToken helpers.ClaimsToken) (entity.UserToken, error) {
	userToken, err := s.userTokenRepository.FindByAccessToken(ctx, accessToken)
	if err != nil {
		return entity.UserToken{}, err
	}

	newAccessToken, accessTokenExp, err := s.jwtHelper.GenerateAccessToken(userToken.UserID, claimsToken.Email, claimsToken.Role)
	if err != nil {
		return entity.UserToken{}, err
	}

	refreshToken, refreshTokenExp, err := s.jwtHelper.GenerateRefreshToken(userToken.UserID)
	if err != nil {
		return entity.UserToken{}, err
	}

	newUserToken := &entity.UserToken{
		UserID:                userToken.UserID,
		AccessToken:           newAccessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessTokenExp,
		RefreshTokenExpiresAt: refreshTokenExp,
	}

	err = s.userTokenRepository.UpdateByRefreshToken(ctx, userToken.RefreshToken, newUserToken)
	if err != nil {
		return entity.UserToken{}, err
	}

	return *newUserToken, nil
}
