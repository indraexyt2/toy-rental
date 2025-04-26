package service

import (
	"context"
	"errors"
	"final-project/entity"
	"final-project/repository"
	"fmt"
	"github.com/gofrs/uuid/v5"
)

type IRentalService interface {
	IBaseService[entity.Rental]
	CreateRental(ctx context.Context, req entity.CreateRentalRequest) (*entity.Rental, error)
	ReturnRental(ctx context.Context, id string, req entity.ReturnRentalRequest) (*entity.Rental, error)
	ExtendRental(ctx context.Context, id string, req entity.ExtendRentalRequest) (*entity.Rental, *entity.Payment, error)
}

type RentalService struct {
	BaseService[entity.Rental]
	rentalRepo repository.IRentalRepository
	userRepo   repository.IUserRepository
	toyRepo    repository.IToyRepository
	paymentSvc IPaymentService
}

func NewRentalService(
	repo repository.IRentalRepository,
	userRepo repository.IUserRepository,
	toyRepo repository.IToyRepository,
	paymentSvc IPaymentService,
) IRentalService {
	return &RentalService{
		BaseService: BaseService[entity.Rental]{repository: repo},
		rentalRepo:  repo,
		userRepo:    userRepo,
		toyRepo:     toyRepo,
		paymentSvc:  paymentSvc,
	}
}

func (s *RentalService) CreateRental(ctx context.Context, req entity.CreateRentalRequest) (*entity.Rental, error) {
	rental := &entity.Rental{
		UserID:             req.UserID,
		Status:             "pending",
		RentalDate:         req.RentalDate,
		ExpectedReturnDate: req.ExpectedReturnDate,
		TotalRentalPrice:   0,
		PaymentStatus:      "unpaid",
		Notes:              req.Notes,
		RentalItems:        make([]entity.RentalItem, 0, len(req.Items)),
	}

	rentalDays := int(req.ExpectedReturnDate.Sub(req.RentalDate).Hours() / 24)
	if rentalDays < 1 {
		return nil, errors.New("tanggal pengembalian harus setelah tanggal rental")
	}

	var totalPrice float64 = 0
	for _, item := range req.Items {
		toy, err := s.toyRepo.FindById(ctx, item.ToyID.String())
		if err != nil {
			return nil, errors.New("mainan tidak ditemukan: " + item.ToyID.String())
		}

		if toy.Stock < item.Quantity {
			return nil, errors.New("stok mainan tidak mencukupi: " + toy.Name)
		}

		pricePerUnit := toy.RentalPrice
		itemTotalPrice := float64(item.Quantity) * pricePerUnit * float64(rentalDays)
		totalPrice += itemTotalPrice

		rentalItem := entity.RentalItem{
			ToyID:           item.ToyID,
			Quantity:        item.Quantity,
			PricePerUnit:    pricePerUnit,
			ConditionBefore: item.ConditionBefore,
			ConditionAfter:  item.ConditionBefore,
			Status:          "rented",
		}

		rental.RentalItems = append(rental.RentalItems, rentalItem)
	}

	rental.TotalRentalPrice = totalPrice
	if err := s.repository.Insert(ctx, rental); err != nil {
		return nil, err
	}

	newRent, err := s.rentalRepo.FindById(ctx, rental.ID.String())
	if err != nil {
		return nil, err
	}

	fmt.Printf("rental: %+v\n", newRent)

	return &newRent, nil
}

func (s *RentalService) ReturnRental(ctx context.Context, id string, req entity.ReturnRentalRequest) (*entity.Rental, error) {
	rental, err := s.repository.FindById(ctx, id)
	if err != nil {
		return nil, errors.New("rental tidak ditemukan")
	}

	for _, item := range rental.RentalItems {
		fmt.Println(item)
	}

	if rental.Status == entity.RentalStatusCompleted || rental.Status == entity.RentalStatusCancelled {
		return nil, errors.New("rental sudah selesai atau dibatalkan")
	}

	if req.ActualReturnDate.Before(rental.RentalDate) {
		return nil, errors.New("tanggal pengembalian tidak boleh sebelum tanggal rental")
	}

	rental.ActualReturnDate = &req.ActualReturnDate

	rentalItemMap := make(map[uuid.UUID]*entity.RentalItem)
	for i := range rental.RentalItems {
		rentalItemMap[rental.RentalItems[i].ID] = &rental.RentalItems[i]
	}

	var totalLateFee float64 = 0

	if req.ActualReturnDate.After(rental.ExpectedReturnDate) {
		days := int(req.ActualReturnDate.Sub(rental.ExpectedReturnDate).Hours()/48) + 1

		for _, rentalItem := range rental.RentalItems {
			toy, err := s.toyRepo.FindById(ctx, rentalItem.ToyID.String())
			if err != nil {
				return nil, errors.New("tidak dapat mendapatkan data mainan: " + rentalItem.ToyID.String())
			}

			itemLateFee := toy.LateFeePerDay * float64(days) * float64(rentalItem.Quantity)
			totalLateFee += itemLateFee
		}

		rental.Status = "overdue"
	} else {
		rental.Status = "completed"
	}

	rental.LateFee = totalLateFee

	var totalDamageFee float64 = 0

	for _, itemReq := range req.Items {
		rentalItem, exists := rentalItemMap[itemReq.RentalItemID]
		if !exists {
			return nil, errors.New("item rental dengan ID " + itemReq.RentalItemID.String() + " tidak ditemukan")
		}

		validConditions := []string{"new", "excellent", "good", "fair", "poor", "damaged", "lost"}
		validCondition := false
		for _, c := range validConditions {
			if itemReq.ConditionAfter == c {
				validCondition = true
				break
			}
		}
		if !validCondition {
			return nil, errors.New("kondisi tidak valid: " + itemReq.ConditionAfter)
		}

		rentalItem.ConditionAfter = itemReq.ConditionAfter
		rentalItem.DamageDescription = itemReq.DamageDescription

		var damageFee float64 = 0

		toy, _ := s.toyRepo.FindById(ctx, rentalItem.ToyID.String())

		if itemReq.ConditionAfter == "lost" {
			damageFee = toy.ReplacementPrice * float64(rentalItem.Quantity)
			rentalItem.Status = "lost"
		} else if itemReq.ConditionAfter == "damaged" {
			damageFee = toy.ReplacementPrice * 0.7 * float64(rentalItem.Quantity)
			rentalItem.Status = "damaged"
		} else {
			conditionValues := map[string]int{
				"new":       5,
				"excellent": 4,
				"good":      3,
				"fair":      2,
				"poor":      1,
			}

			beforeValue := conditionValues[rentalItem.ConditionBefore]
			afterValue := conditionValues[itemReq.ConditionAfter]

			if afterValue < beforeValue {
				damageFee = toy.ReplacementPrice * 0.15 * float64(beforeValue-afterValue) * float64(rentalItem.Quantity)
			}

			rentalItem.Status = "returned"
		}

		rentalItem.DamageFee = damageFee
		totalDamageFee += damageFee

		if err := s.rentalRepo.UpdateRentalItem(ctx, rentalItem); err != nil {
			return nil, err
		}

		toy.Stock += rentalItem.Quantity
		if err := s.toyRepo.UpdateById(ctx, toy.ID.String(), &toy); err != nil {
			return nil, err
		}
	}

	rental.DamageFee = totalDamageFee

	if req.Notes != "" {
		rental.Notes = req.Notes
	}

	rental.TotalAmount = rental.TotalRentalPrice + rental.LateFee + rental.DamageFee

	if err := s.rentalRepo.ReturnRental(ctx, &rental); err != nil {
		return nil, err
	}

	return &rental, nil
}

