package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"databricks-blade-poc/internal/blade"
	"databricks-blade-poc/internal/config"
	"databricks-blade-poc/internal/databricks"
)

func main() {
	// creates background context for request lifecycle
	ctx := context.Background()

	// This context will be passed to all operations and allows for:
    // - Request cancellation
    // - Timeout handling  
    // - Graceful shutdown

	// Load configuration from environment variables and .env file
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate that all required Databricks settings are present
	if cfg.DatabricksHost == "" || cfg.DatabricksToken == "" || cfg.WarehouseID == "" {
		log.Fatal("Missing required Databricks configuration. Check your .env file.")
	}

	// Create authenticated connection to Databricks workspace
	dbClient, err := databricks.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create Databricks client: %v", err)
	}

	// Test the connection before proceeding (REAL Databricks API call)
	if err := dbClient.TestConnection(ctx); err != nil {
		log.Fatalf("Failed to connect to Databricks: %v", err)
	}
	log.Println("✓ Connected to Databricks successfully")

	// sets up data type mapping system
	// translates BLADE data types to Databricks tables
	bladeAdapter := blade.NewBLADEAdapter(
		cfg.BLADEDataSource,
		cfg.BLADEDataPath,
	)

	// Show what data types are available from bladeAdapter
	log.Printf("Supported BLADE data types: %v", bladeAdapter.GetSupportedDataTypes())

	// processes command line arguments
	// defaults to "maintenance", if no argument is provided
	dataType := "maintenance"
	if len(os.Args) > 1 {
		dataType = os.Args[1]
	}

	// Usage examples
	// 	go run cmd/main.go                   # Uses "maintenance" (default)
	//  go run cmd/main.go sortie            # Processes sortie data
	//  go run cmd/main.go deployment        # Processes deployment data
	//  go run cmd/main.go logistics         # Processes logistics data

	log.Printf("Starting ingestion for BLADE %s data...", dataType)

	// converts BLADE data type to ingestion parameters
	// input: "maintenance" (user requested data type)
	// lookup: Find mapping in BLADE adapter
	// output: IngestionRequest struct with all parameters
	req, err := bladeAdapter.PrepareIngestionRequest(dataType)

	// What the adapter does:
	//   Validates the data type is supported
	//   Loads the corresponding JSON file (e.g., maintenance_data.json)
	//   Creates an IngestionRequest with all necessary parameters:
	//   Table name (e.g., blade_maintenance_data)
	//   JSON data content
	//   Metadata for tracking

	if err != nil {
		log.Fatalf("Failed to prepare ingestion request: %v", err)
	}

	// loads mock BLADE data into Databricks tables
	// method: INSERT statements with mock data (POC mode)
	// source: Mock JSON data embedded in adapter
	// destination: Real Databricks Delta table
	result, err := dbClient.IngestBLADEData(ctx, req)

	// What happens during ingestion
	//   Table Creation: Creates Databricks table if it doesn't exist
	//   Data Parsing: Parses JSON array from your mock file
	//   SQL Generation: Builds INSERT statements for each record
	//   Execution: Sends SQL to Databricks warehouse for processing
	//   Validation: Counts rows to verify successful insertion

	if err != nil {
		log.Fatalf("Ingestion failed: %v", err)
	}

	// displays results
	fmt.Printf("\n" + strings.Repeat("=", 50) + "\n")
	fmt.Printf("BLADE INGESTION RESULTS\n")
	fmt.Printf(strings.Repeat("=", 50) + "\n")
	fmt.Printf("Table: %s\n", result.TableName)
	fmt.Printf("Status: %s\n", result.Status)
	fmt.Printf("Rows Ingested: %d\n", result.RowsIngested)
	fmt.Printf("Duration: %v\n", result.Duration)
	fmt.Printf("Source: %s\n", req.DataSource)
	fmt.Printf(strings.Repeat("=", 50) + "\n")

	log.Println("✓ Ingestion completed successfully")
}