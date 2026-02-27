package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
)

type DatabaseConfig struct {
	RenderServers []string          `json:"render_servers"`
	Databases    map[string][]string `json:"databases"`
	Sharding     ShardingConfig    `json:"sharding_config"`
}

type ShardingConfig struct {
	TursoShards     int `json:"turso_shards"`
	NeonShards      int `json:"neon_shards"`
	SupabaseShards  int `json:"supabase_shards"`
	PlanetScaleShards int `json:"planetscale_shards"`
	UpstashShards   int `json:"upstash_shards"`
}

var globalConfig *DatabaseConfig

func loadConfig() error {
	configFile := "config/databases.json"
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read database config: %w", err)
	}

	var config DatabaseConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse database config: %w", err)
	}

	globalConfig = &config
	return nil
}

func saveConfig() error {
	data, err := json.MarshalIndent(globalConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile("config/databases.json", data, 0644)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: scale_manager <command> [args...]")
		fmt.Println("Commands:")
		fmt.Println("  init                    - Initialize database manager")
		fmt.Println("  add-server <url>       - Add new Render server")
		fmt.Println("  add-db <type> <url>    - Add new database")
		fmt.Println("  status                  - Show current status")
		fmt.Println("  scale-servers <count>   - Scale servers to count")
		fmt.Println("  scale-db <type> <count> - Scale database type to count")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		if err := loadConfig(); err != nil {
			log.Fatal("Failed to initialize database manager:", err)
		}
		fmt.Println("Database manager initialized successfully")

	case "add-server":
		if len(os.Args) < 3 {
			log.Fatal("Usage: add-server <url>")
		}
		if globalConfig == nil {
			log.Fatal("Database manager not initialized. Run 'init' first.")
		}
		serverURL := os.Args[2]
		globalConfig.RenderServers = append(globalConfig.RenderServers, serverURL)
		if err := saveConfig(); err != nil {
			log.Fatal("Failed to save config:", err)
		}
		fmt.Printf("Added server: %s\n", serverURL)

	case "add-db":
		if len(os.Args) < 4 {
			log.Fatal("Usage: add-db <type> <url>")
		}
		if globalConfig == nil {
			log.Fatal("Database manager not initialized. Run 'init' first.")
		}
		dbType := os.Args[2]
		connectionURL := os.Args[3]
		if globalConfig.Databases[dbType] == nil {
			globalConfig.Databases[dbType] = []string{}
		}
		globalConfig.Databases[dbType] = append(globalConfig.Databases[dbType], connectionURL)
		if err := saveConfig(); err != nil {
			log.Fatal("Failed to save config:", err)
		}
		fmt.Printf("Added %s database: %s\n", dbType, connectionURL)

	case "status":
		if globalConfig == nil {
			log.Fatal("Database manager not initialized. Run 'init' first.")
		}
		fmt.Printf("Render Servers: %d\n", len(globalConfig.RenderServers))
		
		totalDatabases := 0
		for _, dbs := range globalConfig.Databases {
			totalDatabases += len(dbs)
		}
		fmt.Printf("Total Databases: %d\n", totalDatabases)
		fmt.Printf("Turso Shards: %d\n", globalConfig.Sharding.TursoShards)
		fmt.Printf("Neon Shards: %d\n", globalConfig.Sharding.NeonShards)
		fmt.Printf("Supabase Shards: %d\n", globalConfig.Sharding.SupabaseShards)
		fmt.Printf("PlanetScale Shards: %d\n", globalConfig.Sharding.PlanetScaleShards)
		fmt.Printf("Upstash Shards: %d\n", globalConfig.Sharding.UpstashShards)

	case "scale-servers":
		if len(os.Args) < 3 {
			log.Fatal("Usage: scale-servers <count>")
		}
		if globalConfig == nil {
			log.Fatal("Database manager not initialized. Run 'init' first.")
		}
		count, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatal("Invalid server count:", err)
		}
		
		for i := 1; i <= count; i++ {
			serverURL := fmt.Sprintf("https://app-%d.onrender.com", i)
			globalConfig.RenderServers = append(globalConfig.RenderServers, serverURL)
			fmt.Printf("Added server %d: %s\n", i, serverURL)
		}
		
		if err := saveConfig(); err != nil {
			log.Fatal("Failed to save config:", err)
		}

	case "scale-db":
		if len(os.Args) < 4 {
			log.Fatal("Usage: scale-db <type> <count>")
		}
		if globalConfig == nil {
			log.Fatal("Database manager not initialized. Run 'init' first.")
		}
		dbType := os.Args[2]
		count, err := strconv.Atoi(os.Args[3])
		if err != nil {
			log.Fatal("Invalid database count:", err)
		}
		
		if globalConfig.Databases[dbType] == nil {
			globalConfig.Databases[dbType] = []string{}
		}
		
		for i := 1; i <= count; i++ {
			var connectionURL string
			switch dbType {
			case "turso":
				connectionURL = fmt.Sprintf("turso://db-%d.turso.io", i)
			case "neon":
				connectionURL = fmt.Sprintf("postgresql://neon-%d.ep.io/db", i)
			case "supabase":
				connectionURL = fmt.Sprintf("postgresql://supabase-%d.supabase.co", i)
			case "planetscale":
				connectionURL = fmt.Sprintf("mysql://planetscale-%d.planetscale.com/db", i)
			case "upstash":
				connectionURL = fmt.Sprintf("redis://upstash-%d.upstash.io", i)
			default:
				log.Printf("Unknown database type: %s", dbType)
				continue
			}
			
			globalConfig.Databases[dbType] = append(globalConfig.Databases[dbType], connectionURL)
			fmt.Printf("Added %s database %d: %s\n", dbType, i, connectionURL)
		}
		
		if err := saveConfig(); err != nil {
			log.Fatal("Failed to save config:", err)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
