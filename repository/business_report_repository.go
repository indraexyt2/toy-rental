package repository

import (
	"context"
	"final-project/entity"
	"time"

	_ "github.com/gofrs/uuid/v5"
	"gorm.io/gorm"
)

type IBusinessReportRepository interface {
	GetSalesReport(ctx context.Context, startDate, endDate time.Time, groupBy string) ([]entity.SalesReportItem, error)
	GetPopularToys(ctx context.Context, startDate, endDate time.Time, limit int) ([]entity.PopularToyItem, error)
	GetTopCustomers(ctx context.Context, startDate, endDate time.Time, limit int) ([]entity.TopCustomerItem, error)
	GetRentalStatusCount(ctx context.Context, startDate, endDate time.Time) ([]entity.RentalStatusItem, error)
}

type BusinessReportRepository struct {
	DB *gorm.DB
}

func NewBusinessReportRepository(db *gorm.DB) IBusinessReportRepository {
	return &BusinessReportRepository{
		DB: db,
	}
}

func (r *BusinessReportRepository) GetSalesReport(ctx context.Context, startDate, endDate time.Time, groupBy string) ([]entity.SalesReportItem, error) {
	var items []entity.SalesReportItem
	var query string

	switch groupBy {
	case "day":
		query = `
			SELECT 
				TO_CHAR(r.rental_date, 'YYYY-MM-DD') as date,
				COUNT(r.id) as rental_count,
				SUM(r.total_rental_price) as rental_revenue,
				SUM(COALESCE(r.late_fee, 0)) as late_fee_revenue,
				SUM(COALESCE(r.damage_fee, 0)) as damage_fee_revenue,
				SUM(r.total_rental_price + COALESCE(r.late_fee, 0) + COALESCE(r.damage_fee, 0)) as total_revenue,
				COUNT(DISTINCT r.id) as transaction_count
			FROM 
				rentals r
			WHERE 
				r.rental_date BETWEEN ? AND ?
				AND r.deleted_at IS NULL
			GROUP BY 
				TO_CHAR(r.rental_date, 'YYYY-MM-DD')
			ORDER BY 
				date ASC
		`
	case "week":
		query = `
			SELECT 
				TO_CHAR(DATE_TRUNC('week', r.rental_date), 'YYYY-MM-DD') as date,
				COUNT(r.id) as rental_count,
				SUM(r.total_rental_price) as rental_revenue,
				SUM(COALESCE(r.late_fee, 0)) as late_fee_revenue,
				SUM(COALESCE(r.damage_fee, 0)) as damage_fee_revenue,
				SUM(r.total_rental_price + COALESCE(r.late_fee, 0) + COALESCE(r.damage_fee, 0)) as total_revenue,
				COUNT(DISTINCT r.id) as transaction_count
			FROM 
				rentals r
			WHERE 
				r.rental_date BETWEEN ? AND ?
				AND r.deleted_at IS NULL
			GROUP BY 
				TO_CHAR(DATE_TRUNC('week', r.rental_date), 'YYYY-MM-DD')
			ORDER BY 
				date ASC
		`
	case "month":
		query = `
			SELECT 
				TO_CHAR(r.rental_date, 'YYYY-MM') as date,
				COUNT(r.id) as rental_count,
				SUM(r.total_rental_price) as rental_revenue,
				SUM(COALESCE(r.late_fee, 0)) as late_fee_revenue,
				SUM(COALESCE(r.damage_fee, 0)) as damage_fee_revenue,
				SUM(r.total_rental_price + COALESCE(r.late_fee, 0) + COALESCE(r.damage_fee, 0)) as total_revenue,
				COUNT(DISTINCT r.id) as transaction_count
			FROM 
				rentals r
			WHERE 
				r.rental_date BETWEEN ? AND ?
				AND r.deleted_at IS NULL
			GROUP BY 
				TO_CHAR(r.rental_date, 'YYYY-MM')
			ORDER BY 
				date ASC
		`
	default:
		query = `
			SELECT 
				TO_CHAR(r.rental_date, 'YYYY-MM-DD') as date,
				COUNT(r.id) as rental_count,
				SUM(r.total_rental_price) as rental_revenue,
				SUM(COALESCE(r.late_fee, 0)) as late_fee_revenue,
				SUM(COALESCE(r.damage_fee, 0)) as damage_fee_revenue,
				SUM(r.total_rental_price + COALESCE(r.late_fee, 0) + COALESCE(r.damage_fee, 0)) as total_revenue,
				COUNT(DISTINCT r.id) as transaction_count
			FROM 
				rentals r
			WHERE 
				r.rental_date BETWEEN ? AND ?
				AND r.deleted_at IS NULL
			GROUP BY 
				TO_CHAR(r.rental_date, 'YYYY-MM-DD')
			ORDER BY 
				date ASC
		`
	}

	err := r.DB.WithContext(ctx).Raw(query, startDate, endDate).Scan(&items).Error
	return items, err
}

