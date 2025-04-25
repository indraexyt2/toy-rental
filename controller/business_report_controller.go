package controller

import (
	"final-project/service"
	"final-project/utils/helpers"
	"final-project/utils/response"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type IBusinessReportController interface {
	GetSalesReport(c *gin.Context)
	GetPopularToysReport(c *gin.Context)
	GetTopCustomersReport(c *gin.Context)
	GetRentalStatusReport(c *gin.Context)
}

type BusinessReportController struct {
	reportSvc service.IBusinessReportService
}

func NewBusinessReportController(reportSvc service.IBusinessReportService) IBusinessReportController {
	return &BusinessReportController{
		reportSvc: reportSvc,
	}
}

// GetSalesReport godoc
// @Summary Mendapatkan laporan penjualan
// @Description Mendapatkan laporan penjualan dalam rentang waktu tertentu
// @Tags Business Report
// @Produce json
// @Param start_date query string true "Tanggal mulai (YYYY-MM-DD)"
// @Param end_date query string true "Tanggal akhir (YYYY-MM-DD)"
// @Param group_by query string false "Pengelompokan (day, week, month)"
// @Success 200 {object} response.APISuccessResponse
// @Router /business-report/sales [get]
func (r *BusinessReportController) GetSalesReport(c *gin.Context) {
	var logger = helpers.Logger

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	groupBy := c.DefaultQuery("group_by", "day")

	if startDateStr == "" || endDateStr == "" {
		logger.Error("Tanggal mulai dan akhir wajib diisi")
		response.ResponseError(c, http.StatusBadRequest, "Tanggal mulai dan akhir wajib diisi")
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		logger.Error("Format tanggal mulai tidak valid: ", err)
		response.ResponseError(c, http.StatusBadRequest, "Format tanggal mulai tidak valid (YYYY-MM-DD)")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		logger.Error("Format tanggal akhir tidak valid: ", err)
		response.ResponseError(c, http.StatusBadRequest, "Format tanggal akhir tidak valid (YYYY-MM-DD)")
		return
	}

	endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	if endDate.Before(startDate) {
		logger.Error("Tanggal akhir tidak boleh sebelum tanggal mulai")
		response.ResponseError(c, http.StatusBadRequest, "Tanggal akhir tidak boleh sebelum tanggal mulai")
		return
	}

	salesReport, err := r.reportSvc.GetSalesReport(c.Request.Context(), startDate, endDate, groupBy)
	if err != nil {
		logger.Error("Gagal mendapatkan laporan penjualan: ", err)
		response.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	var totalRevenue float64
	var totalTransactions int
	for _, item := range salesReport {
		totalRevenue += item.TotalRevenue
		totalTransactions += item.TransactionCount
	}

	metadata := map[string]interface{}{
		"periode_mulai":    startDateStr,
		"periode_akhir":    endDateStr,
		"total_pendapatan": totalRevenue,
		"total_transaksi":  totalTransactions,
		"pengelompokan":    groupBy,
	}

	response.ResponseSuccess(c, http.StatusOK, salesReport, metadata, "Berhasil mendapatkan laporan penjualan")
}

// GetPopularToysReport godoc
// @Summary Mendapatkan laporan mainan populer
// @Description Mendapatkan laporan mainan paling populer berdasarkan jumlah penyewaan
// @Tags Business Report
// @Produce json
// @Param start_date query string true "Tanggal mulai (YYYY-MM-DD)"
// @Param end_date query string true "Tanggal akhir (YYYY-MM-DD)"
// @Param limit query int false "Jumlah data (default: 10)"
// @Success 200 {object} response.APISuccessResponse
// @Router /business-report/popular-toys [get]
func (r *BusinessReportController) GetPopularToysReport(c *gin.Context) {
	var logger = helpers.Logger

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	limitStr := c.DefaultQuery("limit", "10")
	limit := helpers.ParseToInt(limitStr)

	if startDateStr == "" || endDateStr == "" {
		logger.Error("Tanggal mulai dan akhir wajib diisi")
		response.ResponseError(c, http.StatusBadRequest, "Tanggal mulai dan akhir wajib diisi")
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		logger.Error("Format tanggal mulai tidak valid: ", err)
		response.ResponseError(c, http.StatusBadRequest, "Format tanggal mulai tidak valid (YYYY-MM-DD)")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		logger.Error("Format tanggal akhir tidak valid: ", err)
		response.ResponseError(c, http.StatusBadRequest, "Format tanggal akhir tidak valid (YYYY-MM-DD)")
		return
	}

	endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	if endDate.Before(startDate) {
		logger.Error("Tanggal akhir tidak boleh sebelum tanggal mulai")
		response.ResponseError(c, http.StatusBadRequest, "Tanggal akhir tidak boleh sebelum tanggal mulai")
		return
	}

	popularToys, err := r.reportSvc.GetPopularToysReport(c.Request.Context(), startDate, endDate, limit)
	if err != nil {
		logger.Error("Gagal mendapatkan laporan mainan populer: ", err)
		response.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	metadata := map[string]interface{}{
		"periode_mulai": startDateStr,
		"periode_akhir": endDateStr,
		"jumlah_mainan": len(popularToys),
	}

	response.ResponseSuccess(c, http.StatusOK, popularToys, metadata, "Berhasil mendapatkan laporan mainan populer")
}

// GetTopCustomersReport godoc
// @Summary Mendapatkan laporan pelanggan teratas
// @Description Mendapatkan laporan pelanggan yang paling aktif berdasarkan jumlah penyewaan dan pengeluaran
// @Tags Business Report
// @Produce json
// @Param start_date query string true "Tanggal mulai (YYYY-MM-DD)"
// @Param end_date query string true "Tanggal akhir (YYYY-MM-DD)"
// @Param limit query int false "Jumlah data (default: 10)"
// @Success 200 {object} response.APISuccessResponse
// @Router /business-report/customers [get]
func (r *BusinessReportController) GetTopCustomersReport(c *gin.Context) {
	var logger = helpers.Logger

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	limitStr := c.DefaultQuery("limit", "10")
	limit := helpers.ParseToInt(limitStr)

	if startDateStr == "" || endDateStr == "" {
		logger.Error("Tanggal mulai dan akhir wajib diisi")
		response.ResponseError(c, http.StatusBadRequest, "Tanggal mulai dan akhir wajib diisi")
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		logger.Error("Format tanggal mulai tidak valid: ", err)
		response.ResponseError(c, http.StatusBadRequest, "Format tanggal mulai tidak valid (YYYY-MM-DD)")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		logger.Error("Format tanggal akhir tidak valid: ", err)
		response.ResponseError(c, http.StatusBadRequest, "Format tanggal akhir tidak valid (YYYY-MM-DD)")
		return
	}

	endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	if endDate.Before(startDate) {
		logger.Error("Tanggal akhir tidak boleh sebelum tanggal mulai")
		response.ResponseError(c, http.StatusBadRequest, "Tanggal akhir tidak boleh sebelum tanggal mulai")
		return
	}

	customers, err := r.reportSvc.GetTopCustomersReport(c.Request.Context(), startDate, endDate, limit)
	if err != nil {
		logger.Error("Gagal mendapatkan laporan pelanggan: ", err)
		response.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	metadata := map[string]interface{}{
		"periode_mulai":    startDateStr,
		"periode_akhir":    endDateStr,
		"jumlah_pelanggan": len(customers),
	}

	response.ResponseSuccess(c, http.StatusOK, customers, metadata, "Berhasil mendapatkan laporan pelanggan teratas")
}

// GetRentalStatusReport godoc
// @Summary Mendapatkan laporan status penyewaan
// @Description Mendapatkan laporan jumlah penyewaan berdasarkan status
// @Tags Business Report
// @Produce json
// @Param start_date query string true "Tanggal mulai (YYYY-MM-DD)"
// @Param end_date query string true "Tanggal akhir (YYYY-MM-DD)"
// @Success 200 {object} response.APISuccessResponse
// @Router /business-report/rental-status [get]
func (r *BusinessReportController) GetRentalStatusReport(c *gin.Context) {
	var logger = helpers.Logger

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		logger.Error("Tanggal mulai dan akhir wajib diisi")
		response.ResponseError(c, http.StatusBadRequest, "Tanggal mulai dan akhir wajib diisi")
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		logger.Error("Format tanggal mulai tidak valid: ", err)
		response.ResponseError(c, http.StatusBadRequest, "Format tanggal mulai tidak valid (YYYY-MM-DD)")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		logger.Error("Format tanggal akhir tidak valid: ", err)
		response.ResponseError(c, http.StatusBadRequest, "Format tanggal akhir tidak valid (YYYY-MM-DD)")
		return
	}

	endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	if endDate.Before(startDate) {
		logger.Error("Tanggal akhir tidak boleh sebelum tanggal mulai")
		response.ResponseError(c, http.StatusBadRequest, "Tanggal akhir tidak boleh sebelum tanggal mulai")
		return
	}

	statusReport, err := r.reportSvc.GetRentalStatusReport(c.Request.Context(), startDate, endDate)
	if err != nil {
		logger.Error("Gagal mendapatkan laporan status penyewaan: ", err)
		response.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	var totalRentals int
	for _, item := range statusReport {
		totalRentals += item.Count
	}

	metadata := map[string]interface{}{
		"periode_mulai":   startDateStr,
		"periode_akhir":   endDateStr,
		"total_penyewaan": totalRentals,
	}

	response.ResponseSuccess(c, http.StatusOK, statusReport, metadata, "Berhasil mendapatkan laporan status penyewaan")
}
