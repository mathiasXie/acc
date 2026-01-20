[中文](README.md)

# acc - A Dynamic Configuration Center

`acc` is a lightweight dynamic configuration center written in Go. It uses a relational database (like MySQL) to store and manage configurations, and provides a simple API for applications to dynamically fetch and update configurations at runtime.

## Features

- **Dynamic Configuration**: Configurations are stored in a database and can be updated in real-time without restarting the application.
- **Versioning**: Each configuration change creates a new version, allowing for easy rollbacks and tracking of configuration history.
- **Namespace Isolation**: Configurations are isolated by namespaces, allowing multiple applications or environments to share the same configuration center.
- **Automatic Refresh**: The configuration is automatically refreshed periodically to ensure that the latest configuration is always used.
- **Typed Configuration**: Supports fetching configurations as specific Go types, automatically handling JSON deserialization.

## Installation

To use `acc` in your Go project, you can use `go get`:

```bash
go get github.com/mathiasXie/acc
```

## Usage

### Initialization

First, you need to initialize the configuration center with a `*gorm.DB` instance and a namespace.

```go
import (
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "github.com/mathiasXie/acc"
)

func main() {
    // Replace with your database connection string
    dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }

    // Initialize the configuration center with the "production" namespace
    acc.Init(db, "production")

    // Your application logic...
}
```

### Getting a Configuration

You can use the `GetConfig` function to fetch a configuration value. It supports generics, so you can get the configuration as a specific type.

```go
type MyConfig struct {
    Host  string `json:"host"`
    Port  int    `json:"port"`
    Debug bool   `json:"debug"`
}

func main() {
    // ... initialization

    // Get the configuration with the key "database"
    dbConfig, err := acc.GetConfig[MyConfig]("database")
    if err != nil {
        // Handle error
    }

    fmt.Printf("Database host: %s\n", dbConfig.Host)
    fmt.Printf("Database port: %d\n", dbConfig.Port)
}
```

### Saving a Configuration

You can use the `SaveConfig` function to create a new configuration or update an existing one. This will also create a new version of the configuration.

```go
func main() {
    // ... initialization

    newConfig := MyConfig{
        Host:  "localhost",
        Port:  5432,
        Debug: true,
    }

    // Save the configuration with the key "database"
    version, err := acc.SaveConfig("database", "Database Configuration", newConfig, "Initial database configuration", "admin")
    if err != nil {
        // Handle error
    }

    fmt.Printf("New configuration version: %d\n", version)
}
```

### Saving a New Version

To create a new version for an existing configuration, you can use the `SaveVersion` function.

```go
func main() {
    // ... initialization

    updatedConfig := MyConfig{
        Host:  "remote.host",
        Port:  5432,
        Debug: false,
    }

    // Save a new version for the "database" configuration
    version, err := acc.SaveVersion("database", updatedConfig, "admin")
    if err != nil {
        // Handle error
    }

    fmt.Printf("New configuration version: %d\n", version)
}
```

### Enabling a Configuration Version

After creating a new version, you need to enable it to make it active.

```go
func main() {
    // ... initialization

    // Enable version 2 of the "database" configuration
    err := acc.EnableConfig("database", 2)
    if err != nil {
        // Handle error
    }

    fmt.Println("Configuration version 2 enabled successfully")
}
```

## Data Model

The configuration center uses two tables to store the configuration data:

- **`acc_configs`**: Stores the metadata for each configuration, such as the namespace, key, and name.
- **`acc_versions`**: Stores the different versions of each configuration, including the configuration value (as a JSON string) and an `enabled` flag.

A `Config` can have multiple `Version`s, but only one version can be enabled at a time for a given configuration.
