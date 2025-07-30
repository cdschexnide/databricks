# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a proof-of-concept (POC) application that demonstrates ingesting BLADE (Basing & Logistics Analytics Data Environment) data into Databricks. The application processes mock UNCLASSIFIED Air Force logistics data and loads it into Databricks Delta tables.

## Development Commands

### Build and Run
```bash
# Build the application
go build -o main cmd/main.go

# Run with default data type (maintenance)
go run cmd/main.go

# Run with specific data type
go run cmd/main.go sortie
go run cmd/main.go deployment  
go run cmd/main.go logistics
go run cmd/main.go maintenance
```

### Dependencies
```bash
# Install/update dependencies
go mod tidy

# Verify dependencies
go mod verify
```

## Architecture

### Core Components

- **cmd/main.go**: Application entry point that orchestrates the entire ingestion process
- **internal/config/**: Environment configuration management (loads .env file)
- **internal/databricks/**: Databricks client, connection management, and ingestion logic
- **internal/blade/**: BLADE data type mapping and adapter logic
- **mock_blade_data/**: Mock JSON data files representing BLADE system data types

### Data Flow

1. Load configuration from .env file (Databricks credentials and settings)
2. Create authenticated Databricks client and test connection
3. Initialize BLADE adapter with supported data types mapping
4. Process command line arguments to determine data type
5. Prepare ingestion request (load JSON, map to table name, set metadata)
6. Execute ingestion (create table if needed, parse JSON, generate INSERT SQL, execute)
7. Display results and validate row counts

### Key Features

- **Multi-data type support**: maintenance, logistics, sortie, deployment
- **Delta table creation**: Automatic table creation with appropriate schema
- **Mock data processing**: Converts JSON arrays to SQL INSERT statements
- **Connection validation**: Tests Databricks connection before processing
- **Comprehensive logging**: Detailed progress and error reporting

## Configuration

### Required Environment Variables (.env file)
```
DATABRICKS_HOST=https://your-workspace.cloud.databricks.com
DATABRICKS_TOKEN=your-personal-access-token
DATABRICKS_WAREHOUSE_ID=your-warehouse-id
```

### Optional Environment Variables
```
DATABRICKS_CATALOG=blade_poc          # Default catalog name
DATABRICKS_SCHEMA=logistics           # Default schema name
```

## Data Types

The BLADE adapter supports four main data categories:

- **maintenance**: Aircraft maintenance records (F-16, A-10, F-22)
- **logistics**: Supply chain, fuel, munitions, equipment transfers
- **sortie**: Flight operations, training missions, combat exercises
- **deployment**: Personnel/equipment deployments, humanitarian operations

Each data type maps to a specific JSON file in `mock_blade_data/` and creates corresponding Databricks tables with naming convention `blade_{datatype}_data`.

## Table Schema

All Databricks tables use this unified schema:
- `item_id` STRING - Unique identifier
- `item_type` STRING - Category within data type
- `classification_marking` STRING - Always "UNCLASSIFIED" for POC
- `timestamp` TIMESTAMP - Event timestamp
- `data_source` STRING - Source system identifier
- `raw_data` STRING - Original JSON data
- `ingestion_timestamp` TIMESTAMP - When data was ingested
- `metadata` MAP<STRING, STRING> - Additional metadata

## Dependencies

- **github.com/databricks/databricks-sdk-go**: Official Databricks SDK
- **github.com/joho/godotenv**: Environment variable loading from .env files
- **Go 1.24.5+**: Required Go version

## Important Notes

- All mock data is UNCLASSIFIED and created for demonstration purposes
- POC uses INSERT statements rather than bulk loading methods
- Row count validation is implemented but not fully functional in current POC version
- Real BLADE integration would require different data sources and security handling