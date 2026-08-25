package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	"exchange/internal/db/repository"
	"exchange/internal/models"
)

type P2PService struct {
	p2pRepo *repository.P2PRepo
}

func NewP2PService(p2pRepo *repository.P2PRepo) *P2PService {
	return &P2PService{
		p2pRepo: p2pRepo,
	}
}

func (s *P2PService) CreateOrder(ctx context.Context, sellerID uuid.UUID, assetID int, fiatCurrency, rate, minAmount, maxAmount, totalAmount, paymentMethod string) (*models.P2POrder, error) {
	// Parse amounts to check basic logic (e.g. max >= min)
	minVal, _ := strconv.ParseFloat(minAmount, 64)
	maxVal, _ := strconv.ParseFloat(maxAmount, 64)
	totalVal, _ := strconv.ParseFloat(totalAmount, 64)

	if minVal <= 0 || maxVal <= 0 || totalVal <= 0 {
		return nil, errors.New("amounts must be greater than zero")
	}
	if maxVal < minVal {
		return nil, errors.New("max amount cannot be less than min amount")
	}
	if totalVal < maxVal {
		return nil, errors.New("total amount cannot be less than max amount")
	}

	order := &models.P2POrder{
		ID:              uuid.New(),
		SellerID:        sellerID,
		AssetID:         assetID,
		FiatCurrency:    fiatCurrency,
		Rate:            rate,
		MinAmount:       minAmount,
		MaxAmount:       maxAmount,
		AvailableAmount: totalAmount,
		PaymentMethod:   paymentMethod,
		Status:          models.P2POrderActive,
		CreatedAt:       time.Now(),
	}

	if err := s.p2pRepo.CreateOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("creating p2p order: %w", err)
	}

	return order, nil
}

func (s *P2PService) ListActiveOrders(ctx context.Context) ([]*models.P2POrder, error) {
	return s.p2pRepo.ListActiveOrders(ctx)
}

func (s *P2PService) CreateTrade(ctx context.Context, buyerID uuid.UUID, orderID uuid.UUID, amount, fiatAmount string) (*models.P2PTrade, error) {
	order, err := s.p2pRepo.GetOrder(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("fetching order: %w", err)
	}
	if order == nil {
		return nil, errors.New("order not found")
	}

	if order.SellerID == buyerID {
		return nil, errors.New("cannot trade with yourself")
	}

	// Basic amount validations
	reqAmt, _ := strconv.ParseFloat(amount, 64)
	minAmt, _ := strconv.ParseFloat(order.MinAmount, 64)
	maxAmt, _ := strconv.ParseFloat(order.MaxAmount, 64)
	availAmt, _ := strconv.ParseFloat(order.AvailableAmount, 64)

	if reqAmt < minAmt {
		return nil, errors.New("amount below minimum")
	}
	if reqAmt > maxAmt {
		return nil, errors.New("amount above maximum")
	}
	if reqAmt > availAmt {
		return nil, errors.New("amount exceeds available balance")
	}

	rateFloat, _ := strconv.ParseFloat(order.Rate, 64)
	calculatedFiatAmount := fmt.Sprintf("%.2f", reqAmt*rateFloat)

	trade := &models.P2PTrade{
		ID:           uuid.New(),
		OrderID:      order.ID,
		BuyerID:      buyerID,
		SellerID:     order.SellerID,
		AssetID:      order.AssetID,
		Amount:       amount,
		FiatAmount:   calculatedFiatAmount,
		Rate:         order.Rate,
		Status:       models.TradeWaitingPayment,
		EscrowLocked: true,
		CreatedAt:    time.Now(),
	}

	if err := s.p2pRepo.CreateTrade(ctx, trade); err != nil {
		return nil, fmt.Errorf("creating trade: %w", err)
	}

	return trade, nil
}

func (s *P2PService) ListUserTrades(ctx context.Context, userID uuid.UUID) ([]*models.P2PTrade, error) {
	return s.p2pRepo.ListUserTrades(ctx, userID)
}

func (s *P2PService) GetTrade(ctx context.Context, tradeID uuid.UUID, userID uuid.UUID) (*models.P2PTrade, error) {
	trade, err := s.p2pRepo.GetTrade(ctx, tradeID)
	if err != nil {
		return nil, err
	}
	if trade == nil {
		return nil, errors.New("trade not found")
	}

	// Ensure caller is buyer or seller
	if trade.BuyerID != userID && trade.SellerID != userID {
		return nil, errors.New("unauthorized access to trade")
	}

	return trade, nil
}

func (s *P2PService) MarkTradePaid(ctx context.Context, tradeID uuid.UUID, buyerID uuid.UUID) error {
	trade, err := s.p2pRepo.GetTrade(ctx, tradeID)
	if err != nil {
		return err
	}
	if trade == nil {
		return errors.New("trade not found")
	}

	if trade.BuyerID != buyerID {
		return errors.New("only the buyer can mark as paid")
	}

	if trade.Status != models.TradeWaitingPayment {
		return errors.New("trade is not awaiting payment")
	}

	return s.p2pRepo.MarkTradePaid(ctx, tradeID)
}

func (s *P2PService) ReleaseTrade(ctx context.Context, tradeID uuid.UUID, sellerID uuid.UUID) error {
	trade, err := s.p2pRepo.GetTrade(ctx, tradeID)
	if err != nil {
		return err
	}
	if trade == nil {
		return errors.New("trade not found")
	}

	if trade.SellerID != sellerID {
		return errors.New("only the seller can release the trade")
	}

	if trade.Status != models.TradePaid {
		return errors.New("trade has not been marked as paid by the buyer yet")
	}

	return s.p2pRepo.ReleaseTrade(ctx, tradeID)
}

func (s *P2PService) CancelTrade(ctx context.Context, tradeID uuid.UUID, userID uuid.UUID) error {
	trade, err := s.p2pRepo.GetTrade(ctx, tradeID)
	if err != nil {
		return err
	}
	if trade == nil {
		return errors.New("trade not found")
	}

	if trade.BuyerID != userID && trade.SellerID != userID {
		return errors.New("unauthorized")
	}

	if trade.Status == models.TradeReleased || trade.Status == models.TradeCancelled {
		return errors.New("trade cannot be cancelled in its current state")
	}

	return s.p2pRepo.CancelTrade(ctx, tradeID)
}
