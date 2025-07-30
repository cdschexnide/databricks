# main.go Tutorial

This file is the entry point for the BLADE-to-Databricks POC application. It orchestrates the entire data ingestion process from start to finish.

## Package Declaration and Imports

```go
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
```

- **Standard Library Imports**: 
  - `context`: Manages request lifecycle and cancellation
  - `fmt`: Formatted output for results display
  - `log`: Application logging
  - `os`: Command-line argument access
  - `strings`: String manipulation for output formatting

- **Internal Package Imports**:
  - `internal/blade`: BLADE adapter for data type mapping
  - `internal/config`: Configuration management (.env file loading)
  - `internal/databricks`: Databricks client and ingestion logic

## Main Function Walkthrough

### 1. Context Creation (lines 15-22)

```go
ctx := context.Background()
```

Creates a background context that will be passed to all operations. This enables:
- Request cancellation capabilities
- Timeout handling for long-running operations
- Graceful shutdown support

### 2. Configuration Loading (lines 24-32)

```go
cfg, err := config.LoadConfig()
if err != nil {
    log.Fatalf("Failed to load configuration: %v", err)
}
```

Loads configuration from the `.env` file. The function validates that all required Databricks settings are present:
- `DATABRICKS_HOST`: Workspace URL
- `DATABRICKS_TOKEN`: Authentication token
- `DATABRICKS_WAREHOUSE_ID`: SQL warehouse identifier

### 3. Databricks Client Creation (lines 35-44)

```go
dbClient, err := databricks.NewClient(cfg)
if err != nil {
    log.Fatalf("Failed to create Databricks client: %v", err)
}

if err := dbClient.TestConnection(ctx); err != nil {
    log.Fatalf("Failed to connect to Databricks: %v", err)
}
```

- Creates an authenticated Databricks client
- Tests the connection with a real API call before proceeding
- Ensures credentials and network connectivity are valid

### 4. BLADE Adapter Initialization (lines 48-54)

```go
bladeAdapter := blade.NewBLADEAdapter(
    cfg.BLADEDataSource,
    cfg.BLADEDataPath,
)
```

Initializes the BLADE adapter which:
- Maps BLADE data types to Databricks table names
- Knows where to find mock JSON data files
- Handles data type validation

### 5. Command-Line Argument Processing (lines 58-68)

```go
dataType := "maintenance"
if len(os.Args) > 1 {
    dataType = os.Args[1]
}
```

Processes command-line arguments to determine which data type to ingest:
- Default: "maintenance" if no argument provided
- Supported types: maintenance, sortie, deployment, logistics
- Usage: `go run cmd/main.go [dataType]`

### 6. Ingestion Request Preparation (lines 75-87)

```go
req, err := bladeAdapter.PrepareIngestionRequest(dataType)
```

The adapter performs several tasks:
- Validates the requested data type is supported
- Loads the corresponding JSON file (e.g., `maintenance_data.json`)
- Creates an `IngestionRequest` containing:
  - Target table name (e.g., `blade_maintenance_data`)
  - JSON data content to be ingested
  - Metadata for tracking and auditing

### 7. Data Ingestion Execution (lines 93-104)

```go
result, err := dbClient.IngestBLADEData(ctx, req)
```

The ingestion process involves:
- **Table Creation**: Creates the Databricks Delta table if it doesn't exist
- **Data Parsing**: Parses the JSON array from the mock file
- **SQL Generation**: Builds INSERT statements for each record
- **Execution**: Sends SQL to Databricks warehouse for processing
- **Validation**: Counts rows to verify successful insertion

### 8. Results Display (lines 107-117)

```go
fmt.Printf("\n" + strings.Repeat("=", 50) + "\n")
fmt.Printf("BLADE INGESTION RESULTS\n")
fmt.Printf(strings.Repeat("=", 50) + "\n")
fmt.Printf("Table: %s\n", result.TableName)
fmt.Printf("Status: %s\n", result.Status)
fmt.Printf("Rows Ingested: %d\n", result.RowsIngested)
fmt.Printf("Duration: %v\n", result.Duration)
fmt.Printf("Source: %s\n", req.DataSource)
```

Displays a formatted summary of the ingestion results including:
- Target table name
- Success/failure status
- Number of rows ingested
- Time taken for the operation
- Data source identifier

## Error Handling

The application uses `log.Fatalf()` for all error conditions, which:
- Logs the error message with timestamp
- Exits the program with status code 1
- Ensures no partial operations continue after failure

## Key Design Decisions

1. **Sequential Processing**: Operations are performed sequentially to ensure each step completes successfully before proceeding
2. **Early Validation**: Connection testing happens before any data processing
3. **Clear Feedback**: Success checkmarks (✓) and detailed results provide clear operational feedback
4. **Simple CLI**: Command-line argument processing is straightforward and user-friendly