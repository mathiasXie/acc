package cloud_config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

var (
	db             *gorm.DB
	configMap      = make(map[string]string)
	configInstance = make(map[string]interface{})
	configLock     sync.RWMutex
	namespace      string
)

func Init(configDB *gorm.DB, configNamespace string) {
	db = configDB
	namespace = configNamespace
	// Check if the table exists. If it doesn't exist, create it.
	if !db.Migrator().HasTable(&CloudConfig{}) {
		if err := db.Debug().Migrator().CreateTable(&CloudConfig{}); err != nil {
			log.Fatalf("Failed to create cloud_configs table: %v", err)
		}
		log.Println("cloud_configs table created")
	}

	go loadConfigFromDB()
	// Timed refresh configuration
	go refreshConfig()
}

func loadConfigFromDB() {
	var configs []CloudConfig
	result := db.Where("namespace=? AND enabled=?", namespace, true).Find(&configs)
	if result.Error != nil {
		log.Fatalf("Failed to query cloud_configs table: %v", result.Error)
	}

	configLock.Lock()
	defer configLock.Unlock()

	for _, config := range configs {
		// 配置发生过变更
		if configMap[config.ConfigKey] != config.ConfigValue {
			delete(configInstance, config.ConfigKey)
		}
		configMap[config.ConfigKey] = config.ConfigValue

	}
}

func refreshConfig() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		loadConfigFromDB()
	}
}

func GetConfig[T any](key string) (T, error) {

	if configInstance[key] != nil {
		return configInstance[key].(T), nil
	}

	config, ok := configMap[key]
	if !ok {
		fmt.Printf("Config with key '%s' not found\n", key)
		var zero T
		return zero, fmt.Errorf("config with key '%s' not found", key)
	}

	// 将 JSON 字符串反序列化为泛型类型 T
	var result T
	if err := json.Unmarshal([]byte(config), &result); err != nil {
		fmt.Printf("Failed to unmarshal config to type %T: %v\n", result, err)
		var zero T
		return zero, fmt.Errorf("failed to unmarshal config to type %T: %v", result, err)
	}
	configLock.Lock()
	defer configLock.Unlock()
	configInstance[key] = result

	return result, nil
}

func SaveConfig(key, name string, data interface{}, description string) (int64, error) {
	// Check if the config already exists
	var existingConfig CloudConfig
	cfgModel := &CloudConfig{}
	existErr := db.Where("namespace = ? and config_key = ?", namespace, key).Order("id desc").First(&existingConfig).Error
	if existErr != nil {
		if errors.Is(existErr, gorm.ErrRecordNotFound) {
			cfgModel.Version = 1
		} else {
			return 0, existErr
		}
	} else {
		cfgModel.Version = existingConfig.Version + 1
	}
	cfgStr, _ := json.Marshal(data)
	cfgModel.ConfigValue = string(cfgStr)
	cfgModel.ConfigKey = key
	cfgModel.Enabled = false
	cfgModel.Namespace = namespace
	cfgModel.ConfigName = name
	cfgModel.Description = description

	result := db.Save(cfgModel)
	if result.Error != nil {
		return 0, result.Error
	}
	//configMap[key] = data
	return cfgModel.Version, nil
}

func EnableConfig(configKey string, version int64) error {
	configLock.Lock()
	defer configLock.Unlock()

	var configContent *CloudConfig
	txErr := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("namespace = ? AND config_key = ? AND version = ?", namespace, configKey, version).First(&configContent).Error
		if err != nil {
			log.Printf("Failed to query config from the database: %v", err)
			return err
		}
		//disable all config
		err = tx.Model(&CloudConfig{}).Where("namespace = ? AND config_key = ?", namespace, configKey).
			Update("enabled", false).Error
		if err != nil {
			log.Printf("Failed to update config in the database: %v", err)
			return err
		}
		//enable current version
		err = tx.Model(&CloudConfig{}).Where("id = ?", configContent.Id).
			Update("enabled", true).Error
		if err != nil {
			log.Printf("Failed to update config in the database: %v", err)
			return err
		}
		return nil
	})
	if txErr != nil {
		return txErr
	}
	configJSON, _ := json.Marshal(configContent)
	configMap[configKey] = string(configJSON)
	return nil
}

func RemoveConfig(key string) {
	configLock.Lock()
	defer configLock.Unlock()

	// Soft delete config
	result := db.Where("namespace = ? AND config_key = ?", namespace, key).Delete(&CloudConfig{})
	if result.Error != nil {
		log.Fatalf("Failed to delete config in the database: %v", result.Error)
	}
	delete(configMap, key)
}
