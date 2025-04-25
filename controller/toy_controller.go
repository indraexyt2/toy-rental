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

type IToyController interface {
	FindAll(c *gin.Context)
	FinById(c *gin.Context)
	Insert(c *gin.Context)
	UpdateById(c *gin.Context)
	DeleteById(c *gin.Context)
}

type ToyController struct {
	toySvc service.IToyService
}

func NewToyController(toySvc service.IToyService) IToyController {
	return &ToyController{
		toySvc: toySvc,
	}
}

// FindAll godoc
// @Summary Mengambil semua data toy
// @Tags Toy
// @Produce json
// @Param page query string false "Page"
// @Param limit query string false "Limit"
// @Success 200 {object} entity.Toy
// @Router /toy [get]
func (t ToyController) FindAll(c *gin.Context) {
	var logger = helpers.Logger

	var page = c.DefaultQuery("page", "1")
	var pageInt = helpers.ParseToInt(page)

	var limit = c.DefaultQuery("limit", "10")
	var limitInt = helpers.ParseToInt(limit)

	var offset = (pageInt - 1) * limitInt

	data, totalData, err := t.toySvc.FindAll(c.Request.Context(), limitInt, offset)
	if err != nil {
		logger.Error("Failed to find all toys: ", err)
		response.ResponseError(c, http.StatusInternalServerError, "Failed to find all toys")
		return
	}

	metaData := response.Page{
		Limit:     limitInt,
		Total:     int(totalData),
		Page:      pageInt,
		TotalPage: int(math.Ceil(float64(totalData) / float64(limitInt))),
	}

	response.ResponseSuccess(c, http.StatusOK, data, metaData, "Success to find all toys")
}

// FindById godoc
// @Summary Mengambil data toy berdasarkan id
// @Tags Toy
// @Produce json
// @Param id path string true "Toy ID"
// @Success 200 {object} entity.Toy
// @Router /toy/{id} [get]
func (t ToyController) FinById(c *gin.Context) {
	var logger = helpers.Logger

	var id = c.Param("id")
	if id == "" {
		logger.Error("Id is required")
		response.ResponseError(c, http.StatusBadRequest, "Id is required")
		return
	}

	data, err := t.toySvc.FindById(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error(fmt.Errorf("toy with id %s not found", id))
			response.ResponseError(c, http.StatusNotFound, "Toy not found")
			return
		}

		logger.Error(fmt.Errorf("failed to find toy by id %s: %v", id, err))
		response.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.ResponseSuccess(c, http.StatusOK, data, nil, "Success to find toy")
}

// Insert godoc
// @Summary Menambahkan mainan baru
// @Tags Toy
// @Accept json
// @Produce json
// @Param toy body entity.ToyRequest true "Data mainan baru"
// @Success 200 {object} entity.Toy
// @Router /toy [post]
func (t ToyController) Insert(c *gin.Context) {
	var logger = helpers.Logger

	var toyRequest entity.ToyRequest
	if err := c.ShouldBindJSON(&toyRequest); err != nil {
		logger.Error("Gagal binding request: ", err)
		response.ResponseError(c, http.StatusBadRequest, "Format data tidak valid")
		return
	}

	toy, err := t.toySvc.CreateToy(c.Request.Context(), toyRequest)
	if err != nil {
		logger.Error("Gagal membuat mainan: ", err)
		response.ResponseError(c, http.StatusBadRequest, err.Error())
		return
	}

	response.ResponseSuccess(c, http.StatusOK, toy, nil, "Berhasil menambahkan mainan baru")
}

// UpdateById godoc
// @Summary Update mainan berdasarkan id
// @Tags Toy
// @Accept json
// @Produce json
// @Param id path string true "Toy ID"
// @Param toy body entity.ToyUpdateRequest true "Data mainan yang diperbarui"
// @Success 200 {object} entity.Toy "Updated Toy"
// @Router /toy/{id} [put]
func (t ToyController) UpdateById(c *gin.Context) {
	var logger = helpers.Logger

	id := c.Param("id")
	if id == "" {
		logger.Error("ID wajib diisi")
		response.ResponseError(c, http.StatusBadRequest, "ID wajib diisi")
		return
	}

	// Binding request
	var updateRequest entity.ToyUpdateRequest
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		logger.Error("Gagal binding request: ", err)
		response.ResponseError(c, http.StatusBadRequest, "Format data tidak valid")
		return
	}

	// Delegasikan ke service
	updatedToy, err := t.toySvc.UpdateToy(c.Request.Context(), id, updateRequest)
	if err != nil {
		logger.Error("Gagal memperbarui mainan: ", err)
		response.ResponseError(c, http.StatusBadRequest, err.Error())
		return
	}

	response.ResponseSuccess(c, http.StatusOK, updatedToy, nil, "Berhasil memperbarui mainan")
}

// DeleteById godoc
// @Summary Menghapus mainan berdasarkan id
// @Tags Toy
// @Produce json
// @Param id path string true "Toy Image ID"
// @Success 200 {object} entity.Toy
// @Router /toy/{id} [delete]
func (t ToyController) DeleteById(c *gin.Context) {
	var logger = helpers.Logger

	var id = c.Param("id")
	if id == "" {
		logger.Error("Id is required")
		response.ResponseError(c, http.StatusBadRequest, "Id is required")
		return
	}

	err := t.toySvc.DeleteById(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error(fmt.Errorf("toy with id %s not found", id))
			response.ResponseError(c, http.StatusNotFound, "Toy not found")
			return
		}

		logger.Error(fmt.Errorf("failed to delete toy by id %s: %v", id, err))
		response.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.ResponseSuccess(c, http.StatusOK, nil, nil, "Success delete toy")
}
