package nft

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"bkc_coin_v2/internal/database"
)

// DynamicNFT - динамический NFT с уровнями и прогрессом
type DynamicNFT struct {
	ID              int64             `json:"id"`
	BaseNFT         NFTItem           `json:"base_nft"`
	Level           int               `json:"level"`
	Experience      int64             `json:"experience"`
	MaxExperience   int64             `json:"max_experience"`
	Abilities       []Ability         `json:"abilities"`
	UpgradeHistory  []Upgrade         `json:"upgrade_history"`
	OwnerID         int64             `json:"owner_id"`
	Branding        *NFTBranding      `json:"branding,omitempty"`
	Stats           NFTStats          `json:"stats"`
	CreatedAt       time.Time         `json:"created_at"`
	LastUpdated     time.Time         `json:"last_updated"`
}

// NFTItem - базовый NFT
type NFTItem struct {
	NFTID       int64     `json:"nft_id"`
	Title       string    `json:"title"`
	ImageURL    string    `json:"image_url"`
	PriceCoins  int64     `json:"price_coins"`
	SupplyLeft  int64     `json:"supply_left"`
	CreatedAt   time.Time `json:"created_at"`
}

// Ability - способность NFT
type Ability struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Level       int     `json:"level"`
	Effect      float64 `json:"effect"`
	UnlockedAt  int     `json:"unlocked_at"`
}

// Upgrade - улучшение NFT
type Upgrade struct {
	ID         int64     `json:"id"`
	Type       string    `json:"type"`
	FromLevel  int       `json:"from_level"`
	ToLevel    int       `json:"to_level"`
	Cost       int64     `json:"cost"`
	Timestamp  time.Time `json:"timestamp"`
	NewAbility *Ability  `json:"new_ability,omitempty"`
}

// NFTBranding - брендирование NFT
type NFTBranding struct {
	BrandName        string    `json:"brand_name"`
	BrandLogo        string    `json:"brand_logo"`
	Collaboration    bool      `json:"collaboration"`
	SpecialRewards   []Reward  `json:"special_rewards"`
	BrandColor       string    `json:"brand_color"`
	BrandDescription string    `json:"brand_description"`
	CreatedAt        time.Time `json:"created_at"`
}

// Reward - награда от бренда
type Reward struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Value       int64  `json:"value"`
	Requirement string `json:"requirement"`
}

// NFTStats - статистика NFT
type NFTStats struct {
	TimesUsed        int64     `json:"times_used"`
	TotalEarned      int64     `json:"total_earned"`
	BattlesWon       int64     `json:"battles_won"`
	BattlesLost      int64     `json:"battles_lost"`
	LastUsed         time.Time `json:"last_used"`
	PopularityScore  float64   `json:"popularity_score"`
	RarityScore      float64   `json:"rarity_score"`
}

// DynamicNFTManager - менеджер динамических NFT
type DynamicNFTManager struct {
	db *database.UnifiedDB
}

// NewDynamicNFTManager - создание менеджера
func NewDynamicNFTManager(db *database.UnifiedDB) *DynamicNFTManager {
	return &DynamicNFTManager{db: db}
}

// CreateDynamicNFT - создание динамического NFT
func (dnm *DynamicNFTManager) CreateDynamicNFT(ctx context.Context, baseNFT NFTItem, ownerID int64, branding *NFTBranding) (*DynamicNFT, error) {
	dynamicNFT := &DynamicNFT{
		BaseNFT:        baseNFT,
		Level:          1,
		Experience:     0,
		MaxExperience:  1000,
		Abilities:      dnm.generateInitialAbilities(),
		UpgradeHistory: []Upgrade{},
		OwnerID:        ownerID,
		Branding:       branding,
		Stats: NFTStats{
			TimesUsed:       0,
			TotalEarned:     0,
			BattlesWon:      0,
			BattlesLost:     0,
			PopularityScore: 0.0,
			RarityScore:     dnm.calculateRarityScore(baseNFT),
		},
		CreatedAt:   time.Now(),
		LastUpdated: time.Now(),
	}

	// Сохраняем в базу данных
	err := dnm.saveDynamicNFT(ctx, dynamicNFT)
	if err != nil {
		return nil, fmt.Errorf("failed to save dynamic NFT: %w", err)
	}

	log.Printf("Dynamic NFT created: ID %d, Owner %d, Level %d", dynamicNFT.ID, ownerID, dynamicNFT.Level)
	return dynamicNFT, nil
}

