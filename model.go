package acc

import (
	"time"

	"gorm.io/gorm"
)

type Config struct {
	Id          int64           `gorm:"column:id;primaryKey" json:"id"`
	Namespace   string          `gorm:"column:namespace;type:varchar(180);index:idx_namespace_config_key,unique;" json:"namespace"`
	ConfigKey   string          `gorm:"column:config_key;type:varchar(180);index:idx_namespace_config_key,unique;" json:"config_key"`
	ConfigName  string          `gorm:"column:config_name;type:varchar(180)" json:"config_name"`
	Description string          `gorm:"column:description;type:varchar(180)" json:"description"`
	Operator    string          `gorm:"column:operator;type:varchar(180)" json:"operator"`
	CreatedAt   *time.Time      `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   *time.Time      `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt   *gorm.DeletedAt `gorm:"column:deleted_at;type:timestamp;" json:"deleted_at"`
	Version     Version         `gorm:"foreignKey:ConfigID;references:Id;constraint:OnDelete:CASCADE;" json:"versions"`
}

func (Config) TableName() string {
	return "acc_configs"
}

type Version struct {
	Id          int64           `gorm:"column:id;primaryKey" json:"id"`
	ConfigID    int64           `gorm:"column:config_id;type:bigint;index:idx_config_id_version,unique" json:"config_id"`
	Number      int64           `gorm:"column:number;type:bigint;index:idx_config_id_version,unique" json:"number"`
	Enabled     bool            `gorm:"column:enabled;type:tinyint(1);default:0;" json:"enabled"`
	ConfigValue string          `gorm:"column:config_value;type:longtext" json:"config_value"`
	Operator    string          `gorm:"column:operator;type:varchar(180)" json:"operator"`
	CreatedAt   *time.Time      `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   *time.Time      `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt   *gorm.DeletedAt `gorm:"column:deleted_at;type:timestamp;" json:"deleted_at"`
}

func (Version) TableName() string {
	return "acc_versions"
}

func NewVersion() *Version {
	return &Version{}
}

func NewConfig() *Config {
	return &Config{}
}
