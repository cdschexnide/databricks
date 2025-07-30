package databricks

import (
	"context"
	"fmt"
	"github.com/databricks/databricks-sdk-go"
	"github.com/databricks/databricks-sdk-go/service/sql" 
	"databricks-blade-poc/internal/config"
)

type Client struct {
	workspace *databricks.WorkspaceClient // Databricks SDK Client
	warehouseID string // which SQL warehouse to use
	catalog string // default catalog name
	schema string // default schema name
}

// takes in a Config pointer
// returns a Client pointer and an error
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

// receiver function for the Client struct, tests Databricks connection
// takes in context, returns error
func (c *Client) TestConnection(ctx context.Context) error {
	err := c.workspace.Warehouses.List(ctx, sql.ListWarehousesRequest{})
	if err != nil {
		return fmt.Errorf("failed to connect to Databricks: %w", err)
	}
	return nil
}

// private method (lowercase name means only this package can call it)
// creates the target table if it doesn't exist
func (c *Client) ensureTableExists(ctx context.Context, req *IngestionRequest) error {
	// Create table with BLADE-appropriate schema
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

	// ExecuteStatement -> SDK method to run SQL
	_, err := c.workspace.StatementExecution.ExecuteStatement(
		ctx,
		sql.ExecuteStatementRequest{
			Statement:   createTableSQL,
			WarehouseId: c.warehouseID,
			Catalog:     c.catalog,
			Schema:      c.schema,
			WaitTimeout: "60s",
		},
	)

	if err != nil {
		return fmt.Errorf("failed to create table %s: %w", req.TableName, err)
	}

	return nil
}


// executeSQL executes a SQL statement and returns the response
func (c *Client) executeSQL(ctx context.Context, sqlStatement string) (*sql.StatementResponse, error) {
	return c.workspace.StatementExecution.ExecuteStatement(
		ctx,
		sql.ExecuteStatementRequest{
			Statement:   sqlStatement,
			WarehouseId: c.warehouseID,
			Catalog:     c.catalog,
			Schema:      c.schema,
			WaitTimeout: "300s", // 5 minutes for large operations
		},
	)
}

// Client receiver function that performs Row validation (simple validation for PoC)
// counts rows in a table for validation after ingestion
// returns a 64-bit integer and error
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

	// For POC purposes, we'll return the number of records we attempted to insert
	// In a real implementation, you'd parse the result to get the actual count
	return 0, fmt.Errorf("row count validation not fully implemented in POC")
}

