package service

import (
	"context"
	"final-project/entity"
	"final-project/repository"
	"time"
)

type IBusinessReportService interface {
	GetSalesReport(ctx context.Context, startDate, endDate time.Time, groupBy string) ([]entity.SalesReportItem, error)
	GetPopularToysReport(ctx context.Context, startDate, endDate time.Time, limit int) ([]entity.PopularToyItem, error)
	GetTopCustomersReport(ctx context.Context, startDate, endDate time.Time, limit int) ([]entity.TopCustomerItem, error)
	GetRentalStatusReport(ctx context.Context, startDate, endDate time.Time) ([]entity.RentalStatusItem, error)
}

type BusinessReportService struct {
	reportRepo repository.IBusinessReportRepository
}

func NewBusinessReportService(reportRepo repository.IBusinessReportRepository) IBusinessReportService {
	return &BusinessReportService{
		reportRepo: reportRepo,
	}
}

func (s *BusinessReportService) GetSalesReport(ctx context.Context, startDate, endDate time.Time, groupBy string) ([]entity.SalesReportItem, error) {
	validGroupBy := map[string]bool{
		"day":   true,
		"week":  true,
		"month": true,
	}

	if !validGroupBy[groupBy] {
		groupBy = "day"
	}

	return s.reportRepo.GetSalesReport(ctx, startDate, endDate, groupBy)
}

func (s *BusinessReportService) GetPopularToysReport(ctx context.Context, startDate, endDate time.Time, limit int) ([]entity.PopularToyItem, error) {
	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	return s.reportRepo.GetPopularToys(ctx, startDate, endDate, limit)
}

func (s *BusinessReportService) GetTopCustomersReport(ctx context.Context, startDate, endDate time.Time, limit int) ([]entity.TopCustomerItem, error) {
	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	return s.reportRepo.GetTopCustomers(ctx, startDate, endDate, limit)
}

func (s *BusinessReportService) GetRentalStatusReport(ctx context.Context, startDate, endDate time.Time) ([]entity.RentalStatusItem, error) {
	return s.reportRepo.GetRentalStatusCount(ctx, startDate, endDate)
}
