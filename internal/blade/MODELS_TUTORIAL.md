# models.go Tutorial

This file defines the data structures and mappings for BLADE data types. It establishes the relationship between BLADE data categories and their corresponding Databricks table configurations.

## Package Declaration

```go
package blade
```

This file is part of the `blade` package, which handles all BLADE-specific logic and data type mappings.

## BLADEDataMapping Struct (lines 3-8)

```go
type BLADEDataMapping struct {
    DataType    string `json:"dataType"` // BLADE data type
    TableName   string `json:"tableName"` // corresponding Databricks table name
    SourcePath  string `json:"sourcePath"` // mock path for POC (not real files)
    Description string `json:"description"`
}
```

This struct defines the mapping between BLADE data types and Databricks tables:

### Fields:
- **DataType**: The identifier for the BLADE data category (e.g., "maintenance", "sortie")
- **TableName**: The Databricks Delta table name where this data will be stored
- **SourcePath**: A mock URI indicating the data source (uses "mock://" scheme for POC)
- **Description**: Human-readable description of what this data type contains

### JSON Tags:
Each field has JSON tags for potential serialization needs:
- Allows mappings to be exported as JSON configuration
- Enables future REST API integration
- Supports configuration file loading

## GetBLADEMappings Function (lines 12-39)

```go
func GetBLADEMappings() []BLADEDataMapping {
    return []BLADEDataMapping{
        // ... mappings
    }
}
```

This function returns all available BLADE data type mappings. It serves as the single source of truth for supported data types.

### Mapping 1: Maintenance Data (lines 14-19)
```go
{
    DataType:    "maintenance",
    TableName:   "blade_maintenance_data",
    SourcePath:  "mock://maintenance",
    Description: "Aircraft maintenance schedules and predictive maintenance data",
}
```

**Purpose**: Tracks aircraft maintenance activities
- **Data Type**: "maintenance" - used as command-line argument
- **Table Name**: "blade_maintenance_data" - follows naming convention with prefix
- **Mock Path**: Indicates this is mock data for POC
- **Content**: Maintenance schedules, repair records, predictive maintenance alerts

### Mapping 2: Sortie Data (lines 20-25)
```go
{
    DataType:    "sortie",
    TableName:   "blade_sortie_schedules",
    SourcePath:  "mock://sortie", 
    Description: "Flight schedules and sortie planning data",
}
```

**Purpose**: Manages flight operations
- **Data Type**: "sortie" - military term for a combat mission
- **Table Name**: "blade_sortie_schedules" - specific to flight scheduling
- **Content**: Flight schedules, mission planning, training exercises

### Mapping 3: Deployment Data (lines 26-31)
```go
{
    DataType:    "deployment",
    TableName:   "blade_deployment_plans",
    SourcePath:  "mock://deployment", 
    Description: "Deployment preparation and logistics planning",
}
```

**Purpose**: Handles deployment operations
- **Data Type**: "deployment" - for troop and equipment movements
- **Table Name**: "blade_deployment_plans" - focused on planning aspect
- **Content**: Personnel deployments, equipment transfers, humanitarian operations

### Mapping 4: Logistics Data (lines 32-37)
```go
{
    DataType:    "logistics",
    TableName:   "blade_logistics_general",
    SourcePath:  "mock://logistics",
    Description: "General logistics and supply chain data",
}
```

**Purpose**: General supply chain management
- **Data Type**: "logistics" - broad category for supplies
- **Table Name**: "blade_logistics_general" - catch-all for logistics data
- **Content**: Supply requests, fuel management, munitions tracking, equipment inventory

## Design Decisions

### 1. Table Naming Convention
All tables follow the pattern: `blade_{datatype}_{suffix}`
- **Prefix**: "blade_" identifies these as BLADE-sourced tables
- **Data Type**: Matches the command-line argument
- **Suffix**: Describes the specific aspect (data, schedules, plans, general)

### 2. Mock URI Scheme
Uses "mock://" prefix for source paths:
- Clearly indicates POC/test data
- Prevents confusion with real file paths
- Easy to replace with real URIs in production

### 3. Static Configuration
Mappings are hardcoded rather than loaded from config:
- Simplifies POC implementation
- Ensures consistency across runs
- Easy to extend by adding new entries

### 4. Descriptive Metadata
Each mapping includes a description:
- Documents the purpose of each data type
- Helps users understand what data to expect
- Useful for generating documentation

## Usage in the System

1. **Adapter Initialization**: The adapter loads these mappings into a map for fast lookup
2. **Validation**: Used to validate command-line arguments against supported types
3. **Table Creation**: Table names are used in CREATE TABLE statements
4. **Metadata**: Descriptions are included in ingestion metadata

## Extending the Mappings

To add a new BLADE data type:

1. Add a new entry to the returned slice:
```go
{
    DataType:    "weather",
    TableName:   "blade_weather_data",
    SourcePath:  "mock://weather",
    Description: "Weather conditions affecting operations",
}
```

2. Create corresponding mock data file:
   - `mock_blade_data/weather/weather_data.json`

3. The adapter will automatically support the new type

## Future Enhancements

In a production system, these mappings might:
1. Be loaded from a configuration file
2. Include additional metadata (schema version, update frequency)
3. Support multiple source paths per data type
4. Include data validation rules
5. Define custom table properties per type