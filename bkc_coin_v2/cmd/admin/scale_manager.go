package main

import (
	"bkc_coin_v2/internal/config"
	"fmt"
	"log"
	"os"
	"strconv"
)

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
		if err := config.InitDatabaseManager(); err != nil {
			log.Fatal("Failed to initialize database manager:", err)
		}
		fmt.Println("Database manager initialized successfully")

	case "add-server":
		if len(os.Args) < 3 {
			log.Fatal("Usage: add-server <url>")
		}
		serverURL := os.Args[2]
		if err := config.GlobalDBManager.AddRenderServer(serverURL); err != nil {
			log.Fatal("Failed to add server:", err)
		}
		fmt.Printf("Added server: %s\n", serverURL)

	case "add-db":
		if len(os.Args) < 4 {
			log.Fatal("Usage: add-db <type> <url>")
		}
		dbType := os.Args[2]
		connectionURL := os.Args[3]
		if err := config.GlobalDBManager.AddDatabase(dbType, connectionURL); err != nil {
			log.Fatal("Failed to add database:", err)
		}
		fmt.Printf("Added %s database: %s\n", dbType, connectionURL)

	case "status":
		fmt.Printf("Render Servers: %d\n", config.GlobalDBManager.GetTotalServers())
		fmt.Printf("Total Databases: %d\n", config.GlobalDBManager.GetTotalDatabases())
		fmt.Printf("Turso Shards: %d\n", config.GlobalDBManager.GetShardingConfig().TursoShards)
		fmt.Printf("Neon Shards: %d\n", config.GlobalDBManager.GetShardingConfig().NeonShards)
		fmt.Printf("Supabase Shards: %d\n", config.GlobalDBManager.GetShardingConfig().SupabaseShards)
		fmt.Printf("PlanetScale Shards: %d\n", config.GlobalDBManager.GetShardingConfig().PlanetScaleShards)
		fmt.Printf("Upstash Shards: %d\n", config.GlobalDBManager.GetShardingConfig().UpstashShards)

	case "scale-servers":
		if len(os.Args) < 3 {
			log.Fatal("Usage: scale-servers <count>")
		}
		count, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatal("Invalid server count:", err)
		}

		for i := 1; i <= count; i++ {
			serverURL := fmt.Sprintf("https://app-%d.onrender.com", i)
			if err := config.GlobalDBManager.AddRenderServer(serverURL); err != nil {
				log.Printf("Failed to add server %d: %v", i, err)
				continue
			}
			fmt.Printf("Added server %d: %s\n", i, serverURL)
		}

	case "scale-db":
		if len(os.Args) < 4 {
			log.Fatal("Usage: scale-db <type> <count>")
		}
		dbType := os.Args[2]
		count, err := strconv.Atoi(os.Args[3])
		if err != nil {
			log.Fatal("Invalid database count:", err)
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

			if err := config.GlobalDBManager.AddDatabase(dbType, connectionURL); err != nil {
				log.Printf("Failed to add %s database %d: %v", dbType, i, err)
				continue
			}
			fmt.Printf("Added %s database %d: %s\n", dbType, i, connectionURL)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
