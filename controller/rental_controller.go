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

type IRentalController interface {
	FindAll(c *gin.Context)
	FinById(c *gin.Context)
	Insert(c *gin.Context)
	UpdateById(c *gin.Context)
	DeleteById(c *gin.Context)
	ReturnRental(c *gin.Context)
}

type RentalController struct {
	RentalSvc service.IRentalService
}

func NewRentalController(rentalSvc service.IRentalService) IRentalController {
	return &RentalController{
		RentalSvc: rentalSvc,
	}
}

// FindAll godoc
// @Summary Mendapatkan semua data rental
// @Tags Rental
// @Produce json
// @Param page query string false "Page"
// @Param limit query string false "Limit"
// @Success 200 {object} entity.Rental
// @Router /rental [get]
func (r *RentalController) FindAll(c *gin.Context) {
	var logger = helpers.Logger

	var page = c.DefaultQuery("page", "1")
	var pageInt = helpers.ParseToInt(page)

	var limit = c.DefaultQuery("limit", "10")
	var limitInt = helpers.ParseToInt(limit)

	var offset = (pageInt - 1) * limitInt

	data, totalData, err := r.RentalSvc.FindAll(c.Request.Context(), limitInt, offset)
	if err != nil {
		logger.Error("Failed to find all rentals: ", err)
		response.ResponseError(c, http.StatusInternalServerError, "Failed to find all rentals")
		return
	}

	metaData := response.Page{
		Limit:     limitInt,
		Total:     int(totalData),
		Page:      pageInt,
		TotalPage: int(math.Ceil(float64(totalData) / float64(limitInt))),
	}

	response.ResponseSuccess(c, http.StatusOK, data, metaData, "Success get all rentals")
}

// FinById godoc
// @Summary Mendapatkan data rental berdasarkan id
// @Tags Rental
// @Produce json
// @Param id path string true "Rental ID"
// @Success 200 {object} entity.Rental
// @Router /rental/{id} [get]
func (r *RentalController) FinById(c *gin.Context) {
	var logger = helpers.Logger

	var id = c.Param("id")
	if id == "" {
		logger.Error("Id is required")
		response.ResponseError(c, http.StatusBadRequest, "Id is required")
		return
	}

	data, err := r.RentalSvc.FindById(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error(fmt.Errorf("rental with id %s not found", id))
			response.ResponseError(c, http.StatusNotFound, "Rental not found")
			return
		}

		logger.Error(fmt.Errorf("failed to find rental by id %s: %v", id, err))
		response.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.ResponseSuccess(c, http.StatusOK, data, nil, "Success get rental")
}

// Insert godoc
// @Summary Insert rental baru
// @Tags Rental
// @Accept json
// @Produce json
// @Param rental body entity.CreateRentalRequest true "Rental"
// @Success 200 {object} entity.Rental
// @Router /rental [post]
func (r *RentalController) Insert(c *gin.Context) {
	var logger = helpers.Logger

	claims, exists := c.Get("claims")
	if !exists {
		logger.Error("Claims not found in context")
		response.ResponseError(c, http.StatusUnauthorized, "Claims not found in context")
		return
	}

	claimsData, ok := claims.(*helpers.ClaimsToken)
	if !ok {
		logger.Error("Invalid claims type")
		response.ResponseError(c, http.StatusUnauthorized, "Invalid claims type")
		return
	}

	var reqBody entity.CreateRentalRequest
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		logger.Error("Failed to bind JSON: ", err)
		response.ResponseError(c, http.StatusBadRequest, "Failed to bind JSON")
		return
	}

	reqBody.UserID = claimsData.UserID

	rental, err := r.RentalSvc.CreateRental(c.Request.Context(), reqBody)
	if err != nil {
		logger.Error("Failed to insert rental: ", err)
		response.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.ResponseSuccess(c, http.StatusOK, rental, nil, "Success insert rental")
}

// UpdateById godoc
// @Summary Perpanjang sewa rental
// @Tags Rental
// @Accept json
// @Produce json
// @Param id path string true "ID Rental"
// @Param request body entity.ExtendRentalRequest true "Data perpanjangan rental"
// @Security ApiCookieAuth
// @Router /rental/{id} [put]
func (r *RentalController) UpdateById(c *gin.Context) {
	var logger = helpers.Logger

	var id = c.Param("id")
	if id == "" {
		logger.Error("Id is required")
		response.ResponseError(c, http.StatusBadRequest, "Id is required")
		return
	}

	var request entity.ExtendRentalRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Failed to bind JSON: ", err)
		response.ResponseError(c, http.StatusBadRequest, err.Error())
		return
	}

	if request.NewExpectedReturnDate.IsZero() {
		logger.Error("Tanggal perpanjangan wajib diisi")
		response.ResponseError(c, http.StatusBadRequest, "Tanggal perpanjangan wajib diisi")
		return
	}

	rental, payment, err := r.RentalSvc.ExtendRental(c.Request.Context(), id, request)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "rental tidak ditemukan" {
			status = http.StatusNotFound
		}
		logger.Error("Failed to extend rental: ", err)
		response.ResponseError(c, status, err.Error())
		return
	}

	metadata := map[string]interface{}{
		"additional_cost":       payment.GrossAmount,
		"snap_token":            payment.SnapToken,
		"snap_url":              payment.SnapURL,
		"original_rental_price": rental.TotalRentalPrice - payment.GrossAmount,
		"new_total_price":       rental.TotalRentalPrice,
	}

	responseData := map[string]interface{}{
		"rental":  rental,
		"payment": payment,
	}

	response.ResponseSuccess(c, http.StatusOK, responseData, metadata, "Berhasil memperpanjang rental dan membuat pembayaran tambahan")
}

// DeleteById godoc
// @Summary Hapus rental
// @Tags Rental
// @Param id path string true "Rental ID"
// @Success 200 {object} response.APISuccessResponse
// @Router /rental/{id} [delete]
func (r *RentalController) DeleteById(c *gin.Context) {
	var logger = helpers.Logger

	var id = c.Param("id")
	if id == "" {
		logger.Error("Id is required")
		response.ResponseError(c, http.StatusBadRequest, "Id is required")
		return
	}

	err := r.RentalSvc.DeleteById(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error(fmt.Errorf("rental with id %s not found", id))
			response.ResponseError(c, http.StatusNotFound, "Rental not found")
			return
		}

		logger.Error(fmt.Errorf("failed to delete rental by id %s: %v", id, err))
		response.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.ResponseSuccess(c, http.StatusOK, nil, nil, "Success delete rental")
}

// Return godoc
// @Summary Pengembalian rental
// @Tags Rental
// @Accept json
// @Produce json
// @Param id path string true "Rental ID"
// @Param request body entity.ReturnRentalRequest true "Data pengembalian rental"
// @Success 200 {object} entity.Rental
// @Router /rental/{id}/return [put]
func (r *RentalController) ReturnRental(c *gin.Context) {
	idStr := c.Param("id")

	var request entity.ReturnRentalRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.ResponseError(c, http.StatusBadRequest, err.Error())
		return
	}

	rental, err := r.RentalSvc.ReturnRental(c.Request.Context(), idStr, request)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "rental tidak ditemukan" {
			status = http.StatusNotFound
		}
		response.ResponseError(c, status, err.Error())
		return
	}

	response.ResponseSuccess(c, http.StatusOK, rental, nil, "Success return")
}