func (r *BusinessReportRepository) GetPopularToys(ctx context.Context, startDate, endDate time.Time, limit int) ([]entity.PopularToyItem, error) {
	var items []entity.PopularToyItem

	query := `
		WITH category_names AS (
			SELECT 
				tc.toy_id,
				STRING_AGG(c.name, ', ') AS category_names
			FROM 
				toy_categories tc
			JOIN 
				categories c ON tc.toy_category_id = c.id
			WHERE 
				c.deleted_at IS NULL
			GROUP BY 
				tc.toy_id
		)
		SELECT 
			t.id AS toy_id,
			t.name AS toy_name,
			t.primary_image AS image_url,
			COUNT(DISTINCT ri.rental_id) AS rental_count,
			SUM(EXTRACT(DAY FROM (COALESCE(r.actual_return_date, CURRENT_DATE) - r.rental_date)) + 1) AS total_rental_days,
			AVG(EXTRACT(DAY FROM (COALESCE(r.actual_return_date, CURRENT_DATE) - r.rental_date)) + 1) AS average_duration,
			SUM(ri.price_per_unit * ri.quantity) AS revenue,
			ARRAY_AGG(DISTINCT c.name) AS categories
		FROM 
			toys t
		JOIN 
			rental_items ri ON t.id = ri.toy_id
		JOIN 
			rentals r ON ri.rental_id = r.id
		LEFT JOIN 
			toy_categories tc ON t.id = tc.toy_id
		LEFT JOIN 
			categories c ON tc.toy_category_id = c.id
		WHERE 
			r.rental_date BETWEEN ? AND ?
			AND r.deleted_at IS NULL 
			AND ri.deleted_at IS NULL
			AND t.deleted_at IS NULL
		GROUP BY 
			t.id, t.name, t.primary_image
		ORDER BY 
			rental_count DESC, revenue DESC
		LIMIT ?
	`

	err := r.DB.WithContext(ctx).Raw(query, startDate, endDate, limit).Scan(&items).Error
	return items, err
}

func (r *BusinessReportRepository) GetTopCustomers(ctx context.Context, startDate, endDate time.Time, limit int) ([]entity.TopCustomerItem, error) {
	var items []entity.TopCustomerItem

	query := `
		SELECT 
			u.id AS user_id,
			u.full_name,
			u.email,
			u.phone_number,
			COUNT(r.id) AS rental_count,
			SUM(r.total_rental_price + COALESCE(r.late_fee, 0) + COALESCE(r.damage_fee, 0)) AS total_spent,
			AVG(r.total_rental_price) AS average_rental_value,
			COUNT(CASE WHEN r.late_fee > 0 THEN 1 END) AS late_fee_count,
			COUNT(CASE WHEN r.damage_fee > 0 THEN 1 END) AS damage_fee_count,
			MIN(r.rental_date) AS first_rental_date,
			MAX(r.rental_date) AS last_rental_date
		FROM 
			users u
		JOIN 
			rentals r ON u.id = r.user_id
		WHERE 
			r.rental_date BETWEEN ? AND ?
			AND r.deleted_at IS NULL
			AND u.deleted_at IS NULL
		GROUP BY 
			u.id, u.full_name, u.email, u.phone_number
		ORDER BY 
			total_spent DESC, rental_count DESC
		LIMIT ?
	`

	err := r.DB.WithContext(ctx).Raw(query, startDate, endDate, limit).Scan(&items).Error
	return items, err
}

func (r *BusinessReportRepository) GetRentalStatusCount(ctx context.Context, startDate, endDate time.Time) ([]entity.RentalStatusItem, error) {
	var items []entity.RentalStatusItem
	var totalCount int64

	if err := r.DB.WithContext(ctx).Model(&entity.Rental{}).
		Where("rental_date BETWEEN ? AND ?", startDate, endDate).
		Where("deleted_at IS NULL").
		Count(&totalCount).Error; err != nil {
		return nil, err
	}

	query := `
		SELECT 
			status,
			COUNT(*) AS count
		FROM 
			rentals
		WHERE 
			rental_date BETWEEN ? AND ?
			AND deleted_at IS NULL
		GROUP BY 
			status
		ORDER BY 
			count DESC
	`

	err := r.DB.WithContext(ctx).Raw(query, startDate, endDate).Scan(&items).Error
	if err != nil {
		return nil, err
	}

	for i := range items {
		if totalCount > 0 {
			items[i].PercentageTotal = float64(items[i].Count) / float64(totalCount) * 100
		} else {
			items[i].PercentageTotal = 0
		}
	}

	return items, nil
}
