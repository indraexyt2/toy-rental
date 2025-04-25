package service

import (
	"context"
	"errors"
	"final-project/entity"
	"final-project/repository"
	"final-project/utils/helpers"
	"gorm.io/gorm"
)

type IPaymentService interface {
	IBaseService[entity.Payment]
	CreatePaymentForRental(ctx context.Context, rentalID string) (*entity.Payment, error)
	CreatePaymentForExtension(ctx context.Context, rentalID string, metadata *entity.ExtensionMetadata) (*entity.Payment, error)
	ProcessPaymentCallback(ctx context.Context, notification map[string]interface{}) error
	GetPaymentByTransactionID(ctx context.Context, transactionID string) (*entity.Payment, error)
	FindByRentalID(ctx context.Context, rentalID string) ([]entity.Payment, error)
}

type PaymentService struct {
	BaseService[entity.Payment]
	paymentRepo     repository.IPaymentRepository
	rentalRepo      repository.IRentalRepository
	midtransService IMidtransService
}

func NewPaymentService(
	paymentRepo repository.IPaymentRepository,
	rentalRepo repository.IRentalRepository,
	midtransService IMidtransService,
) IPaymentService {
	return &PaymentService{
		BaseService:     BaseService[entity.Payment]{repository: paymentRepo},
		paymentRepo:     paymentRepo,
		rentalRepo:      rentalRepo,
		midtransService: midtransService,
	}
}

func (s *PaymentService) CreatePaymentForRental(ctx context.Context, rentalID string) (*entity.Payment, error) {
	var logger = helpers.Logger

	rental, err := s.rentalRepo.FindById(ctx, rentalID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("rental tidak ditemukan")
		}
		return nil, err
	}

	if rental.PaymentStatus == entity.PaymentStatusPaid {
		return nil, errors.New("rental sudah dibayar")
	}

	totalAmount := rental.TotalRentalPrice
	if rental.LateFee > 0 {
		totalAmount += rental.LateFee
	}
	if rental.DamageFee > 0 {
		totalAmount += rental.DamageFee
	}

	paymentType := entity.PaymentTypeRental
	if rental.LateFee > 0 && rental.DamageFee > 0 {
		paymentType = entity.PaymentTypeCombined
	} else if rental.LateFee > 0 {
		paymentType = entity.PaymentTypeLateFee
	} else if rental.DamageFee > 0 {
		paymentType = entity.PaymentTypeDamageFee
	}

	payment := &entity.Payment{
		RentalID:          rental.ID,
		PaymentType:       paymentType,
		GrossAmount:       totalAmount,
		TransactionStatus: entity.TransactionStatusPending,
	}

	payment, err = s.midtransService.CreateTransaction(ctx, payment, &rental)
	if err != nil {
		return nil, err
	}

	if err := s.paymentRepo.Insert(ctx, payment); err != nil {
		return nil, err
	}

	rental.PaymentStatus = entity.PaymentStatusPending
	if err := s.rentalRepo.UpdatePaymentStatus(ctx, rental.ID.String(), entity.PaymentStatusPending); err != nil {
		logger.Error("Gagal update status pembayaran rental: ", err)
	}

	return payment, nil
}

func (s *PaymentService) CreatePaymentForExtension(ctx context.Context, rentalID string, metadata *entity.ExtensionMetadata) (*entity.Payment, error) {
	rental, err := s.rentalRepo.FindById(ctx, rentalID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("rental tidak ditemukan")
		}
		return nil, err
	}

	if metadata.AdditionalCost <= 0 {
		return nil, errors.New("biaya perpanjangan harus lebih dari 0")
	}

	payment := &entity.Payment{
		RentalID:          rental.ID,
		PaymentType:       entity.PaymentTypeExtension,
		GrossAmount:       metadata.AdditionalCost,
		TransactionStatus: entity.TransactionStatusPending,
	}

	err = payment.SetExtensionMetadata(metadata)
	if err != nil {
		return nil, errors.New("gagal menyimpan metadata perpanjangan: " + err.Error())
	}

	payment, err = s.midtransService.CreateTransaction(ctx, payment, &rental)
	if err != nil {
		return nil, err
	}

	if err := s.paymentRepo.SavePaymentWithMetadata(ctx, payment); err != nil {
		return nil, err
	}

	return payment, nil
}

func (s *PaymentService) ProcessPaymentCallback(ctx context.Context, notification map[string]interface{}) error {
	txStatus, err := s.midtransService.VerifyPayment(ctx, notification)
	if err != nil {
		return err
	}

	payment, err := s.GetPaymentByTransactionID(ctx, txStatus.OrderID)
	if err != nil {
		return err
	}

	payment.TransactionStatus = txStatus.TransactionStatus

	if err := s.paymentRepo.UpdateByID(ctx, payment.ID.String(), payment); err != nil {
		return err
	}

	if payment.PaymentType == entity.PaymentTypeExtension {
		metadata, err := payment.GetExtensionMetadata()
		if err != nil {
			return errors.New("gagal mendapatkan metadata perpanjangan: " + err.Error())
		}

		switch txStatus.TransactionStatus {
		case "capture", "settlement":
			return nil
		case "pending":
			return nil
		case "deny", "cancel", "expire", "failure":
			err = s.rentalRepo.RollbackExtension(ctx, payment.RentalID.String(),
				metadata.OldExpectedReturnDate, metadata.OriginalRentalPrice)
			if err != nil {
				return errors.New("gagal membatalkan perpanjangan: " + err.Error())
			}
			return nil
		}
	} else {
		var rentalPaymentStatus string
		var rentalStatus string

		switch txStatus.TransactionStatus {
		case "capture", "settlement":
			rentalPaymentStatus = entity.PaymentStatusPaid
			rentalStatus = entity.RentalStatusActive
		case "pending":
			rentalPaymentStatus = entity.PaymentStatusPending
			rentalStatus = entity.RentalStatusPending
		case "deny", "cancel", "expire", "failure":
			rentalPaymentStatus = entity.PaymentStatusFailed
			rentalStatus = entity.RentalStatusPending
		case "refund":
			rentalPaymentStatus = entity.PaymentStatusRefunded
			rentalStatus = ""
		}

		if rentalPaymentStatus != "" {
			if err := s.rentalRepo.UpdatePaymentStatus(ctx, payment.RentalID.String(), rentalPaymentStatus); err != nil {
				return err
			}
		}

		if rentalStatus != "" {
			if err := s.rentalRepo.UpdateStatus(ctx, payment.RentalID.String(), rentalStatus); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *PaymentService) GetPaymentByTransactionID(ctx context.Context, transactionID string) (*entity.Payment, error) {
	payment, err := s.paymentRepo.FindByOrderID(ctx, transactionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("payment tidak ditemukan")
		}
		return nil, err
	}
	return &payment, nil
}

func (s *PaymentService) FindByRentalID(ctx context.Context, rentalID string) ([]entity.Payment, error) {
	return s.paymentRepo.FindByRentalID(ctx, rentalID)
}
