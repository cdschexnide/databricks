# client.go Tutorial

This file contains the Databricks client implementation that handles all interactions with the Databricks workspace using the official Databricks SDK for Go.

## Package Declaration and Imports

```go
package databricks

import (
    "context"
    "fmt"
    "github.com/databricks/databricks-sdk-go"
    "github.com/databricks/databricks-sdk-go/service/sql" 
    "databricks-blade-poc/internal/config"
)
```

- **context**: For managing request lifecycle and cancellation
- **fmt**: For formatting error messages
- **databricks-sdk-go**: Official Databricks SDK for Go
- **service/sql**: SQL-specific services from the Databricks SDK
- **internal/config**: Application configuration types

## Client Struct (lines 11-16)

```go
type Client struct {
    workspace *databricks.WorkspaceClient // Databricks SDK Client
    warehouseID string // which SQL warehouse to use
    catalog string // default catalog name
    schema string // default schema name
}
```

The `Client` struct encapsulates all the information needed to interact with Databricks:
- **workspace**: The authenticated SDK client for making API calls
- **warehouseID**: Identifies which SQL warehouse will execute queries
- **catalog**: The default catalog (database) for table operations
- **schema**: The default schema within the catalog

## NewClient Constructor (lines 20-36)

```go
func NewClient(cfg *config.Config) (*Client, error) {
    w, err := databricks.NewWorkspaceClient(&databricks.Config{
        Host:  cfg.DatabricksHost,
        Token: cfg.DatabricksToken,
    })

    if err != nil {
        return nil, fmt.Errorf("failed to create Databricks client: %w", err)
    }

    return &Client{
        workspace:   w,
        warehouseID: cfg.WarehouseID,
        catalog:     cfg.CatalogName,
        schema:      cfg.SchemaName,
    }, nil
}
```

This constructor function:
1. **Creates a Databricks workspace client** using host URL and authentication token
2. **Handles authentication** automatically through the SDK
3. **Wraps errors** with context using `fmt.Errorf` and `%w` verb
4. **Returns initialized client** with all necessary configuration

## TestConnection Method (lines 40-51)

```go
func (c *Client) TestConnection(ctx context.Context) error {
    // Fixed: List() returns an iterator, need to iterate through it to test connection
    warehouses := c.workspace.Warehouses.List(ctx, sql.ListWarehousesRequest{})
    _, err := warehouses.Next(ctx)
    if err != nil {
        return fmt.Errorf("failed to connect to Databricks: %w", err)
    }
    return nil
}
```

This method validates the connection by:
1. **Listing warehouses** - a lightweight API call to verify credentials
2. **Using an iterator pattern** - The SDK returns an iterator, not a list
3. **Calling Next()** - Attempts to retrieve at least one warehouse
4. **Error handling** - Any error indicates connection or authentication issues

Note the fix from the original broken code that tried to assign an iterator to an error variable.

## ensureTableExists Method (lines 55-95)

```go
func (c *Client) ensureTableExists(ctx context.Context, req *IngestionRequest) error {
    createTableSQL := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s.%s.%s (
            item_id STRING,
            item_type STRING,
            classification_marking STRING,
            timestamp TIMESTAMP,
            data_source STRING,
            raw_data STRING,
            ingestion_timestamp TIMESTAMP,
            metadata MAP<STRING, STRING>
        )
        USING DELTA
        TBLPROPERTIES (
            'delta.feature.allowColumnDefaults' = 'supported',
            'source_system' = 'BLADE',
            'data_type' = '%s'
        )
    `, c.catalog, c.schema, req.TableName, req.Metadata["data_type"])
```

This private method (lowercase name) creates tables with:

### Schema Definition:
- **item_id**: Unique identifier for each record
- **item_type**: Category within the data type
- **classification_marking**: Security classification (always "UNCLASSIFIED" for POC)
- **timestamp**: When the event occurred
- **data_source**: System that generated the data
- **raw_data**: Original JSON data as string
- **ingestion_timestamp**: When data was loaded into Databricks
- **metadata**: Flexible key-value pairs for additional information

### Table Properties:
- **USING DELTA**: Creates a Delta Lake table for ACID transactions
- **delta.feature.allowColumnDefaults**: Enables column defaults
- **source_system**: Tags table as coming from BLADE
- **data_type**: Stores the specific BLADE data type

### Execution:
```go
_, err := c.workspace.StatementExecution.ExecuteStatement(
    ctx,
    sql.ExecuteStatementRequest{
        Statement:   createTableSQL,
        WarehouseId: c.warehouseID,
        Catalog:     c.catalog,
        Schema:      c.schema,
        WaitTimeout: "30s", // Fixed: within allowed range of 5s-50s
    },
)
```

Note the timeout fix - Databricks API requires timeouts between 5s and 50s.

## executeSQL Helper Method (lines 99-112)

```go
func (c *Client) executeSQL(ctx context.Context, sqlStatement string) (*sql.StatementResponse, error) {
    return c.workspace.StatementExecution.ExecuteStatement(
        ctx,
        sql.ExecuteStatementRequest{
            Statement:   sqlStatement,
            WarehouseId: c.warehouseID,
            Catalog:     c.catalog,
            Schema:      c.schema,
            WaitTimeout: "50s", // Fixed: maximum allowed timeout
        },
    )
}
```

A reusable helper for executing SQL statements:
- **Encapsulates common parameters**: warehouse, catalog, schema
- **Returns full response**: Allows callers to process results
- **Maximum timeout**: Uses 50s for potentially long-running operations

## getRowCount Validation Method (lines 117-139)

```go
func (c *Client) getRowCount(ctx context.Context, tableName string) (int64, error) {
    countSQL := fmt.Sprintf("SELECT COUNT(*) as row_count FROM %s.%s.%s", 
        c.catalog, c.schema, tableName)

    _, err := c.workspace.StatementExecution.ExecuteStatement(
        ctx,
        sql.ExecuteStatementRequest{
            Statement:   countSQL,
            WarehouseId: c.warehouseID,
            Catalog:     c.catalog,
            Schema:      c.schema,
            WaitTimeout: "30s",
        },
    )

    if err != nil {
        return 0, fmt.Errorf("failed to get row count: %w", err)
    }
    
    return 0, fmt.Errorf("row count validation not fully implemented in POC")
}
```

This method is intended for validation but is not fully implemented:
1. **Builds COUNT query** using fully qualified table name
2. **Executes the query** but doesn't process results
3. **Returns placeholder error** indicating incomplete implementation

In a complete implementation, this would parse the query results and return the actual row count.

## Key Design Patterns

1. **Receiver Methods**: All methods use pointer receivers `(c *Client)` for efficiency
2. **Context Propagation**: Every method accepts context for cancellation support
3. **Error Wrapping**: Uses `fmt.Errorf` with `%w` verb for error chain preservation
4. **Private Methods**: `ensureTableExists` is private (lowercase) to hide implementation details
5. **Configuration Encapsulation**: Client stores all needed config to avoid passing it repeatedly