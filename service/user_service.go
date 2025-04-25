package service

import (
	"context"
	"errors"
	"final-project/entity"
	"final-project/repository"
	"final-project/utils/helpers"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type IUserService interface {
	IBaseService[entity.User]
	Login(ctx context.Context, emailOrUsername string, password string) (entity.User, entity.UserToken, error)
}

type UserService struct {
	BaseService[entity.User]
	UserRepository      repository.IUserRepository
	UserTokenRepository repository.IUserTokenRepository
	JwtHelper           helpers.JWTHelper
}

func NewUserService(
	userRepo repository.IUserRepository,
	userTokenRepo repository.IUserTokenRepository,
	jwtHelper helpers.JWTHelper,
) IUserService {
	return &UserService{
		BaseService:         BaseService[entity.User]{repository: userRepo},
		UserRepository:      userRepo,
		UserTokenRepository: userTokenRepo,
		JwtHelper:           jwtHelper,
	}
}

func (s *UserService) Insert(ctx context.Context, entity *entity.User) error {
	userData, err := s.UserRepository.FindByEmailOrUsername(ctx, entity.Email)
	if err == nil && userData != nil {
		return errors.New("Email sudah terdaftar")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	userData, err = s.UserRepository.FindByEmailOrUsername(ctx, entity.Username)
	if err == nil && userData != nil {
		return errors.New("Username sudah terdaftar")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(entity.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	entity.Password = string(hashedPassword)
	return s.repository.Insert(ctx, entity)
}

func (s *UserService) Login(ctx context.Context, emailOrUsername string, password string) (entity.User, entity.UserToken, error) {
	user, err := s.UserRepository.FindByEmailOrUsername(ctx, emailOrUsername)
	if err != nil {
		return entity.User{}, entity.UserToken{}, errors.New("Username atau email tidak ditemukan")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return entity.User{}, entity.UserToken{}, errors.New("Password salah")
	}

	accessToken, accessTokenExp, _ := s.JwtHelper.GenerateAccessToken(user.ID, user.Email, user.Role)
	refreshToken, refreshTokenExp, _ := s.JwtHelper.GenerateRefreshToken(user.ID)

	userToken := &entity.UserToken{
		UserID:                user.ID,
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessTokenExp,
		RefreshTokenExpiresAt: refreshTokenExp,
	}

	err = s.UserTokenRepository.Insert(ctx, userToken)
	if err != nil {
		return entity.User{}, entity.UserToken{}, err
	}

	return *user, *userToken, nil
}
