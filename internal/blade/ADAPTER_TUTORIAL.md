# adapter.go Tutorial

This file implements the BLADE adapter that bridges between BLADE data types and Databricks ingestion requirements. It handles data type mapping, file loading, and request preparation.

## Package Declaration and Imports

```go
package blade

import (
    "fmt"
    "io/ioutil"
    "path/filepath"
    "databricks-blade-poc/internal/databricks"
)
```

- **fmt**: For error formatting and string operations
- **io/ioutil**: For reading mock data files from disk
- **path/filepath**: For constructing file paths in a platform-independent way
- **internal/databricks**: To use the IngestionRequest type

## BLADEAdapter Struct (lines 10-14)

```go
type BLADEAdapter struct {
    dataSource string // a specific BLADE deployment
    basePath string // the root volume path where BLADE stores data files
    mappings map[string]BLADEDataMapping // map of data type -> table configuration (for quick lookup)
}
```

The adapter maintains:
- **dataSource**: Identifies which BLADE deployment this is (e.g., "BLADE_LOGISTICS")
- **basePath**: Root directory containing mock data files (e.g., "mock_blade_data/")
- **mappings**: Pre-loaded map for O(1) lookup of data type configurations

## NewBLADEAdapter Constructor (lines 17-30)

```go
func NewBLADEAdapter(dataSource, basePath string) *BLADEAdapter {
    mappings := make(map[string]BLADEDataMapping)

    // loads all available BLADE mappings, indexing by BLADE data type
    for _, mapping := range GetBLADEMappings() {
        mappings[mapping.DataType] = mapping
    }

    return &BLADEAdapter{
        dataSource: dataSource,
        basePath:   basePath,
        mappings:   mappings,
    }
}
```

The constructor:
1. **Creates empty map** for storing data type mappings
2. **Loads all mappings** from `GetBLADEMappings()` (defined in models.go)
3. **Indexes by data type** for fast lookup (e.g., mappings["maintenance"])
4. **Returns configured adapter** ready to process requests

## PrepareIngestionRequest Method (lines 34-63)

This is the core method that transforms a data type into an ingestion request:

### 1. Data Type Validation (lines 36-40)
```go
mapping, exists := b.mappings[dataType]

if !exists {
    return nil, fmt.Errorf("Unsupported BLADE data type: %s", dataType)
}
```
- Looks up the data type in the mappings
- Returns error if data type is not supported
- Provides clear error message for debugging

### 2. Mock Data Loading (lines 43-46)
```go
sampleData, err := b.loadMockDataFile(dataType)
if err != nil {
    return nil, fmt.Errorf("failed to load mock data for %s: %w", dataType, err)
}
```
- Delegates to `loadMockDataFile` to read JSON from disk
- Wraps any file reading errors with context

### 3. Request Construction (lines 48-62)
```go
return &databricks.IngestionRequest{
    TableName:     mapping.TableName,
    SourcePath:    "mock://" + dataType, // Mock path for POC
    FileFormat:    "JSON",
    FormatOptions: "'multiLine' = 'true', 'inferSchema' = 'true'",
    DataSource:    b.dataSource,
    SampleData:    sampleData, // Add the loaded mock data
    Metadata: map[string]string{
        "source_system": "BLADE",
        "data_type":     dataType,
        "integration":   "databricks_poc",
        "description":   mapping.Description,
        "mode":          "mock_data",
    },
}, nil
```

Creates a complete ingestion request with:
- **TableName**: From the mapping (e.g., "blade_maintenance_data")
- **SourcePath**: Mock URI scheme to indicate POC mode
- **FileFormat**: Always JSON for mock data
- **FormatOptions**: Spark options for multiline JSON with schema inference
- **DataSource**: The BLADE deployment identifier
- **SampleData**: The actual JSON data loaded from file
- **Metadata**: Rich metadata for tracking and auditing

## GetSupportedDataTypes Method (lines 66-72)

```go
func (b *BLADEAdapter) GetSupportedDataTypes() []string {
    types := make([]string, 0, len(b.mappings))
    for dataType := range b.mappings {
        types = append(types, dataType)
    }
    return types
}
```

Utility method that:
1. **Creates slice** with capacity hint for efficiency
2. **Iterates map keys** to collect all data types
3. **Returns list** of supported types (e.g., ["maintenance", "sortie", "deployment", "logistics"])

Used for validation and displaying available options to users.

## loadMockDataFile Method (lines 75-85)

```go
func (b *BLADEAdapter) loadMockDataFile(dataType string) (string, error) {
    fileName := fmt.Sprintf("%s_data.json", dataType)
    filePath := filepath.Join(b.basePath, dataType, fileName)
    
    data, err := ioutil.ReadFile(filePath)
    if err != nil {
        return "", fmt.Errorf("failed to read mock data file %s: %w", filePath, err)
    }
    
    return string(data), nil
}
```

Private method that:
1. **Constructs filename** using pattern: `{dataType}_data.json`
2. **Builds full path**: Joins base path, data type subdirectory, and filename
3. **Reads file**: Uses ioutil.ReadFile to load entire file
4. **Returns string**: Converts bytes to string for JSON data

### Expected File Structure:
```
mock_blade_data/
├── maintenance/
│   └── maintenance_data.json
├── sortie/
│   └── sortie_data.json
├── deployment/
│   └── deployment_data.json
└── logistics/
    └── logistics_data.json
```

## Design Patterns

1. **Adapter Pattern**: Converts between BLADE concepts and Databricks requirements
2. **Factory Method**: `PrepareIngestionRequest` creates configured request objects
3. **Map-based Registry**: Pre-loads mappings for O(1) lookup performance
4. **Error Wrapping**: Uses `%w` verb to maintain error chain
5. **Mock Data Strategy**: Loads test data from files for POC functionality

## Usage Flow

1. **Initialization**:
   ```go
   adapter := NewBLADEAdapter("BLADE_LOGISTICS", "mock_blade_data/")
   ```

2. **Check Supported Types**:
   ```go
   types := adapter.GetSupportedDataTypes()
   // Returns: ["maintenance", "sortie", "deployment", "logistics"]
   ```

3. **Prepare Request**:
   ```go
   req, err := adapter.PrepareIngestionRequest("maintenance")
   // Returns configured IngestionRequest with mock data
   ```

## Error Handling

The adapter provides clear error messages for common issues:
- Unsupported data type: "Unsupported BLADE data type: {type}"
- File not found: "failed to read mock data file {path}: {error}"
- Data loading failure: "failed to load mock data for {type}: {error}"