// generateInitialAbilities - генерация начальных способностей
func (dnm *DynamicNFTManager) generateInitialAbilities() []Ability {
	return []Ability{
		{
			ID:          "basic_boost",
			Name:        "Basic Boost",
			Description: "Increases tap rewards by 5%",
			Level:       1,
			Effect:      0.05,
			UnlockedAt:  1,
		},
		{
			ID:          "energy_efficiency",
			Name:        "Energy Efficiency",
			Description: "Reduces energy cost by 3%",
			Level:       1,
			Effect:      0.03,
			UnlockedAt:  1,
		},
	}
}

// calculateRarityScore - расчет рейтинга редкости
func (dnm *DynamicNFTManager) calculateRarityScore(nft NFTItem) float64 {
	// Базовый расчет на основе цены и доступности
	baseScore := float64(nft.PriceCoins) / 10000.0
	supplyBonus := (1000.0 - float64(nft.SupplyLeft)) / 1000.0
	
	return baseScore + supplyBonus
}

// AddExperience - добавление опыта NFT
func (dnm *DynamicNFTManager) AddExperience(ctx context.Context, nftID int64, exp int64) (*DynamicNFT, error) {
	nft, err := dnm.GetDynamicNFT(ctx, nftID)
	if err != nil {
		return nil, fmt.Errorf("failed to get NFT: %w", err)
	}

	nft.Experience += exp
	nft.LastUpdated = time.Now()

	// Проверка на повышение уровня
	for nft.Experience >= nft.MaxExperience {
		err = dnm.levelUp(ctx, nft)
		if err != nil {
			return nil, fmt.Errorf("failed to level up: %w", err)
		}
	}

	// Обновляем статистику
	nft.Stats.TimesUsed++
	nft.Stats.PopularityScore = dnm.calculatePopularityScore(nft)

	// Сохраняем изменения
	err = dnm.saveDynamicNFT(ctx, nft)
	if err != nil {
		return nil, fmt.Errorf("failed to save NFT: %w", err)
	}

	log.Printf("Experience added: NFT %d, +%d exp, Level %d", nftID, exp, nft.Level)
	return nft, nil
}

// levelUp - повышение уровня NFT
func (dnm *DynamicNFTManager) levelUp(ctx context.Context, nft *DynamicNFT) error {
	oldLevel := nft.Level
	nft.Level++
	nft.Experience -= nft.MaxExperience
	nft.MaxExperience = int64(float64(nft.MaxExperience) * 1.5)

	// Разблокируем новые способности
	newAbility := dnm.unlockAbility(nft.Level)
	if newAbility != nil {
		nft.Abilities = append(nft.Abilities, *newAbility)
	}

	// Записываем в историю улучшений
	upgrade := Upgrade{
		ID:        time.Now().UnixNano(),
		Type:      "level_up",
		FromLevel: oldLevel,
		ToLevel:   nft.Level,
		Cost:      0,
		Timestamp: time.Now(),
	}
	if newAbility != nil {
		upgrade.NewAbility = newAbility
	}
	nft.UpgradeHistory = append(nft.UpgradeHistory, upgrade)

	log.Printf("NFT leveled up: ID %d, Level %d -> %d", nft.ID, oldLevel, nft.Level)
	return nil
}

// unlockAbility - разблокировка новой способности
func (dnm *DynamicNFTManager) unlockAbility(level int) *Ability {
	abilities := map[int]Ability{
		5: {
			ID:          "super_boost",
			Name:        "Super Boost",
			Description: "Increases tap rewards by 15%",
			Level:       1,
			Effect:      0.15,
			UnlockedAt:  5,
		},
		10: {
			ID:          "master_efficiency",
			Name:        "Master Efficiency",
			Description: "Reduces energy cost by 10%",
			Level:       1,
			Effect:      0.10,
			UnlockedAt:  10,
		},
		15: {
			ID:          "legendary_power",
			Name:        "Legendary Power",
			Description: "Increases all rewards by 25%",
			Level:       1,
			Effect:      0.25,
			UnlockedAt:  15,
		},
	}

	if ability, exists := abilities[level]; exists {
		return &ability
	}
	return nil
}

