# ingestion.go Tutorial

This file handles the core data ingestion logic, transforming mock BLADE JSON data into SQL INSERT statements and executing them against Databricks Delta tables.

## Package Declaration and Imports

```go
package databricks

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"
    "time"
    "github.com/databricks/databricks-sdk-go/service/sql"
)
```

- **context**: For request lifecycle management
- **encoding/json**: To parse JSON data from mock files
- **fmt**: For string formatting and error messages
- **strings**: For string manipulation (escaping quotes, joining values)
- **time**: For timestamps and duration tracking
- **service/sql**: Databricks SQL service types

## IngestBLADEData Method (lines 14-62)

```go
func (c *Client) IngestBLADEData(ctx context.Context, req *IngestionRequest) (*IngestionResult, error) {
    start := time.Now()
```

This is the main entry point for data ingestion, orchestrating the entire process:

### 1. Performance Tracking
```go
start := time.Now()
```
Records start time to measure ingestion duration.

### 2. Table Creation (lines 18-25)
```go
if err := c.ensureTableExists(ctx, req); err != nil {
    return &IngestionResult{
        TableName: req.TableName,        
        Status:    "failed",               
        Error:     err,                   
        Duration:  time.Since(start),    
    }, fmt.Errorf("failed to ensure table exists: %w", err)
}
```
- Calls `ensureTableExists` to create table if needed
- Returns detailed error result if table creation fails
- Includes duration even for failures

### 3. Mock Data Mode Check (lines 28-59)
```go
if req.SampleData != "" && req.Metadata["mode"] == "mock_data" {
    // Process mock data
}
```
Checks if this is POC mode by verifying:
- `SampleData` field contains JSON data
- Metadata indicates "mock_data" mode

### 4. Mock Data Insertion (lines 30-38)
```go
rowsInserted, err := c.insertMockData(ctx, req)
if err != nil {
    return &IngestionResult{
        TableName: req.TableName,
        Status:    "failed",        
        Error:     err,               
        Duration:  time.Since(start), 
    }, fmt.Errorf("failed to insert mock data: %w", err)
}
```
Delegates to `insertMockData` method and handles errors.

### 5. Row Count Validation (lines 41-44)
```go
rowCount, err := c.getRowCount(ctx, req.TableName)
if err != nil {
    rowCount = rowsInserted
}
```
Attempts to validate insertion by counting rows. Falls back to inserted count if validation fails.

### 6. Success Result (lines 46-58)
```go
return &IngestionResult{
    RowsIngested: rowCount,      
    Duration:     time.Since(start),  
    TableName:    req.TableName,      
    Status:       "completed",      
    Metadata: map[string]interface{}{ 
        "source_path":    req.SourcePath,    
        "file_format":    req.FileFormat,      
        "data_source":    req.DataSource,      
        "blade_metadata": req.Metadata,      
        "ingestion_type": "mock_data_insert",  
    },
}, nil
```
Returns comprehensive result including:
- Row count
- Total duration
- Detailed metadata for auditing

### 7. Real BLADE Placeholder (line 61)
```go
return nil, fmt.Errorf("real BLADE ingestion not implemented - use mock data mode for POC")
```
Returns error if not in mock mode, as real BLADE integration isn't implemented.

## insertMockData Method (lines 66-145)

This method performs the actual data insertion using SQL INSERT statements:

### 1. JSON Parsing (lines 67-72)
```go
var records []map[string]interface{} 

if err := json.Unmarshal([]byte(req.SampleData), &records); err != nil {
    return 0, fmt.Errorf("failed to parse sample data: %w", err)
}
```
- Parses JSON array into slice of maps
- Each map represents one record to insert
- Uses `interface{}` for flexible value types

### 2. Batch Preparation (lines 75-77)
```go
var values []string
batchID := fmt.Sprintf("%d", time.Now().Unix())
```
- Creates slice to hold SQL VALUES clauses
- Generates unique batch ID using Unix timestamp
- Batch ID helps track related records

### 3. Record Transformation Loop (lines 79-105)
```go
for _, record := range records {
    // Convert record back to JSON for raw_data storage
    rawDataJSON, _ := json.Marshal(record) 
    rawDataEscaped := strings.ReplaceAll(string(rawDataJSON), "'", "''")
```

For each record:
1. **Marshals back to JSON** for `raw_data` column
2. **Escapes single quotes** by doubling them (SQL standard)

### 4. VALUES Clause Building (lines 85-103)
```go
value := fmt.Sprintf(`(
    '%s',                              // item_id
    '%s',                              // item_type
    '%s',                              // classification_marking
    TIMESTAMP '%s',                    // timestamp
    '%s',                              // data_source
    '%s',                              // raw_data (escaped JSON)
    current_timestamp(),               // ingestion_timestamp
    map('source', 'mock_blade', 'batch_id', '%s', 'data_type', '%s')  // metadata
)`,
    record["item_id"],                  
    record["item_type"],            
    record["classification_marking"],  
    record["timestamp"],   
    req.DataSource,                     
    rawDataEscaped,             
    batchID,                        
    req.Metadata["data_type"],  
)
```

Creates VALUES clause with:
- Direct field mapping from JSON
- `current_timestamp()` for ingestion time
- MAP literal for metadata with source, batch_id, and data_type

### 5. SQL Statement Construction (lines 108-123)
```go
insertSQL := fmt.Sprintf(`
    INSERT INTO %s.%s.%s (
        item_id,
        item_type,
        classification_marking,
        timestamp,
        data_source,
        raw_data,
        ingestion_timestamp,
        metadata
    ) VALUES %s
`, 
    c.catalog,    
    c.schema,   
    req.TableName, 
    strings.Join(values, ",\n"))
```

Builds complete INSERT statement:
- Fully qualified table name (catalog.schema.table)
- Explicit column list
- Multiple VALUES clauses joined with commas

### 6. Execution (lines 128-141)
```go
_, err := c.workspace.StatementExecution.ExecuteStatement(
    ctx,
    sql.ExecuteStatementRequest{ 
        Statement:   insertSQL,   
        WarehouseId: c.warehouseID,  
        Catalog:     c.catalog,     
        Schema:      c.schema,       
        WaitTimeout: "30s", // Fixed: within allowed range of 5s-50s
    },
)
```

Executes the batch INSERT with:
- Proper timeout (fixed from original 60s)
- All required connection parameters
- Error handling for execution failures

### 7. Return Count (line 144)
```go
return int64(len(records)), nil
```
Returns the number of records inserted.

## Key Design Decisions

1. **Batch Insertion**: All records inserted in one SQL statement for efficiency
2. **Raw Data Preservation**: Stores original JSON in `raw_data` column
3. **Metadata Tracking**: Uses MAP type for flexible metadata storage
4. **Error Details**: Returns comprehensive error information for debugging
5. **SQL Injection Prevention**: Uses quote escaping (though parameterized queries would be better)
6. **Timeout Management**: Respects Databricks API timeout constraints (5s-50s)

## Limitations

1. **No Parameterized Queries**: Uses string concatenation instead of prepared statements
2. **Limited Validation**: Row count validation not fully implemented
3. **Mock Data Only**: Real BLADE integration placeholder exists but not implemented
4. **No Transaction Support**: Each INSERT is independent, no rollback capability