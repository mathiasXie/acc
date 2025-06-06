package cloud_config

import (
	"encoding/json"
	"fmt"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type CozeConfig struct {
	BotID     string `json:"bot_id"`
	PublicKey string `json:"public_key"`
	AppID     string `json:"app_id"`
}

type UserRoleConfig struct {
	LLM    string `json:"llm"`
	User   string `json:"user"`
	Role   string `json:"role"`
	Speech string `json:"speech"`
}

func TestInit(t *testing.T) {

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=True&loc=Local", "root", "12345678", "127.0.0.1", "3306", "config_center")

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}

	// Call Init to initialize the database and start the refreshConfig goroutine
	Init(db, "asr-server")

	config := CozeConfig{
		BotID:     "1234567890",
		PublicKey: "1234567890",
		AppID:     "1234567890",
	}
	jsonConfig, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}
	err = SaveConfig("coze_auth_config", "coze_auth_config", string(jsonConfig), "扣子Auth授权配置")
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	} else {
		t.Logf("save config success")
	}

	userRoleConfig := UserRoleConfig{
		LLM:    "gpt-4o",
		User:   "user1",
		Role:   "admin",
		Speech: "admin",
	}
	jsonUserRoleConfig, err := json.Marshal(userRoleConfig)
	if err != nil {
		t.Fatalf("Failed to marshal user role config: %v", err)
	}
	err = SaveConfig("user_role_config", "user_role_config", string(jsonUserRoleConfig), "用户角色配置")
	if err != nil {
		t.Fatalf("Failed to save user role config: %v", err)
	} else {
		t.Logf("save user role config success")
	}

	cloudConfig, err := GetConfig[CozeConfig]("coze_auth_config")
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}

	t.Logf("config: %+v\n", cloudConfig)

	//RemoveConfig("coze_auth_config")
}