// calculatePopularityScore - расчет популярности
func (dnm *DynamicNFTManager) calculatePopularityScore(nft *DynamicNFT) float64 {
	usageScore := float64(nft.Stats.TimesUsed) / 100.0
	levelScore := float64(nft.Level) * 0.1
	rarityScore := nft.Stats.RarityScore * 0.3
	
	return usageScore + levelScore + rarityScore
}

// GetDynamicNFT - получение динамического NFT
func (dnm *DynamicNFTManager) GetDynamicNFT(ctx context.Context, nftID int64) (*DynamicNFT, error) {
	// В реальном приложении здесь будет запрос к БД
	// Для примера возвращаем тестовые данные
	nft := &DynamicNFT{
		ID: nftID,
		BaseNFT: NFTItem{
			NFTID:      nftID,
			Title:      "Dynamic NFT #" + fmt.Sprintf("%d", nftID),
			ImageURL:   fmt.Sprintf("/images/nft_%d.png", nftID),
			PriceCoins: 10000,
			SupplyLeft: 100,
			CreatedAt:  time.Now().Add(-24 * time.Hour),
		},
		Level:          5,
		Experience:     2500,
		MaxExperience:  5000,
		Abilities:      dnm.generateInitialAbilities(),
		UpgradeHistory: []Upgrade{},
		OwnerID:        12345,
		Stats: NFTStats{
			TimesUsed:       150,
			TotalEarned:     50000,
			BattlesWon:      25,
			BattlesLost:     5,
			LastUsed:        time.Now().Add(-1 * time.Hour),
			PopularityScore: 7.5,
			RarityScore:     8.2,
		},
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		LastUpdated: time.Now(),
	}

	return nft, nil
}

// saveDynamicNFT - сохранение динамического NFT
func (dnm *DynamicNFTManager) saveDynamicNFT(ctx context.Context, nft *DynamicNFT) error {
	// Здесь должна быть реальная запись в базу данных
	// Для примера просто логируем
	log.Printf("Saving dynamic NFT: ID %d, Level %d, Exp %d", nft.ID, nft.Level, nft.Experience)
	return nil
}

// GetUserDynamicNFTs - получение всех динамических NFT пользователя
func (dnm *DynamicNFTManager) GetUserDynamicNFTs(ctx context.Context, userID int64) ([]DynamicNFT, error) {
	// В реальном приложении здесь будет запрос к БД
	// Для примера возвращаем тестовые данные
	nfts := make([]DynamicNFT, 0, 3)
	
	for i := 1; i <= 3; i++ {
		nft, err := dnm.GetDynamicNFT(ctx, int64(i))
		if err != nil {
			continue
		}
		nft.OwnerID = userID
		nfts = append(nfts, *nft)
	}

	return nfts, nil
}

// UpgradeNFT - улучшение NFT за монеты
func (dnm *DynamicNFTManager) UpgradeNFT(ctx context.Context, nftID int64, upgradeType string, cost int64) (*DynamicNFT, error) {
	nft, err := dnm.GetDynamicNFT(ctx, nftID)
	if err != nil {
		return nil, fmt.Errorf("failed to get NFT: %w", err)
	}

	// Проверяем баланс пользователя
	userState, err := dnm.db.GetUserState(ctx, nft.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user state: %w", err)
	}

	if userState.Balance < cost {
		return nil, fmt.Errorf("insufficient balance: need %d, have %d", cost, userState.Balance)
	}

	// Списываем монеты
	err = dnm.db.UpdateUserBalance(ctx, nft.OwnerID, -cost)
	if err != nil {
		return nil, fmt.Errorf("failed to deduct balance: %w", err)
	}

	// Применяем улучшение
	switch upgradeType {
	case "instant_level":
		err = dnm.levelUp(ctx, nft)
		if err != nil {
			// Возвращаем монеты при ошибке
			dnm.db.UpdateUserBalance(ctx, nft.OwnerID, cost)
			return nil, fmt.Errorf("failed to level up: %w", err)
		}
	case "boost_experience":
		nft.Experience += nft.MaxExperience / 2
	case "unlock_ability":
		newAbility := dnm.unlockAbility(nft.Level + 1)
		if newAbility != nil {
			nft.Abilities = append(nft.Abilities, *newAbility)
		}
	}

	// Записываем в историю
	upgrade := Upgrade{
		ID:        time.Now().UnixNano(),
		Type:      upgradeType,
		FromLevel: nft.Level,
		ToLevel:   nft.Level,
		Cost:      cost,
		Timestamp: time.Now(),
	}
	nft.UpgradeHistory = append(nft.UpgradeHistory, upgrade)

	// Сохраняем изменения
	err = dnm.saveDynamicNFT(ctx, nft)
	if err != nil {
		return nil, fmt.Errorf("failed to save NFT: %w", err)
	}

	log.Printf("NFT upgraded: ID %d, Type %s, Cost %d", nftID, upgradeType, cost)
	return nft, nil
}

