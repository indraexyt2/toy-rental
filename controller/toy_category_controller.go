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

type IToyCategoryController interface {
	FindAll(c *gin.Context)
	FinById(c *gin.Context)
	Insert(c *gin.Context)
	UpdateById(c *gin.Context)
	DeleteById(c *gin.Context)
}

type ToyCategoryController struct {
	toyCategorySvc service.IToyCategoryService
}

func NewToyCategoryController(toyCategorySvc service.IToyCategoryService) IToyCategoryController {
	return &ToyCategoryController{
		toyCategorySvc: toyCategorySvc,
	}
}

// FindAll godoc
// @Tags Toy Category
// @Summary Mengambil semua data toy category
// @Produce json
// @Param page query string false "Page"
// @Param limit query string false "Limit"
// @Success 200 {object} entity.ToyCategory
// @Router /toy/category [get]
func (tc *ToyCategoryController) FindAll(c *gin.Context) {
	var logger = helpers.Logger

	var page = c.DefaultQuery("page", "1")
	var pageInt = helpers.ParseToInt(page)

	var limit = c.DefaultQuery("limit", "10")
	var limitInt = helpers.ParseToInt(limit)

	var offset = (pageInt - 1) * limitInt

	data, totalData, err := tc.toyCategorySvc.FindAll(c.Request.Context(), limitInt, offset)
	if err != nil {
		logger.Error("Failed to find all toy categories: ", err)
		response.ResponseError(c, http.StatusInternalServerError, "Failed to find all toy categories")
		return
	}

	metaData := response.Page{
		Limit:     limitInt,
		Total:     int(totalData),
		Page:      pageInt,
		TotalPage: int(math.Ceil(float64(totalData) / float64(limitInt))),
	}

	response.ResponseSuccess(c, http.StatusOK, data, metaData, "Success get all toy categories")
}

// @Summary Mengambil data toy category berdasarkan id
// @Tags Toy Category
// @Produce json
// @Param id path string true "Toy Category ID"
// @Success 200 {object} entity.ToyCategory
// @Router /toy/category/{id} [get]
func (tc *ToyCategoryController) FinById(c *gin.Context) {
	var logger = helpers.Logger

	var id = c.Param("id")
	if id == "" {
		logger.Error("Id is required")
		response.ResponseError(c, http.StatusBadRequest, "Id is required")
		return
	}

	data, err := tc.toyCategorySvc.FindById(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error(fmt.Errorf("toy category with id %s not found", id))
			response.ResponseError(c, http.StatusNotFound, "Toy category not found")
			return
		}

		logger.Error(fmt.Errorf("failed to find toy category by id %s: %v", id, err))
		response.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.ResponseSuccess(c, http.StatusOK, data, nil, "Success get toy category")
}

// Insert godoc
// @Summary Insert toy category
// @Tags Toy Category
// @Accept json
// @Produce json
// @Param toy_category body entity.ToyCategory true "Toy Category"
// @Success 200 {object} response.APISuccessResponse
// @Router /toy/category [post]
func (tc *ToyCategoryController) Insert(c *gin.Context) {
	var logger = helpers.Logger

	var reqBody entity.ToyCategory
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		logger.Error("Failed to bind JSON: ", err)
		response.ResponseError(c, http.StatusBadRequest, "Failed to bind JSON")
		return
	}

	if err := reqBody.Validate(); err != nil {
		logger.Error("Failed to validate toy category: ", err)
		response.ResponseError(c, http.StatusBadRequest, "Failed to validate toy category")
		return
	}

	err := tc.toyCategorySvc.Insert(c.Request.Context(), &reqBody)
	if err != nil {
		logger.Error("Failed to insert toy category: ", err)
		response.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.ResponseSuccess(c, http.StatusOK, nil, nil, "Success insert toy category")
}

// UpdateById godoc
// @Summary Update toy berdasarkan id
// @Tags Toy Category
// @Accept json
// @Produce json
// @Param id path string true "Toy Category ID"
// @Param toy_category body entity.ToyCategory true "Toy Category"
// @Success 200 {object} response.APISuccessResponse
// @Router /toy/category/{id} [put]
func (tc *ToyCategoryController) UpdateById(c *gin.Context) {
	var logger = helpers.Logger

	var id = c.Param("id")
	if id == "" {
		logger.Error("Id is required")
		response.ResponseError(c, http.StatusBadRequest, "Id is required")
		return
	}

	var reqBody entity.ToyCategory
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		logger.Error("Failed to bind JSON: ", err)
		response.ResponseError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := reqBody.Validate(); err != nil {
		logger.Error("Failed to validate toy category: ", err)
		response.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	err := tc.toyCategorySvc.UpdateById(c.Request.Context(), id, &reqBody)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error(fmt.Errorf("toy category with id %s not found", id))
			response.ResponseError(c, http.StatusNotFound, "Toy category not found")
			return
		}

		logger.Error(fmt.Errorf("failed to update toy category by id %s: %v", id, err))
		response.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.ResponseSuccess(c, http.StatusOK, nil, nil, "Success update toy category")
}

// DeleteById godoc
// @Summary Delete toy berdasarkan id
// @Tags Toy Category
// @Param id path string true "Toy Category ID"
// @Success 200 {object} response.APISuccessResponse
// @Router /toy/category/{id} [delete]
func (tc *ToyCategoryController) DeleteById(c *gin.Context) {
	var logger = helpers.Logger

	var id = c.Param("id")
	if id == "" {
		logger.Error("Id is required")
		response.ResponseError(c, http.StatusBadRequest, "Id is required")
		return
	}

	err := tc.toyCategorySvc.DeleteById(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error(fmt.Errorf("toy category with id %s not found", id))
			response.ResponseError(c, http.StatusNotFound, "Toy category not found")
			return
		}

		logger.Error(fmt.Errorf("failed to delete toy category by id %s: %v", id, err))
		response.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.ResponseSuccess(c, http.StatusOK, nil, nil, "Success delete toy category")
}
