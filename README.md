[English](readme_en.md)

# acc - 动态配置中心

`acc` 是一个使用 Go 语言编写的轻量级动态配置中心。它使用关系型数据库（如 MySQL）来存储和管理配置，并提供简单的 API，方便应用程序在运行时动态获取和更新配置。

## 功能特性

- **动态配置**: 配置存储在数据库中，可以实时更新而无需重启应用程序。
- **版本控制**: 每次配置变更都会创建一个新版本，方便回滚和追溯配置历史。
- **命名空间隔离**: 配置按命名空间进行隔离，允许多个应用程序或环境共享同一个配置中心。
- **自动刷新**: 定期自动刷新配置，确保始终使用最新的配置。
- **类型化配置**: 支持将配置获取为特定的 Go 类型，自动处理 JSON 反序列化。

## 安装

在你的 Go 项目中使用 `acc`，可以通过 `go get` 命令：

```bash
go get github.com/mathiasXie/acc
```

## 使用方法

### 初始化

首先，你需要使用一个 `*gorm.DB` 实例和一个命名空间来初始化配置中心。

```go
import (
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "github.com/mathiasXie/acc"
)

func main() {
    // 替换为你的数据库连接字符串
    dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("无法连接数据库")
    }

    // 使用 "production" 命名空间初始化配置中心
    acc.Init(db, "production")

    // 你的应用程序逻辑...
}
```

### 获取配置

你可以使用 `GetConfig` 函数来获取配置值。它支持泛型，因此你可以将配置获取为特定的类型。

```go
type MyConfig struct {
    Host  string `json:"host"`
    Port  int    `json:"port"`
    Debug bool   `json:"debug"`
}

func main() {
    // ... 初始化

    // 获取键为 "database" 的配置
    dbConfig, err := acc.GetConfig[MyConfig]("database")
    if err != nil {
        // 处理错误
    }

    fmt.Printf("数据库主机: %s\n", dbConfig.Host)
    fmt.Printf("数据库端口: %d\n", dbConfig.Port)
}
```

### 保存配置

你可以使用 `SaveConfig` 函数来创建新配置或更新现有配置。这也将创建一个新的配置版本。

```go
func main() {
    // ... 初始化

    newConfig := MyConfig{
        Host:  "localhost",
        Port:  5432,
        Debug: true,
    }

    // 保存键为 "database" 的配置
    version, err := acc.SaveConfig("database", "数据库配置", newConfig, "初始数据库配置", "admin")
    if err != nil {
        // 处理错误
    }

    fmt.Printf("新配置版本: %d\n", version)
}
```

### 保存新版本

要为现有配置创建新版本，你可以使用 `SaveVersion` 函数。

```go
func main() {
    // ... 初始化

    updatedConfig := MyConfig{
        Host:  "remote.host",
        Port:  5432,
        Debug: false,
    }

    // 为 "database" 配置保存一个新版本
    version, err := acc.SaveVersion("database", updatedConfig, "admin")
    if err != nil {
        // 处理错误
    }

    fmt.Printf("新配置版本: %d\n", version)
}
```

### 启用配置版本

创建新版本后，你需要启用它才能使其生效。

```go
func main() {
    // ... 初始化

    // 启用 "database" 配置的版本 2
    err := acc.EnableConfig("database", 2)
    if err != nil {
        // 处理错误
    }

    fmt.Println("配置版本 2 已成功启用")
}
```

## 数据模型

配置中心使用两张表来存储配置数据：

- **`acc_configs`**: 存储每个配置的元数据，如命名空间、键和名称。
- **`acc_versions`**: 存储每个配置的不同版本，包括配置值（作为 JSON 字符串）和一个 `enabled` 标志。

一个 `Config` 可以有多个 `Version`，但对于给定的配置，一次只能启用一个版本。
