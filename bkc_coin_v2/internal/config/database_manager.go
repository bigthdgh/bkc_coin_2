package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type DatabaseConfig struct {
	RenderServers []string            `json:"render_servers"`
	Databases     map[string][]string `json:"databases"`
	Sharding      ShardingConfig      `json:"sharding_config"`
}

type ShardingConfig struct {
	TursoShards       int `json:"turso_shards"`
	NeonShards        int `json:"neon_shards"`
	SupabaseShards    int `json:"supabase_shards"`
	PlanetScaleShards int `json:"planetscale_shards"`
	UpstashShards     int `json:"upstash_shards"`
}

type DatabaseManager struct {
	config *DatabaseConfig
	mu     sync.RWMutex
}

// GetShardingConfig возвращает конфигурацию шардинга
func (dm *DatabaseManager) GetShardingConfig() *ShardingConfig {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return &dm.config.Sharding
}

var GlobalDBManager *DatabaseManager

func InitDatabaseManager() error {
	configFile := "config/databases.json"
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read database config: %w", err)
	}

	var config DatabaseConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse database config: %w", err)
	}

	GlobalDBManager = &DatabaseManager{
		config: &config,
	}

	return nil
}

// AddRenderServer добавляет новый Render сервер
func (dm *DatabaseManager) AddRenderServer(serverURL string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.config.RenderServers = append(dm.config.RenderServers, serverURL)
	return dm.saveConfig()
}

// AddDatabase добавляет новую базу данных
func (dm *DatabaseManager) AddDatabase(dbType, connectionURL string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if dm.config.Databases[dbType] == nil {
		dm.config.Databases[dbType] = []string{}
	}

	dm.config.Databases[dbType] = append(dm.config.Databases[dbType], connectionURL)
	return dm.saveConfig()
}

// GetRenderServer получает сервер по round-robin
func (dm *DatabaseManager) GetRenderServer(userID int64) string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	servers := dm.config.RenderServers
	if len(servers) == 0 {
		return ""
	}

	// Round-robin по userID
	index := int(userID % int64(len(servers)))
	return servers[index]
}

// GetTursoDB получает Turso базу по шардированию
func (dm *DatabaseManager) GetTursoDB(userID int64) string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	dbs := dm.config.Databases["turso"]
	if len(dbs) == 0 {
		return ""
	}

	shards := dm.config.Sharding.TursoShards
	index := int(userID % int64(shards))
	return dbs[index]
}

// GetNeonDB получает Neon базу по шардированию
func (dm *DatabaseManager) GetNeonDB(userID int64) string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	dbs := dm.config.Databases["neon"]
	if len(dbs) == 0 {
		return ""
	}

	shards := dm.config.Sharding.NeonShards
	index := int(userID % int64(shards))
	return dbs[index]
}

// GetSupabaseDB получает Supabase базу по шардированию
func (dm *DatabaseManager) GetSupabaseDB(userID int64) string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	dbs := dm.config.Databases["supabase"]
	if len(dbs) == 0 {
		return ""
	}

	shards := dm.config.Sharding.SupabaseShards
	index := int(userID % int64(shards))
	return dbs[index]
}

// GetPlanetScaleDB получает PlanetScale базу по шардированию
func (dm *DatabaseManager) GetPlanetScaleDB(userID int64) string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	dbs := dm.config.Databases["planetscale"]
	if len(dbs) == 0 {
		return ""
	}

	shards := dm.config.Sharding.PlanetScaleShards
	index := int(userID % int64(shards))
	return dbs[index]
}

// GetUpstashRedis получает Redis по round-robin
func (dm *DatabaseManager) GetUpstashRedis(userID int64) string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	dbs := dm.config.Databases["upstash"]
	if len(dbs) == 0 {
		return ""
	}

	shards := dm.config.Sharding.UpstashShards
	index := int(userID % int64(shards))
	return dbs[index]
}

// GetTotalServers возвращает общее количество серверов
func (dm *DatabaseManager) GetTotalServers() int {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	return len(dm.config.RenderServers)
}

// GetTotalDatabases возвращает общее количество баз данных
func (dm *DatabaseManager) GetTotalDatabases() int {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	total := 0
	for _, dbs := range dm.config.Databases {
		total += len(dbs)
	}
	return total
}

// GetTursoDBs возвращает все Turso базы
func (dm *DatabaseManager) GetTursoDBs() []string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.config.Databases["turso"]
}

// GetNeonDBs возвращает все Neon базы
func (dm *DatabaseManager) GetNeonDBs() []string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.config.Databases["neon"]
}

// GetSupabaseDBs возвращает все Supabase базы
func (dm *DatabaseManager) GetSupabaseDBs() []string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.config.Databases["supabase"]
}

// GetPlanetScaleDBs возвращает все PlanetScale базы
func (dm *DatabaseManager) GetPlanetScaleDBs() []string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.config.Databases["planetscale"]
}

// GetUpstashRedisList возвращает все Redis базы
func (dm *DatabaseManager) GetUpstashRedisList() []string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.config.Databases["upstash"]
}

// GetCockroachDBs возвращает все CockroachDB базы
func (dm *DatabaseManager) GetCockroachDBs() []string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.config.Databases["cockroachdb"]
}

// saveConfig сохраняет конфигурацию в файл
func (dm *DatabaseManager) saveConfig() error {
	data, err := json.MarshalIndent(dm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile("config/databases.json", data, 0644)
}
