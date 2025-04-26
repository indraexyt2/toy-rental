package entity

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type SalesReportItem struct {
	Date             string  `json:"date"`
	RentalCount      int     `json:"rental_count"`
	RentalRevenue    float64 `json:"rental_revenue"`
	LateFeeRevenue   float64 `json:"late_fee_revenue"`
	DamageFeeRevenue float64 `json:"damage_fee_revenue"`
	TotalRevenue     float64 `json:"total_revenue"`
	TransactionCount int     `json:"transaction_count"`
}

type PopularToyItem struct {
	ToyID           uuid.UUID `json:"toy_id"`
	ToyName         string    `json:"toy_name"`
	ImageURL        string    `json:"image_url"`
	RentalCount     int       `json:"rental_count"`
	AverageDuration float64   `json:"average_duration"`
	Revenue         float64   `json:"revenue"`
}

type TopCustomerItem struct {
	UserID             uuid.UUID `json:"user_id"`
	FullName           string    `json:"full_name"`
	Email              string    `json:"email"`
	PhoneNumber        string    `json:"phone_number"`
	RentalCount        int       `json:"rental_count"`
	TotalSpent         float64   `json:"total_spent"`
	AverageRentalValue float64   `json:"average_rental_value"`
	LateFeeCount       int       `json:"late_fee_count"`
	DamageFeeCount     int       `json:"damage_fee_count"`
	FirstRentalDate    time.Time `json:"first_rental_date"`
	LastRentalDate     time.Time `json:"last_rental_date"`
}

type InventoryStatusItem struct {
	ToyID           uuid.UUID `json:"toy_id"`
	ToyName         string    `json:"toy_name"`
	ImageURL        string    `json:"image_url"`
	Categories      []string  `json:"categories"`
	CurrentStock    int       `json:"current_stock"`
	TotalStock      int       `json:"total_stock"`
	RentedCount     int       `json:"rented_count"`
	AvailableCount  int       `json:"available_count"`
	DamagedCount    int       `json:"damaged_count"`
	LostCount       int       `json:"lost_count"`
	Condition       string    `json:"condition"`
	ReplacementCost float64   `json:"replacement_cost"`
}

type RentalStatusItem struct {
	Status          string  `json:"status"`
	Count           int     `json:"count"`
	PercentageTotal float64 `json:"percentage_total"`
}

type CategorySummary struct {
	Name        string  `json:"name"`
	RentalCount int     `json:"rental_count"`
	Percentage  float64 `json:"percentage"`
}
