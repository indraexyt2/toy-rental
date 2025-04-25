package controller

import (
	"final-project/entity"
	"final-project/service"
	"final-project/utils/helpers"
	"final-project/utils/response"
	"github.com/gin-gonic/gin"
	"net/http"
)

type IPaymentController interface {
	CreatePayment(c *gin.Context)
	GetPaymentByID(c *gin.Context)
	GetPaymentsByRentalID(c *gin.Context)
	HandlePaymentCallback(c *gin.Context)
}

type PaymentController struct {
	paymentSvc service.IPaymentService
}

func NewPaymentController(paymentSvc service.IPaymentService) IPaymentController {
	return &PaymentController{
		paymentSvc: paymentSvc,
	}
}

// CreatePayment godoc
// @Summary Membuat pembayaran rental
// @Description Buat pembayaran baru menggunakan midtrans
// @Tags Payment
// @Accept json
// @Produce json
// @Param rental_id body entity.CreatePaymentRequest true "ID Rental"
// @Success 200 {object} entity.Payment
// @Router /payment [post]
func (p *PaymentController) CreatePayment(c *gin.Context) {
	var logger = helpers.Logger
	var request entity.CreatePaymentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Error("Failed to bind JSON: ", err)
		response.ResponseError(c, http.StatusBadRequest, "Format data tidak valid")
		return
	}

	payment, err := p.paymentSvc.CreatePaymentForRental(c.Request.Context(), request.RentalID)
	if err != nil {
		logger.Error("Failed to create payment: ", err)
		response.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.ResponseSuccess(c, http.StatusOK, payment, nil, "Pembayaran berhasil dibuat")
}

// GetPaymentByID godoc
// @Summary Mendapatkan detail pembayaran
// @Description Mendapatkan detail pembayaran berdasarkan ID
// @Tags Payment
// @Produce json
// @Param id path string true "ID Pembayaran"
// @Success 200 {object} entity.Payment
// @Router /payment/{id} [get]
func (p *PaymentController) GetPaymentByID(c *gin.Context) {
	var logger = helpers.Logger

	id := c.Param("id")
	if id == "" {
		logger.Error("ID is required")
		response.ResponseError(c, http.StatusBadRequest, "ID wajib diisi")
		return
	}

	payment, err := p.paymentSvc.FindById(c.Request.Context(), id)
	if err != nil {
		logger.Error("Failed to get payment: ", err)
		response.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.ResponseSuccess(c, http.StatusOK, payment, nil, "Berhasil mendapatkan data pembayaran")
}

// GetPaymentsByRentalID godoc
// @Summary Mendapatkan semua pembayaran untuk rental
// @Description Mendapatkan semua pembayaran berdasarkan ID rental
// @Tags Payment
// @Produce json
// @Param rental_id path string true "ID Rental"
// @Success 200 {array} entity.Payment
// @Router /payment/rental/{rental_id} [get]
func (p *PaymentController) GetPaymentsByRentalID(c *gin.Context) {
	var logger = helpers.Logger

	rentalID := c.Param("rental_id")
	if rentalID == "" {
		logger.Error("Rental ID is required")
		response.ResponseError(c, http.StatusBadRequest, "ID rental wajib diisi")
		return
	}

	payments, err := p.paymentSvc.FindByRentalID(c.Request.Context(), rentalID)
	if err != nil {
		logger.Error("Failed to get payments: ", err)
		response.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.ResponseSuccess(c, http.StatusOK, payments, nil, "Berhasil mendapatkan data pembayaran")
}

// HandlePaymentCallback godoc
// @Summary Menangani callback dari Midtrans
// @Description Endpoint untuk menerima notifikasi dari Midtrans
// @Tags Payment
// @Accept json
// @Produce json
// @Success 200 {object} response.APISuccessResponse
// @Router /payment/callback [post]
func (p *PaymentController) HandlePaymentCallback(c *gin.Context) {
	var logger = helpers.Logger
	var notification map[string]interface{}
	if err := c.ShouldBindJSON(&notification); err != nil {
		logger.Error("Failed to bind JSON: ", err)
		response.ResponseError(c, http.StatusBadRequest, "Format data tidak valid")
		return
	}

	err := p.paymentSvc.ProcessPaymentCallback(c.Request.Context(), notification)
	if err != nil {
		logger.Error("Failed to process payment callback: ", err)
		response.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.ResponseSuccess(c, http.StatusOK, nil, nil, "Callback berhasil diproses")
}