// ApplyBranding - применение брендирования к NFT
func (dnm *DynamicNFTManager) ApplyBranding(ctx context.Context, nftID int64, branding NFTBranding) (*DynamicNFT, error) {
	nft, err := dnm.GetDynamicNFT(ctx, nftID)
	if err != nil {
		return nil, fmt.Errorf("failed to get NFT: %w", err)
	}

	nft.Branding = &branding
	nft.LastUpdated = time.Now()

	// Добавляем специальные награды
	if len(branding.SpecialRewards) > 0 {
		for _, reward := range branding.SpecialRewards {
			// Применяем награды к NFT
			dnm.applyReward(ctx, nft, reward)
		}
	}

	// Сохраняем изменения
	err = dnm.saveDynamicNFT(ctx, nft)
	if err != nil {
		return nil, fmt.Errorf("failed to save NFT: %w", err)
	}

	log.Printf("Branding applied: NFT %d, Brand %s", nftID, branding.BrandName)
	return nft, nil
}

// applyReward - применение награды к NFT
func (dnm *DynamicNFTManager) applyReward(ctx context.Context, nft *DynamicNFT, reward Reward) {
	switch reward.Type {
	case "experience_boost":
		nft.Experience += reward.Value
	case "ability_boost":
		for i, ability := range nft.Abilities {
			if ability.ID == reward.Requirement {
				nft.Abilities[i].Effect += float64(reward.Value) / 100.0
				break
			}
		}
	case "stats_boost":
		nft.Stats.RarityScore += float64(reward.Value) / 10.0
	}
}

// GetTopNFTs - получение топ NFT по популярности
func (dnm *DynamicNFTManager) GetTopNFTs(ctx context.Context, limit int) ([]DynamicNFT, error) {
	// В реальном приложении здесь будет запрос к БД
	// Для примера возвращаем тестовые данные
	topNFTs := make([]DynamicNFT, 0, limit)
	
	for i := 1; i <= limit; i++ {
		nft, err := dnm.GetDynamicNFT(ctx, int64(i))
		if err != nil {
			continue
		}
		// Увеличиваем популярность для топ NFT
		nft.Stats.PopularityScore = float64(10-i) * 1.5
		topNFTs = append(topNFTs, *nft)
	}

	return topNFTs, nil
}

// CalculateNFTValue - расчет стоимости NFT
func (dnm *DynamicNFTManager) CalculateNFTValue(nft *DynamicNFT) int64 {
	baseValue := nft.BaseNFT.PriceCoins
	levelMultiplier := 1.0 + float64(nft.Level-1)*0.2
	abilityMultiplier := 1.0
	for _, ability := range nft.Abilities {
		abilityMultiplier += ability.Effect
	}
	popularityMultiplier := 1.0 + nft.Stats.PopularityScore/10.0
	
	value := float64(baseValue) * levelMultiplier * abilityMultiplier * popularityMultiplier
	
	return int64(value)
}

// GetNFTProgress - получение прогресса NFT
func (dnm *DynamicNFTManager) GetNFTProgress(nft *DynamicNFT) map[string]interface{} {
	progress := make(map[string]interface{})
	
	progress["current_level"] = nft.Level
	progress["current_experience"] = nft.Experience
	progress["max_experience"] = nft.MaxExperience
	progress["experience_percentage"] = float64(nft.Experience) / float64(nft.MaxExperience) * 100
	progress["abilities_count"] = len(nft.Abilities)
	progress["upgrades_count"] = len(nft.UpgradeHistory)
	progress["total_value"] = dnm.CalculateNFTValue(nft)
	progress["popularity_score"] = nft.Stats.PopularityScore
	progress["rarity_score"] = nft.Stats.RarityScore
	
	return progress
}

// toJSON - конвертация в JSON
func (dnm *DynamicNFTManager) toJSON(data interface{}) string {
	bytes, _ := json.Marshal(data)
	return string(bytes)
}
