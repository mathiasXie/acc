# cloud_config

安装：
```bash
go get github.com/mathiasXie/cloud_config
```

cloud_config 是一个基于 Go + GORM 实现的通用配置管理库，适用于自定义结构体的配置项统一读写存储。

## 主要功能

- 支持多种结构体作为配置，支持泛型接口 `GetConfig[T]` 类型安全获取配置
- 支持配置多版本及启用回滚，支持多 namespace 隔离
- 配置项自动定时刷新，首次读取时自动解析为目标结构体并缓存，提升后续读取性能
- 提供配置的保存、启用、删除等常用接口

> **注意：无需手动建表，Init 时会自动用 GORM Migrator 检查并创建表结构**

## 常用接口示例

```go
// 定义配置结构体
 type CozeConfig struct {
     BotID     string `json:"bot_id"`
     PublicKey string `json:"public_key"`
     AppID     string `json:"app_id"`
 }

// 初始化
cloud_config.Init(db, "my-namespace")

// 保存配置
cfg := CozeConfig{BotID: "xxx", PublicKey: "yyy", AppID: "zzz"}
cfgJson, _ := json.Marshal(cfg)
cloud_config.SaveConfig("coze_auth_config", "描述", string(cfgJson), "备注")

// 启用某个版本
cloud_config.EnableConfig("coze_auth_config", version)

// 类型安全获取配置
cozeCfg, err := cloud_config.GetConfig[CozeConfig]("coze_auth_config")
```

## 运行机制简述

- 配置项在启用后，后台协程每分钟自动刷新至内存
- 读取配置时，若已缓存目标类型则直接返回，否则解析后缓存
- 支持多 namespace 多版本，实现多业务隔离与灵活切换
