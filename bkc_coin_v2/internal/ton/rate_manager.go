package ton

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// RateManager управляет курсами TON/USD
type RateManager struct {
	apiKey      string
	client      *http.Client
	currentRate float64
	lastUpdate  time.Time
}

// CoinGecko API response
type CoinGeckoResponse struct {
	Ton struct {
		Usd float64 `json:"usd"`
	} `json:"ton"`
}

// NewRateManager создает новый менеджер курсов
func NewRateManager() *RateManager {
	return &RateManager{
		apiKey:      "", // CoinGecko не требует API ключа
		client:      &http.Client{Timeout: 30 * time.Second},
		currentRate: 5.0, // Значение по умолчанию
		lastUpdate:  time.Now(),
	}
}

// GetTONRate получает текущий курс TON/USD
func (rm *RateManager) GetTONRate() (float64, error) {
	// Обновляем курс если прошло более 5 минут
	if time.Since(rm.lastUpdate) > 5*time.Minute {
		err := rm.updateTONRate()
		if err != nil {
			log.Printf("Ошибка обновления курса TON: %v", err)
			return rm.currentRate, err
		}
	}

	return rm.currentRate, nil
}

// SetTONRate устанавливает курс вручную (для админа)
func (rm *RateManager) SetTONRate(rate float64) error {
	if rate <= 0 {
		return fmt.Errorf("курс должен быть положительным")
	}

	rm.currentRate = rate
	rm.lastUpdate = time.Now()

	log.Printf("Курс TON/USD установлен вручную: %.2f", rate)
	return nil
}

// updateTONRate обновляет курс с CoinGecko API
func (rm *RateManager) updateTONRate() error {
	url := "https://api.coingecko.com/api/v3/simple/price?ids=the-open-network&vs_currencies=usd"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := rm.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("API вернул статус: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var response CoinGeckoResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	if response.Ton.Usd <= 0 {
		return fmt.Errorf("невалидный курс от API")
	}

	rm.currentRate = response.Ton.Usd
	rm.lastUpdate = time.Now()

	log.Printf("Курс TON/USD обновлен: %.6f", rm.currentRate)
	return nil
}

// ConvertTONtoBKC конвертирует TON в BKC
func (rm *RateManager) ConvertTONtoBKC(tonAmount float64) float64 {
	// 1 BKC = $0.001
	// tonAmount USD = tonAmount * currentRate
	// bkcAmount = tonAmount USD / 0.001
	return tonAmount * rm.currentRate / 0.001
}

// ConvertBKCtoTON конвертирует BKC в TON
func (rm *RateManager) ConvertBKCtoTON(bkcAmount float64) float64 {
	// 1 BKC = $0.001
	// tonAmount = bkcAmount USD / currentRate
	return (bkcAmount * 0.001) / rm.currentRate
}

// GetCurrentRatesInfo возвращает информацию о текущих курсах
func (rm *RateManager) GetCurrentRatesInfo() map[string]interface{} {
	return map[string]interface{}{
		"ton_usd":     rm.currentRate,
		"bkc_usd":     0.001,
		"ton_bkc":     rm.ConvertTONtoBKC(1),
		"bkc_ton":     rm.ConvertBKCtoTON(1),
		"last_update": rm.lastUpdate.Format("2006-01-02 15:04:05"),
		"source":      "coingecko",
	}
}

// GetRateHistory возвращает историю изменений курса (можно расширить)
func (rm *RateManager) GetRateHistory() []map[string]interface{} {
	// Здесь можно добавить хранение истории в базе данных
	return []map[string]interface{}{
		{
			"rate":      rm.currentRate,
			"timestamp": rm.lastUpdate.Unix(),
			"source":    "coingecko",
		},
	}
}
