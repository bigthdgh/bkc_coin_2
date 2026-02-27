package marketplace

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"bkc_coin_v2/internal/database"
	"bkc_coin_v2/internal/nft"
)

// NFTAuction - аукцион NFT
type NFTAuction struct {
	AuctionID       int64         `json:"auction_id"`
	NFTID           int64         `json:"nft_id"`
	NFT             *nft.DynamicNFT `json:"nft"`
	SellerID        int64         `json:"seller_id"`
	StartingBid     int64         `json:"starting_bid"`
	CurrentBid      int64         `json:"current_bid"`
	MinBidIncrement int64         `json:"min_bid_increment"`
	BuyoutPrice     *int64        `json:"buyout_price,omitempty"`
	Status          string        `json:"status"` // active, ended, cancelled
	StartTime       time.Time     `json:"start_time"`
	EndTime         time.Time     `json:"end_time"`
	Bids            []AuctionBid  `json:"bids"`
	WinnerID        *int64        `json:"winner_id,omitempty"`
	FinalPrice      *int64        `json:"final_price,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

// AuctionBid - ставка на аукционе
type AuctionBid struct {
	BidID        int64     `json:"bid_id"`
	AuctionID    int64     `json:"auction_id"`
	BidderID     int64     `json:"bidder_id"`
	Amount       int64     `json:"amount"`
	IsAutoBid    bool      `json:"is_auto_bid"`
	MaxAutoBid   *int64    `json:"max_auto_bid,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// AuctionManager - менеджер аукционов
type AuctionManager struct {
	db           *database.UnifiedDB
	nftManager   *nft.DynamicNFTManager
	activeAuctions map[int64]*NFTAuction
}

// NewAuctionManager - создание менеджера аукционов
func NewAuctionManager(db *database.UnifiedDB, nftManager *nft.DynamicNFTManager) *AuctionManager {
	am := &AuctionManager{
		db:           db,
		nftManager:   nftManager,
		activeAuctions: make(map[int64]*NFTAuction),
	}
	
	// Запускаем обработку аукционов
	go am.processAuctions()
	
	return am
}

// CreateAuction - создание нового аукциона
func (am *AuctionManager) CreateAuction(ctx context.Context, nftID int64, sellerID int64, startingBid int64, duration time.Duration, buyoutPrice *int64) (*NFTAuction, error) {
	// Получаем NFT
	dynamicNFT, err := am.nftManager.GetDynamicNFT(ctx, nftID)
	if err != nil {
		return nil, fmt.Errorf("failed to get NFT: %w", err)
	}

	// Проверяем, что NFT принадлежит продавцу
	if dynamicNFT.OwnerID != sellerID {
		return nil, fmt.Errorf("NFT does not belong to seller")
	}

	// Создаем аукцион
	auction := &NFTAuction{
		AuctionID:       time.Now().UnixNano(),
		NFTID:           nftID,
		NFT:             dynamicNFT,
		SellerID:        sellerID,
		StartingBid:     startingBid,
		CurrentBid:      startingBid,
		MinBidIncrement: startingBid / 10, // 10% минимальное повышение
		BuyoutPrice:     buyoutPrice,
		Status:          "active",
		StartTime:       time.Now(),
		EndTime:         time.Now().Add(duration),
		Bids:            []AuctionBid{},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Сохраняем аукцион
	err = am.saveAuction(ctx, auction)
	if err != nil {
		return nil, fmt.Errorf("failed to save auction: %w", err)
	}

	// Добавляем в активные аукционы
	am.activeAuctions[auction.AuctionID] = auction

	log.Printf("Auction created: ID %d, NFT %d, Starting bid %d", auction.AuctionID, nftID, startingBid)
	return auction, nil
}

// PlaceBid - размещение ставки
func (am *AuctionManager) PlaceBid(ctx context.Context, auctionID int64, bidderID int64, amount int64, isAutoBid bool, maxAutoBid *int64) (*AuctionBid, error) {
	auction, exists := am.activeAuctions[auctionID]
	if !exists {
		return nil, fmt.Errorf("auction not found")
	}

	if auction.Status != "active" {
		return nil, fmt.Errorf("auction is not active")
	}

	// Проверяем, что ставка не от продавца
	if bidderID == auction.SellerID {
		return nil, fmt.Errorf("seller cannot bid on own auction")
	}

	// Проверяем минимальное повышение
	if amount < auction.CurrentBid+auction.MinBidIncrement {
		return nil, fmt.Errorf("bid too low, minimum is %d", auction.CurrentBid+auction.MinBidIncrement)
	}

	// Проверяем баланс
	userState, err := am.db.GetUserState(ctx, bidderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user state: %w", err)
	}

	if userState.Balance < amount {
		return nil, fmt.Errorf("insufficient balance: need %d, have %d", amount, userState.Balance)
	}

	// Создаем ставку
	bid := AuctionBid{
		BidID:      time.Now().UnixNano(),
		AuctionID:  auctionID,
		BidderID:   bidderID,
		Amount:     amount,
		IsAutoBid:  isAutoBid,
		MaxAutoBid: maxAutoBid,
		CreatedAt:  time.Now(),
	}

	// Обновляем текущую ставку
	auction.CurrentBid = amount
	auction.Bids = append(auction.Bids, bid)
	auction.UpdatedAt = time.Now()

	// Сохраняем изменения
	err = am.saveAuction(ctx, auction)
	if err != nil {
		return nil, fmt.Errorf("failed to save auction: %w", err)
	}

	log.Printf("Bid placed: Auction %d, Bidder %d, Amount %d", auctionID, bidderID, amount)
	return &bid, nil
}

// Buyout - выкуп по цене
func (am *AuctionManager) Buyout(ctx context.Context, auctionID int64, buyerID int64) error {
	auction, exists := am.activeAuctions[auctionID]
	if !exists {
		return fmt.Errorf("auction not found")
	}

	if auction.Status != "active" {
		return fmt.Errorf("auction is not active")
	}

	if auction.BuyoutPrice == nil {
		return fmt.Errorf("auction does not have buyout price")
	}

	// Проверяем баланс
	userState, err := am.db.GetUserState(ctx, buyerID)
	if err != nil {
		return fmt.Errorf("failed to get user state: %w", err)
	}

	if userState.Balance < *auction.BuyoutPrice {
		return fmt.Errorf("insufficient balance: need %d, have %d", *auction.BuyoutPrice, userState.Balance)
	}

	// Завершаем аукцион
	auction.Status = "ended"
	auction.WinnerID = &buyerID
	auction.FinalPrice = auction.BuyoutPrice
	auction.UpdatedAt = time.Now()

	// Переводим NFT покупателю
	err = am.transferNFT(ctx, auction.NFTID, auction.SellerID, buyerID, *auction.BuyoutPrice)
	if err != nil {
		return fmt.Errorf("failed to transfer NFT: %w", err)
	}

	// Удаляем из активных
	delete(am.activeAuctions, auctionID)

	// Сохраняем изменения
	err = am.saveAuction(ctx, auction)
	if err != nil {
		return fmt.Errorf("failed to save auction: %w", err)
	}

	log.Printf("Auction bought out: ID %d, Buyer %d, Price %d", auctionID, buyerID, *auction.BuyoutPrice)
	return nil
}

// EndAuction - завершение аукциона
func (am *AuctionManager) EndAuction(ctx context.Context, auctionID int64) error {
	auction, exists := am.activeAuctions[auctionID]
	if !exists {
		return fmt.Errorf("auction not found")
	}

	if auction.Status != "active" {
		return fmt.Errorf("auction is not active")
	}

	auction.Status = "ended"
	auction.UpdatedAt = time.Now()

	// Определяем победителя
	if len(auction.Bids) > 0 {
		// Находим последнюю ставку
		lastBid := auction.Bids[len(auction.Bids)-1]
		auction.WinnerID = &lastBid.BidderID
		auction.FinalPrice = &lastBid.Amount

		// Переводим NFT победителю
		err := am.transferNFT(ctx, auction.NFTID, auction.SellerID, lastBid.BidderID, lastBid.Amount)
		if err != nil {
			return fmt.Errorf("failed to transfer NFT: %w", err)
		}
	} else {
		// Аукцион завершен без ставок
		log.Printf("Auction ended without bids: ID %d", auctionID)
	}

	// Удаляем из активных
	delete(am.activeAuctions, auctionID)

	// Сохраняем изменения
	err := am.saveAuction(ctx, auction)
	if err != nil {
		return fmt.Errorf("failed to save auction: %w", err)
	}

	log.Printf("Auction ended: ID %d, Winner %v, Final price %v", auctionID, auction.WinnerID, auction.FinalPrice)
	return nil
}

// transferNFT - передача NFT
func (am *AuctionManager) transferNFT(ctx context.Context, nftID int64, fromID int64, toID int64, price int64) error {
	// Списываем монеты у покупателя
	err := am.db.UpdateUserBalance(ctx, toID, -price)
	if err != nil {
		return fmt.Errorf("failed to deduct buyer balance: %w", err)
	}

	// Начисляем монеты продавцу (минус комиссия)
	commission := int64(float64(price) * 0.05) // 5% комиссия
	err = am.db.UpdateUserBalance(ctx, fromID, price-commission)
	if err != nil {
		// Возвращаем монеты покупателю при ошибке
		am.db.UpdateUserBalance(ctx, toID, price)
		return fmt.Errorf("failed to credit seller balance: %w", err)
	}

	// Обновляем владельца NFT
	// В реальном приложении здесь будет обновление в БД
	log.Printf("NFT transferred: ID %d, From %d, To %d, Price %d", nftID, fromID, toID, price)

	return nil
}

// processAuctions - обработка активных аукционов
func (am *AuctionManager) processAuctions() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		am.checkExpiredAuctions()
	}
}

