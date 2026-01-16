package acc

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
	if !db.Migrator().HasTable(&Config{}) {
		if err := db.Debug().Migrator().CreateTable(&Config{}); err != nil {
			log.Fatalf("Failed to create configs table: %v", err)
		}
		log.Println("configs table created")
	}

	if !db.Migrator().HasTable(&Version{}) {
		if err := db.Debug().Migrator().CreateTable(&Version{}); err != nil {
			log.Fatalf("Failed to create versions table: %v", err)
		}
		log.Println("versions table created")
	}

	go loadConfigFromDB()
	// Timed refresh configuration
	go refreshConfig()
}

func loadConfigFromDB() {
	var configs []Config
	result := db.
		Where("namespace=?", namespace).
		Preload("Version", "enabled=?", true).
		Find(&configs)
	if result.Error != nil {
		log.Fatalf("Failed to query configs table: %v", result.Error)
	}

	configLock.Lock()
	defer configLock.Unlock()

	for _, config := range configs {
		// 配置发生过变更
		if configMap[config.ConfigKey] != config.Version.ConfigValue {
			delete(configInstance, config.ConfigKey)
		}
		configMap[config.ConfigKey] = config.Version.ConfigValue
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
		var zero T
		return zero, fmt.Errorf("config with key '%s' not found", key)
	}
	// 将 JSON 字符串反序列化为泛型类型 T
	var result T
	if err := json.Unmarshal([]byte(config), &result); err != nil {
		var zero T
		return zero, fmt.Errorf("failed to unmarshal config to type %T: %v", result, err)
	}
	configLock.Lock()
	defer configLock.Unlock()
	configInstance[key] = result

	return result, nil
}

func SaveConfig(key, name string, data interface{}, description string, operator string) (int64, error) {
	// Check if the config already exists
	cfgModel := &Config{}
	existErr := db.
		Where("namespace = ? and config_key = ?", namespace, key).
		Order("id desc").
		Preload("Version", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("id DESC").Limit(1)
		}).
		First(&cfgModel).Error
	if existErr != nil && !errors.Is(existErr, gorm.ErrRecordNotFound) {
		return 0, existErr
	}

	cfgModel.ConfigKey = key
	cfgModel.Namespace = namespace
	cfgModel.ConfigName = name
	cfgModel.Operator = operator
	cfgModel.Description = description

	result := db.Save(cfgModel)
	if result.Error != nil {
		return 0, result.Error
	}

	cfgStr, err := json.Marshal(data)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal config data to JSON: %v", err)
	}
	versionModel := NewVersion()
	versionModel.Number = cfgModel.Version.Number + 1
	versionModel.ConfigValue = string(cfgStr)
	versionModel.Enabled = false
	versionModel.Operator = operator
	versionModel.ConfigID = cfgModel.Id
	result = db.Save(versionModel)
	if result.Error != nil {
		return 0, result.Error
	}
	return versionModel.Number, nil
}

func SaveVersion(key string, data interface{}, operator string) (int64, error) {
	// Check if the config already exists
	cfgModel := &Config{}
	existErr := db.
		Where("namespace = ? and config_key = ?", namespace, key).
		Preload("Version", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("id DESC").Limit(1)
		}).
		First(&cfgModel).Error
	if existErr != nil {
		if errors.Is(existErr, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("config with key '%s' not found", key)
		}
		return 0, existErr
	}

	cfgStr, err := json.Marshal(data)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal config data to JSON: %v", err)
	}
	versionModel := NewVersion()
	versionModel.Number = cfgModel.Version.Number + 1
	versionModel.ConfigValue = string(cfgStr)
	versionModel.Enabled = false
	versionModel.Operator = operator
	versionModel.ConfigID = cfgModel.Id
	result := db.Save(versionModel)
	if result.Error != nil {
		return 0, result.Error
	}
	return versionModel.Number, nil
}

func EnableConfig(configKey string, version int64) error {
	configLock.Lock()
	defer configLock.Unlock()

	var configContent *Config
	txErr := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("namespace = ? AND config_key = ?", namespace, configKey).First(&configContent).Error
		if err != nil {
			return fmt.Errorf("failed to query config from the database: %v", err)
		}
		//disable all config
		err = tx.Model(&Version{}).
			Where("config_id = ?", configContent.Id).
			Update("enabled", false).Error
		if err != nil {
			return fmt.Errorf("failed to update config in the database: %v", err)
		}
		//enable current version
		err = tx.Model(&Version{}).
			Where("config_id = ? AND number = ?", configContent.Id, version).
			Update("enabled", true).Error
		if err != nil {
			return fmt.Errorf("failed to update config in the database: %v", err)
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

func RemoveConfig(key string) error {
	configLock.Lock()
	defer configLock.Unlock()

	// Soft delete config
	result := db.Where("namespace = ? AND config_key = ?", namespace, key).Delete(&Config{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete config in the database: %v", result.Error)
	}
	delete(configMap, key)
	return nil
}

func GetConfigs() ([]Config, error) {
	var config []Config
	result := db.Where("namespace = ? ", namespace).Order("id desc").Find(&config)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query configs from the database: %v", result.Error)
	}
	return config, nil
}

func GetVersions(configID, page, size int64) ([]Version, int64, error) {
	var version []Version
	result := db.
		Where("config_id = ? ", configID).
		Order("id desc").
		Limit(int(size)).
		Offset(int((page - 1) * size)).
		Find(&version)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to query versions from the database: %v", result.Error)
	}
	total := int64(0)
	err := db.
		Model(&Version{}).
		Where("config_id = ? ", configID).
		Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query versions from the database: %v", err)
	}
	return version, total, nil
}

func GetEnabledVersion(configID int64) (*Version, error) {
	var version *Version
	result := db.
		Where("config_id = ? ", configID).
		Where("enabled = ? ", true).
		Find(&version)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query versions from the database: %v", result.Error)
	}
	total := int64(0)
	err := db.
		Model(&Version{}).
		Where("config_id = ? ", configID).
		Count(&total).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query versions from the database: %v", err)
	}
	return version, nil
}
