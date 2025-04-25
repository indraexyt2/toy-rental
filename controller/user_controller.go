package controller

import (
	"errors"
	"final-project/entity"
	"final-project/service"
	"final-project/utils/helpers"
	"final-project/utils/response"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"math"
	"net/http"
)

type IUserController interface {
	FindAll(c *gin.Context)
	FinById(c *gin.Context)
	Insert(c *gin.Context)
	UpdateById(c *gin.Context)
	DeleteById(c *gin.Context)
	Login(c *gin.Context)
	Logout(c *gin.Context)
	Me(c *gin.Context)
}

type UserController struct {
	userService      service.IUserService
	userTokenService service.ITokenService
}

func NewUserController(
	userSvc service.IUserService,
	userTokenSvc service.ITokenService,
) IUserController {
	return &UserController{
		userService:      userSvc,
		userTokenService: userTokenSvc,
	}
}

// GetUsers godoc
// @Summary      List users
// @Description  Get list of all users
// @Tags         users
// @Security ApiCookieAuth
// @Produce      json
// @Param        page   query string  false  "Page"
// @Param        limit  query string  false  "Limit"
// @Success      200  {array}  entity.User
// @Router       /admin/users [get]
func (uc *UserController) FindAll(c *gin.Context) {
	var log = helpers.Logger

	var page = c.DefaultQuery("page", "1")
	var pageInt = helpers.ParseToInt(page)

	var limit = c.DefaultQuery("limit", "10")
	var limitInt = helpers.ParseToInt(limit)

	var offset = (pageInt - 1) * limitInt

	data, totalData, err := uc.userService.FindAll(c.Request.Context(), limitInt, offset)
	if err != nil {
		log.Error("Failed to find all users: ", err)
		response.ResponseError(c, http.StatusInternalServerError, "Failed to find all users")
		return
	}

	metaData := response.Page{
		Limit:     limitInt,
		Total:     int(totalData),
		Page:      pageInt,
		TotalPage: int(math.Ceil(float64(totalData) / float64(limitInt))),
	}

	response.ResponseSuccess(c, http.StatusOK, data, metaData, "Success to find all users")
}

// GetUserById godoc
// @Summary      Mengambil data user berdasarkan id
// @Tags         users
// @Security ApiCookieAuth
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  entity.User
// @Router       /admin/user/{id} [get]
func (uc *UserController) FinById(c *gin.Context) {
	var log = helpers.Logger

	var id = c.Param("id")
	if id == "" {
		log.Error("Id is required")
		response.ResponseError(c, http.StatusBadRequest, "Id is required")
		return
	}

	data, err := uc.userService.FindById(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(fmt.Errorf("user with id %s not found", id))
			response.ResponseError(c, http.StatusNotFound, "User not found")
			return
		}

		log.Error(fmt.Errorf("failed to find user by id %s: %v", id, err))
		response.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.ResponseSuccess(c, http.StatusOK, data, nil, "Success to find user by id")
}

// CreateUser godoc
// @Summary      Membuat user baru
// @Tags         users
// @Produce      json
// @Param        user  body      entity.User  true  "User"
// @Success      200  {object}  entity.User
// @Router       /user/auth/register [post]
func (uc *UserController) Insert(c *gin.Context) {
	var log = helpers.Logger

	var user entity.User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Error("Failed to bind JSON: ", err)
		response.ResponseError(c, http.StatusBadRequest, "Failed to bind JSON")
		return
	}

	if err := user.Validate(true); err != nil {
		log.Error("Failed to validate user: ", err)
		response.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	err := uc.userService.Insert(c.Request.Context(), &user)
	if err != nil {
		log.Error("Failed to insert user: ", err)
		response.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	user.Password = ""

	response.ResponseSuccess(c, http.StatusOK, user, nil, "Success to insert user")
}

// UpdateUser godoc
// @Summary      Update user berdasarkan id
// @Tags         users
// @Security ApiCookieAuth
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Param        user  body     entity.User  true  "User"
// @Success      200  {object}  entity.User
// @Router       /user/auth/{id} [put]
func (uc *UserController) UpdateById(c *gin.Context) {
	var log = helpers.Logger

	claims, exists := c.Get("claims")
	if !exists {
		log.Error("Claims not found in context")
		response.ResponseError(c, http.StatusUnauthorized, "Claims not found in context")
		return
	}

	claimsData, ok := claims.(*helpers.ClaimsToken)
	if !ok {
		log.Error("Invalid claims type")
		response.ResponseError(c, http.StatusUnauthorized, "Invalid claims type")
		return
	}

	var id = c.Param("id")
	if id == "" {
		log.Error("Id is required")
		response.ResponseError(c, http.StatusBadRequest, "Id is required")
		return
	}

	if id != claimsData.UserID.String() {
		log.Error("User ID does not match")
		response.ResponseError(c, http.StatusUnauthorized, "User ID does not match")
		return
	}

	var user entity.User
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Error("Failed to bind JSON: ", err)
		response.ResponseError(c, http.StatusBadRequest, "Failed to bind JSON")
		return
	}

	if err := user.Validate(false); err != nil {
		log.Error("Failed to validate user: ", err)
		response.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	err := uc.userService.UpdateById(c.Request.Context(), id, &user)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(fmt.Errorf("user with id %s not found", id))
			response.ResponseError(c, http.StatusNotFound, "User not found")
			return
		}

		log.Error(fmt.Errorf("failed to update user by id %s: %v", id, err))
		response.ResponseError(c, http.StatusInternalServerError, err.Error())
	}

	user.Password = ""

	response.ResponseSuccess(c, http.StatusOK, user, nil, "Success to update user by id")
}