// checkExpiredAuctions - проверка истекших аукционов
func (am *AuctionManager) checkExpiredAuctions() {
	now := time.Now()
	for auctionID, auction := range am.activeAuctions {
		if now.After(auction.EndTime) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			err := am.EndAuction(ctx, auctionID)
			cancel()
			
			if err != nil {
				log.Printf("Failed to end auction %d: %v", auctionID, err)
			}
		}
	}
}

// GetActiveAuctions - получение активных аукционов
func (am *AuctionManager) GetActiveAuctions(ctx context.Context, limit int) ([]NFTAuction, error) {
	auctions := make([]NFTAuction, 0, limit)
	
	for _, auction := range am.activeAuctions {
		if len(auctions) >= limit {
			break
		}
		auctions = append(auctions, *auction)
	}

	return auctions, nil
}

// GetAuction - получение аукциона по ID
func (am *AuctionManager) GetAuction(ctx context.Context, auctionID int64) (*NFTAuction, error) {
	auction, exists := am.activeAuctions[auctionID]
	if !exists {
		// Проверяем в базе данных для завершенных аукционов
		return am.getAuctionFromDB(ctx, auctionID)
	}

	return auction, nil
}

// getAuctionFromDB - получение аукциона из базы данных
func (am *AuctionManager) getAuctionFromDB(ctx context.Context, auctionID int64) (*NFTAuction, error) {
	// В реальном приложении здесь будет запрос к БД
	// Для примера возвращаем nil
	return nil, fmt.Errorf("auction not found")
}

