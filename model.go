package cloud_config

import (
	"time"

	"gorm.io/gorm"
)

func (CloudConfig) TableName() string {
	return "cloud_configs"
}

type CloudConfig struct {
	Id          int64  `gorm:"column:id;primaryKey" json:"id"`
	Namespace   string `gorm:"column:namespace;type:varchar(180);index:idx_namespace_config_key,unique" json:"namespace"`
	ConfigKey   string `gorm:"column:config_key;type:varchar(180);index:idx_namespace_config_key,unique" json:"config_key"`
	ConfigName  string `gorm:"column:config_name;type:varchar(180)" json:"config_name"`
	ConfigValue string `gorm:"column:config_value;type:longtext" json:"config_value"`
	Description string `gorm:"column:description;type:varchar(180)" json:"description"`

	Version int64 `gorm:"column:version;type:bigint;index:idx_namespace_config_key,unique" json:"version"`
	Enabled bool  `gorm:"column:enabled;type:tinyint(1);default:0" json:"enabled"`

	CreatedAt *time.Time      `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt *time.Time      `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt *gorm.DeletedAt `gorm:"column:deleted_at;type:timestamp;index:idx_namespace_config_key,unique" json:"deleted_at"`
}

func NewCloudConfig() *CloudConfig {
	return &CloudConfig{}
}