// DeleteUser godoc
// @Summary      Delete user berdasarkan id
// @Tags         users
// @Security ApiCookieAuth
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  entity.User
// @Router       /user/auth/{id} [delete]
func (uc *UserController) DeleteById(c *gin.Context) {
	var log = helpers.Logger

	claims, exists := c.Get("claims")
	if !exists {
		log.Error("Claims not found in context")
		response.ResponseError(c, http.StatusUnauthorized, "Claims not found in context")
		return
	}

	claimsData, ok := claims.(*helpers.ClaimsToken)
	if !ok {
		log.Error("Invalid claims type")
		response.ResponseError(c, http.StatusUnauthorized, "Invalid claims type")
		return
	}

	var id = c.Param("id")
	if id == "" {
		log.Error("Id is required")
		response.ResponseError(c, http.StatusBadRequest, "Id is required")
		return
	}

	if id != claimsData.UserID.String() {
		log.Error("User ID does not match")
		response.ResponseError(c, http.StatusUnauthorized, "User ID does not match")
		return
	}

	err := uc.userService.DeleteById(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(fmt.Errorf("user with id %s not found", id))
			response.ResponseError(c, http.StatusNotFound, "User not found")
			return
		}

		log.Error(fmt.Errorf("failed to delete user by id %s: %v", id, err))
		response.ResponseError(c, http.StatusInternalServerError, "Failed to delete user by id")
	}

	response.ResponseSuccess(c, http.StatusOK, nil, nil, "Success to delete user by id")
}

// Login godoc
// @Summary      Login
// @Tags         users
// @Produce      json
// @Param        user  body      entity.UserLoginRequest  true  "User"
// @Success      200  {object}  entity.User
// @Router       /user/auth/login [post]
func (uc *UserController) Login(c *gin.Context) {
	var log = helpers.Logger

	var userLoginRequest entity.UserLoginRequest
	if err := c.ShouldBindJSON(&userLoginRequest); err != nil {
		log.Error("Failed to bind JSON: ", err)
		response.ResponseError(c, http.StatusBadRequest, "Failed to bind JSON")
		return
	}

	user, userToken, err := uc.userService.Login(c.Request.Context(), userLoginRequest.Email, userLoginRequest.Password)
	if err != nil {
		log.Error("Failed to login: ", err)
		response.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.SetCookie("access_token", userToken.AccessToken, 24*60*60, "/", "localhost", false, true)
	user.Password = ""

	response.ResponseSuccess(c, http.StatusOK, user, nil, "Success to login")
}

// Logout godoc
// @Summary      Logout
// @Security ApiCookieAuth
// @Tags         users
// @Produce      json
// @Success      200  {object}  response.APISuccessResponse
// @Router       /user/auth/logout [delete]
func (uc *UserController) Logout(c *gin.Context) {
	var log = helpers.Logger

	accessToken, exists := c.Get("access_token")
	if !exists {
		log.Error("Access token not found")
		response.ResponseError(c, http.StatusBadRequest, "Access token not found")
		return
	}

	err := uc.userTokenService.DeleteByAccessToken(c.Request.Context(), accessToken.(string))
	if err != nil {
		log.Error("Failed to delete user token: ", err)
		response.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.SetCookie("access_token", "", -1, "/", "localhost", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "localhost", false, true)

	response.ResponseSuccess(c, http.StatusOK, nil, nil, "Success to logout")
}

// GetUserById godoc
// @Summary      Get user berdasarkan token
// @Tags         users
// @Security ApiCookieAuth
// @Produce      json
// @Success      200  {object}  entity.User
// @Router       /user/auth/me [get]
func (uc *UserController) Me(c *gin.Context) {
	var log = helpers.Logger

	var claims, exists = c.Get("claims")
	if !exists {
		log.Error("Claims not found in context")
		response.ResponseError(c, http.StatusUnauthorized, "Claims not found in context")
		return
	}

	claimsData := claims.(*helpers.ClaimsToken)

	data, err := uc.userService.FindById(c.Request.Context(), claimsData.UserID.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error(fmt.Errorf("user with id %s not found", claimsData.ID))
			response.ResponseError(c, http.StatusNotFound, "User not found")
			return
		}

		log.Error(fmt.Errorf("failed to find user by id %s: %v", claimsData.ID, err))
		response.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.ResponseSuccess(c, http.StatusOK, data, nil, "Success to find user by id")
}