func (s *RentalService) ExtendRental(ctx context.Context, id string, req entity.ExtendRentalRequest) (*entity.Rental, *entity.Payment, error) {
	rental, err := s.rentalRepo.FindById(ctx, id)
	if err != nil {
		return nil, nil, errors.New("rental tidak ditemukan")
	}

	if rental.Status != entity.RentalStatusActive {
		return nil, nil, errors.New("hanya rental dengan status aktif yang dapat diperpanjang")
	}

	if rental.PaymentStatus != entity.PaymentStatusPaid {
		return nil, nil, errors.New("rental harus sudah dibayar sebelum dapat diperpanjang")
	}

	if !req.NewExpectedReturnDate.After(rental.ExpectedReturnDate) {
		return nil, nil, errors.New("tanggal perpanjangan harus setelah tanggal pengembalian yang diharapkan saat ini")
	}

	additionalDays := int(req.NewExpectedReturnDate.Sub(rental.ExpectedReturnDate).Hours() / 24)
	if additionalDays <= 0 {
		return nil, nil, errors.New("perpanjangan minimal 1 hari")
	}

	oldExpectedReturnDate := rental.ExpectedReturnDate
	oldTotalPrice := rental.TotalRentalPrice

	var additionalCost float64 = 0
	for i, item := range rental.RentalItems {
		toy, err := s.toyRepo.FindById(ctx, item.ToyID.String())
		if err != nil {
			return nil, nil, errors.New("tidak dapat mendapatkan data mainan: " + item.ToyID.String())
		}

		itemExtensionCost := toy.RentalPrice * float64(additionalDays) * float64(item.Quantity)
		additionalCost += itemExtensionCost
		fmt.Printf("Item %d: %s x %d = %f", i, toy.Name, item.Quantity, itemExtensionCost)
		fmt.Println("")
		fmt.Printf("Total: %f", additionalCost)
	}

	fmt.Println("Additional Cost svc:", additionalCost)

	var extensionNotes string
	if req.Notes != "" {
		extensionNotes = "Perpanjangan: " + req.Notes
	} else {
		extensionNotes = "Perpanjangan dari " + oldExpectedReturnDate.Format("2006-01-02") +
			" ke " + req.NewExpectedReturnDate.Format("2006-01-02")
	}

	err = s.rentalRepo.ExtendRental(ctx, &rental, req.NewExpectedReturnDate, additionalCost, extensionNotes)
	if err != nil {
		return nil, nil, errors.New("gagal memperpanjang rental: " + err.Error())
	}

	rental.ExpectedReturnDate = req.NewExpectedReturnDate
	rental.TotalRentalPrice += additionalCost
	if rental.Notes == "" {
		rental.Notes = extensionNotes
	} else {
		rental.Notes += "\n" + extensionNotes
	}

	metadata := &entity.ExtensionMetadata{
		OldExpectedReturnDate: oldExpectedReturnDate,
		NewExpectedReturnDate: req.NewExpectedReturnDate,
		AdditionalDays:        additionalDays,
		OriginalRentalPrice:   oldTotalPrice,
		AdditionalCost:        additionalCost,
	}

	payment, err := s.paymentSvc.CreatePaymentForExtension(ctx, id, metadata)
	if err != nil {
		s.rentalRepo.RollbackExtension(ctx, id, oldExpectedReturnDate, oldTotalPrice)
		return nil, nil, errors.New("gagal membuat pembayaran perpanjangan: " + err.Error())
	}

	return &rental, payment, nil
}
