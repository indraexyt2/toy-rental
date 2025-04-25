package service

import (
	"context"
	"errors"
	"final-project/config"
	"final-project/entity"
	"final-project/utils/helpers"
	"fmt"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
	"time"
)

type IMidtransService interface {
	CreateTransaction(ctx context.Context, payment *entity.Payment, rental *entity.Rental) (*entity.Payment, error)
	VerifyPayment(ctx context.Context, notificationPayload map[string]interface{}) (*coreapi.TransactionStatusResponse, error)
}

type MidtransService struct {
	snapClient    snap.Client
	coreAPIClient coreapi.Client
	serverKey     string
	clientKey     string
	isProduction  bool
}

func NewMidtransService(cfg *config.Config) IMidtransService {
	isProduction := cfg.MidtransEnv == "production"

	var snapClient snap.Client
	snapClient.New(cfg.MidtransServerKey, midtrans.Sandbox)
	if isProduction {
		snapClient.New(cfg.MidtransServerKey, midtrans.Production)
	} else {
		snapClient.New(cfg.MidtransServerKey, midtrans.Sandbox)
	}

	var coreAPIClient coreapi.Client
	if isProduction {
		coreAPIClient.New(cfg.MidtransServerKey, midtrans.Production)
	} else {
		coreAPIClient.New(cfg.MidtransServerKey, midtrans.Sandbox)
	}

	return &MidtransService{
		snapClient:    snapClient,
		coreAPIClient: coreAPIClient,
		serverKey:     cfg.MidtransServerKey,
		clientKey:     cfg.MidtransClientKey,
		isProduction:  isProduction,
	}
}

func (s *MidtransService) CreateTransaction(ctx context.Context, payment *entity.Payment, rental *entity.Rental) (*entity.Payment, error) {
	var logger = helpers.Logger
	var items []midtrans.ItemDetails

	totalDayRent := rental.ExpectedReturnDate.Sub(rental.RentalDate).Hours() / 24

	if rental.PaymentStatus != entity.PaymentStatusPaid {
		for _, item := range rental.RentalItems {
			toyName := fmt.Sprintf("Item %s", item.ToyID.String())
			if item.Toy.Name != "" {
				toyName = item.Toy.Name
			}

			itemDetail := midtrans.ItemDetails{
				ID:    item.ToyID.String(),
				Name:  toyName,
				Price: int64(item.PricePerUnit) * int64(totalDayRent),
				Qty:   int32(item.Quantity),
			}
			items = append(items, itemDetail)
		}
	}

	if rental.LateFee > 0 {
		lateFeeItem := midtrans.ItemDetails{
			ID:    "late_fee",
			Name:  "Biaya Keterlambatan",
			Price: int64(rental.LateFee),
			Qty:   1,
		}
		items = append(items, lateFeeItem)
	}

	if rental.DamageFee > 0 {
		damageFeeItem := midtrans.ItemDetails{
			ID:    "damage_fee",
			Name:  "Biaya Kerusakan",
			Price: int64(rental.DamageFee),
			Qty:   1,
		}
		items = append(items, damageFeeItem)
	}

	customerDetails := &midtrans.CustomerDetails{
		FName: rental.User.FullName,
		Email: rental.User.Email,
		Phone: rental.User.PhoneNumber,
	}

	expiry := &snap.ExpiryDetails{
		StartTime: time.Now().Format("2006-01-02 15:04:05 -0700"),
		Unit:      "day",
		Duration:  1,
	}

	shortUUID := payment.ID.String()[:8]
	shortTime := time.Now().Format("060102150405")
	uniqueOrderID := shortUUID + "-" + shortTime
	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  uniqueOrderID,
			GrossAmt: int64(payment.GrossAmount),
		},
		CustomerDetail: customerDetails,
		Items:          &items,
		Expiry:         expiry,
	}

	snapResp, err := s.snapClient.CreateTransaction(snapReq)
	if err != nil {
		logger.Error("Error creating snap transaction: ", err)
		return nil, errors.New("gagal membuat transaksi: " + err.Error())
	}

	payment.SnapToken = snapResp.Token
	payment.SnapURL = snapResp.RedirectURL
	payment.OrderID = uniqueOrderID

	expiryTime := time.Now().Add(24 * time.Hour)
	payment.ExpiryTime = &expiryTime

	return payment, nil
}

func (s *MidtransService) VerifyPayment(ctx context.Context, notificationPayload map[string]interface{}) (*coreapi.TransactionStatusResponse, error) {
	var logger = helpers.Logger

	orderID, exists := notificationPayload["order_id"].(string)
	if !exists {
		return nil, errors.New("order_id tidak ditemukan pada notifikasi")
	}

	logger.Info("Verifikasi pembayaran untuk orderID: ", orderID)

	txStatus, err := s.coreAPIClient.CheckTransaction(orderID)
	if err != nil {
		logger.Error("Error checking transaction status: ", err)
		return nil, errors.New("gagal memeriksa status transaksi: " + err.Error())
	}

	logger.Info("Respons dari Midtrans API: OrderID=", txStatus.OrderID, ", Status=", txStatus.TransactionStatus)

	if txStatus.TransactionTime != "" {
		logger.Info("Waktu transaksi: ", txStatus.TransactionTime)
	}

	transactionStatusNotif, exists := notificationPayload["transaction_status"].(string)
	if !exists {
		logger.Warn("transaction_status tidak ditemukan pada notifikasi, menggunakan status dari API")
		return txStatus, nil
	}

	logger.Info("Status transaksi dari notifikasi: ", transactionStatusNotif)
	logger.Info("Status transaksi dari API: ", txStatus.TransactionStatus)

	return txStatus, nil
}
