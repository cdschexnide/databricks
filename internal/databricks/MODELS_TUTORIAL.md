# models.go Tutorial

This file defines the data structures (models) used throughout the Databricks package for ingestion operations. These models represent the contracts between different parts of the system.

## Package Declaration and Imports

```go
package databricks

import (
    "encoding/json"
    "time"
)
```

- **encoding/json**: For JSON marshaling of results
- **time**: For duration tracking in results

## IngestionRequest Struct (lines 9-17)

```go
type IngestionRequest struct {
    TableName     string            `json:"tableName"`
    SourcePath    string            `json:"sourcePath"`
    FileFormat    string            `json:"fileFormat"` // JSON or CSV
    FormatOptions string            `json:"formatOptions"`
    DataSource    string            `json:"dataSource"`  // BLADE/ADVANA
    SampleData    string            `json:"sampleData,omitempty"` // For mock POC mode
    Metadata      map[string]string `json:"metadata"`
}
```

This struct represents a request to ingest data into Databricks:

### Fields:
- **TableName**: The target Databricks table name (e.g., "blade_maintenance_data")
- **SourcePath**: Path to the source data file (used for real BLADE integration)
- **FileFormat**: The format of the source data ("JSON" or "CSV")
- **FormatOptions**: Additional format-specific options (e.g., CSV delimiter)
- **DataSource**: Identifies the data source system ("BLADE" or "ADVANA")
- **SampleData**: Contains actual JSON data for POC mode (omitted if empty)
- **Metadata**: Flexible key-value pairs for additional information

### JSON Tags:
Each field has a `json` tag that controls JSON marshaling:
- Standard tags like `json:"tableName"` map struct fields to JSON keys
- `omitempty` on SampleData means it's excluded from JSON if empty

## IngestionResult Struct (lines 20-27)

```go
type IngestionResult struct {
    RowsIngested int64 `json:"rowsIngested"`
    Duration time.Duration `json:"duration"`
    TableName string `json:"tableName"`
    Status string `json:"status"`
    Error error `json:"error,omitempty"`
    Metadata map[string]interface{} `json:"metadata"`
}
```

This struct represents the outcome of an ingestion operation:

### Fields:
- **RowsIngested**: Number of rows successfully inserted into Databricks
- **Duration**: Total time taken for the ingestion operation
- **TableName**: The table where data was ingested
- **Status**: Operation status ("completed" or "failed")
- **Error**: Contains error details if operation failed (omitted if nil)
- **Metadata**: Flexible metadata using `interface{}` for various value types

### Design Notes:
- Uses `int64` for row count to handle large datasets
- `time.Duration` provides nanosecond precision for performance tracking
- `map[string]interface{}` in Metadata allows any JSON-serializable values

## ToJSON Method (lines 31-34)

```go
func (r *IngestionResult) ToJSON() []byte {
    data, _ := json.Marshal(r)
    return data
}
```

This method converts an IngestionResult to JSON bytes:
- Uses Go's standard `json.Marshal` function
- Ignores potential marshal errors (returns empty bytes on error)
- Useful for logging, API responses, or catalog integration

## BLADEDataType Type and Constants (lines 37-44)

```go
type BLADEDataType string

const (
    MaintenanceData BLADEDataType = "maintenance"
    SortieData BLADEDataType = "sortie"
    DeploymentData BLADEDataType = "deployment"
    LogisticsData BLADEDataType = "logistics"
)
```

Defines a custom type for BLADE data categories:

### Type Definition:
- `BLADEDataType` is a string-based type for type safety
- Prevents arbitrary strings from being used where data types are expected

### Constants:
- **MaintenanceData**: Aircraft maintenance records
- **SortieData**: Flight operations and missions
- **DeploymentData**: Personnel and equipment deployments
- **LogisticsData**: Supply chain and equipment transfers

### Benefits:
1. **Type Safety**: Can't accidentally pass wrong string values
2. **IDE Support**: Auto-completion for valid data types
3. **Documentation**: Self-documenting code with clear options
4. **Refactoring**: Easy to rename or add new types

## Usage Examples

### Creating an Ingestion Request:
```go
req := &IngestionRequest{
    TableName:  "blade_maintenance_data",
    FileFormat: "JSON",
    DataSource: "BLADE",
    SampleData: jsonData,
    Metadata: map[string]string{
        "data_type": "maintenance",
        "mode": "mock_data",
    },
}
```

### Creating an Ingestion Result:
```go
result := &IngestionResult{
    RowsIngested: 150,
    Duration:     5 * time.Second,
    TableName:    "blade_maintenance_data",
    Status:       "completed",
    Metadata: map[string]interface{}{
        "source_file": "maintenance_data.json",
        "batch_id": "12345",
    },
}
```

### Converting Result to JSON:
```go
jsonBytes := result.ToJSON()
fmt.Println(string(jsonBytes))
```

## Design Patterns

1. **Data Transfer Objects (DTOs)**: These structs act as DTOs, carrying data between layers
2. **JSON Serialization**: Built-in JSON support for API integration
3. **Flexible Metadata**: Maps allow extending data without changing struct
4. **Type Safety**: Custom types prevent string mix-ups
5. **Receiver Methods**: ToJSON method attached to struct for OOP-style usage