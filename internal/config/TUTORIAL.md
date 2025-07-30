# config.go Tutorial

This file handles configuration management for the BLADE-to-Databricks POC application. It loads environment variables and provides default values where appropriate.

## Package Declaration and Imports

```go
package config

import (
    "os"
    "github.com/joho/godotenv"
)
```

- **os**: Provides access to environment variables
- **github.com/joho/godotenv**: External library that loads variables from `.env` files into the environment

## Config Struct (lines 8-19)

```go
type Config struct {
    // Databricks Configuration
    DatabricksHost string
    DatabricksToken string
    WarehouseID string
    CatalogName string
    SchemaName string

    // BLADE Configuration (hardcoded for POC)
    BLADEDataPath string
    BLADEDataSource string
}
```

The `Config` struct contains two categories of configuration:

### Databricks Configuration
- **DatabricksHost**: The URL of your Databricks workspace (e.g., `https://your-workspace.cloud.databricks.com`)
- **DatabricksToken**: Personal access token for authentication
- **WarehouseID**: Identifier for the SQL warehouse to execute queries
- **CatalogName**: The catalog where tables will be created (defaults to "blade_poc")
- **SchemaName**: The schema within the catalog (defaults to "logistics")

### BLADE Configuration
- **BLADEDataPath**: Directory containing mock JSON data files (hardcoded to "mock_blade_data/")
- **BLADEDataSource**: Identifier for the data source (hardcoded to "BLADE_LOGISTICS")

## LoadConfig Function (lines 21-36)

```go
func LoadConfig() (*Config, error) {
    // Load .env file if it exists (this is how infinityai-cataloger does)
    _ = godotenv.Load(".env")

    return &Config{
        DatabricksHost: os.Getenv("DATABRICKS_HOST"),
        DatabricksToken: os.Getenv("DATABRICKS_TOKEN"),
        WarehouseID: os.Getenv("DATABRICKS_WAREHOUSE_ID"),
        CatalogName: getEnvOrDefault("DATABRICKS_CATALOG", "blade_poc"),
        SchemaName: getEnvOrDefault("DATABRICKS_SCHEMA", "logistics"),

        // Hardcoded for POC - will be dynamic in integration
        BLADEDataPath: "mock_blade_data/",
        BLADEDataSource: "BLADE_LOGISTICS",
    }, nil
}
```

### Key Behaviors:

1. **Environment File Loading** (line 23):
   ```go
   _ = godotenv.Load(".env")
   ```
   - Attempts to load a `.env` file from the current directory
   - The underscore (`_`) discards any error - the file is optional
   - If the file exists, its variables are loaded into the environment

2. **Required Variables** (lines 26-28):
   - `DATABRICKS_HOST`, `DATABRICKS_TOKEN`, and `DATABRICKS_WAREHOUSE_ID` are read directly
   - No defaults provided - these must be set or the application will fail later

3. **Optional Variables with Defaults** (lines 29-30):
   - Uses the `getEnvOrDefault` helper function
   - `DATABRICKS_CATALOG` defaults to "blade_poc"
   - `DATABRICKS_SCHEMA` defaults to "logistics"

4. **Hardcoded POC Values** (lines 33-34):
   - BLADE configuration is hardcoded for the POC
   - In a real integration, these would likely come from environment variables

## Helper Function: getEnvOrDefault (lines 38-43)

```go
func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value;
    }
    return defaultValue;
}
```

This utility function provides a common pattern for optional environment variables:

1. **Retrieves the environment variable**: `os.Getenv(key)`
2. **Checks if it's empty**: If the variable is unset or empty string
3. **Returns appropriate value**: 
   - If set and non-empty: returns the environment value
   - Otherwise: returns the provided default

## Usage Example

To use this configuration system:

1. Create a `.env` file in your project root:
   ```
   DATABRICKS_HOST=https://your-workspace.cloud.databricks.com
   DATABRICKS_TOKEN=dapi1234567890abcdef
   DATABRICKS_WAREHOUSE_ID=abc123def456
   DATABRICKS_CATALOG=my_catalog
   DATABRICKS_SCHEMA=my_schema
   ```

2. The configuration is loaded in `main.go`:
   ```go
   cfg, err := config.LoadConfig()
   ```

## Design Decisions

1. **Simple Error Handling**: The function always returns `nil` for error, simplifying usage but potentially hiding `.env` loading issues

2. **Mixed Configuration Sources**: Combines environment variables (flexible) with hardcoded values (simple for POC)

3. **Default Values**: Provides sensible defaults for optional parameters to reduce configuration burden

4. **Separation of Concerns**: Keeps all configuration logic in one place, making it easy to modify or extend