// GetUserAuctions - получение аукционов пользователя
func (am *AuctionManager) GetUserAuctions(ctx context.Context, userID int64, status string) ([]NFTAuction, error) {
	var userAuctions []NFTAuction
	
	for _, auction := range am.activeAuctions {
		if (status == "selling" && auction.SellerID == userID) ||
		   (status == "bidding" && am.hasUserBid(auction, userID)) {
			userAuctions = append(userAuctions, *auction)
		}
	}

	return userAuctions, nil
}

// hasUserBid - проверка, делал ли пользователь ставку
func (am *AuctionManager) hasUserBid(auction *NFTAuction, userID int64) bool {
	for _, bid := range auction.Bids {
		if bid.BidderID == userID {
			return true
		}
	}
	return false
}

// CancelAuction - отмена аукциона
func (am *AuctionManager) CancelAuction(ctx context.Context, auctionID int64, sellerID int64) error {
	auction, exists := am.activeAuctions[auctionID]
	if !exists {
		return fmt.Errorf("auction not found")
	}

	if auction.SellerID != sellerID {
		return fmt.Errorf("only seller can cancel auction")
	}

	if auction.Status != "active" {
		return fmt.Errorf("auction is not active")
	}

	// Проверяем, есть ли ставки
	if len(auction.Bids) > 0 {
		return fmt.Errorf("cannot cancel auction with bids")
	}

	auction.Status = "cancelled"
	auction.UpdatedAt = time.Now()

	// Удаляем из активных
	delete(am.activeAuctions, auctionID)

	// Сохраняем изменения
	err := am.saveAuction(ctx, auction)
	if err != nil {
		return fmt.Errorf("failed to save auction: %w", err)
	}

	log.Printf("Auction cancelled: ID %d", auctionID)
	return nil
}

// saveAuction - сохранение аукциона
func (am *AuctionManager) saveAuction(ctx context.Context, auction *NFTAuction) error {
	// Здесь должна быть реальная запись в базу данных
	// Для примера просто логируем
	log.Printf("Saving auction: ID %d, Status %s, Current bid %d", auction.AuctionID, auction.Status, auction.CurrentBid)
	return nil
}

// GetAuctionStats - получение статистики аукционов
func (am *AuctionManager) GetAuctionStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	stats["active_auctions"] = len(am.activeAuctions)
	
	totalValue := int64(0)
	for _, auction := range am.activeAuctions {
		totalValue += auction.CurrentBid
	}
	stats["total_active_value"] = totalValue
	
	// В реальном приложении здесь будет запрос к БД для полной статистики
	stats["total_auctions"] = 1000
	stats["completed_auctions"] = 850
	stats["total_volume"] = int64(10000000)
	
	return stats, nil
}

// toJSON - конвертация в JSON
func (am *AuctionManager) toJSON(data interface{}) string {
	bytes, _ := json.Marshal(data)
	return string(bytes)
}
