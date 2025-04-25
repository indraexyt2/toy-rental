package middleware

import (
	"errors"
	"final-project/entity"
	"final-project/service"
	"final-project/utils/helpers"
	"final-project/utils/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
)

type AuthMiddleware struct {
	jwtHelper    helpers.JWTHelper
	userTokenSvc service.ITokenService
}

func NewAuthMiddleware(jwtHelper helpers.JWTHelper, userTokenSvc service.ITokenService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtHelper:    jwtHelper,
		userTokenSvc: userTokenSvc,
	}
}

func (m *AuthMiddleware) AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var log = helpers.Logger
		accessToken, err := c.Request.Cookie("access_token")
		if err != nil {
			log.Error("Failed to get access token: ", err)
			response.ResponseError(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		claims, err := m.jwtHelper.ValidateAccessToken(accessToken.Value)
		if err != nil {
			log.Error("Failed to validate access token: ", err)
			response.ResponseError(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		if claims.Role != entity.RoleAdmin {
			log.Error("Unauthorized")
			response.ResponseError(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		if errors.Is(err, jwt.ErrTokenExpired) {
			tokenClaims, err := m.jwtHelper.ExtractTokenClaims(accessToken.Value)
			if err != nil {
				log.Error("Failed to extract token claims: ", err)
				response.ResponseError(c, http.StatusUnauthorized, "Unauthorized")
				c.Abort()
				return
			}

			userToken, err := m.userTokenSvc.RefreshToken(c.Request.Context(), accessToken.Value, *tokenClaims)
			if err != nil {
				log.Error("Failed to refresh token: ", err)
				response.ResponseError(c, http.StatusUnauthorized, "Unauthorized")
				c.Abort()
				return
			}

			c.SetCookie("access_token", userToken.AccessToken, 24*60*60, "/", "localhost", false, true)
			claims, _ = m.jwtHelper.ValidateAccessToken(userToken.AccessToken)

			c.Set("claims", claims)
			c.Set("access_token", userToken.AccessToken)

			c.Next()
			return
		}

		c.Set("claims", claims)
		c.Set("access_token", accessToken.Value)

		c.Next()
	}
}

func (m *AuthMiddleware) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var log = helpers.Logger
		accessToken, err := c.Request.Cookie("access_token")
		if err != nil {
			log.Error("Failed to get access token: ", err)
			response.ResponseError(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		claims, err := m.jwtHelper.ValidateAccessToken(accessToken.Value)
		if err != nil {
			log.Error("Failed to validate access token: ", err)
			response.ResponseError(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		if errors.Is(err, jwt.ErrTokenExpired) {
			tokenClaims, err := m.jwtHelper.ExtractTokenClaims(accessToken.Value)
			if err != nil {
				log.Error("Failed to extract token claims: ", err)
				response.ResponseError(c, http.StatusUnauthorized, "Unauthorized")
				c.Abort()
				return
			}

			userToken, err := m.userTokenSvc.RefreshToken(c.Request.Context(), accessToken.Value, *tokenClaims)
			if err != nil {
				log.Error("Failed to refresh token: ", err)
				response.ResponseError(c, http.StatusUnauthorized, "Unauthorized")
				c.Abort()
				return
			}

			c.SetCookie("access_token", userToken.AccessToken, 24*60*60, "/", "localhost", false, true)
			claims, _ = m.jwtHelper.ValidateAccessToken(userToken.AccessToken)

			c.Set("claims", claims)
			c.Set("access_token", userToken.AccessToken)

			c.Next()
			return
		}

		c.Set("claims", claims)
		c.Set("access_token", accessToken.Value)

		c.Next()
	}
}
