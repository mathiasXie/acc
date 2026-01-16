package acc

import (
	"fmt"
	"testing"
	"time"

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

func initConfig(t *testing.T) {

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=True&loc=Local", "root", "12345678", "127.0.0.1", "3306", "config_center")

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}

	// Call Init to initialize the database and start the refreshConfig goroutine
	Init(db, "asr")
}

func TestGetConfig(t *testing.T) {
	initConfig(t)
	time.Sleep(2 * time.Second)
	config, err := GetConfig[CozeConfig]("coze_auth_config")
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}

	t.Logf("config: %+v\n", config.AppID)

	//RemoveConfig("coze_auth_config")
}

func TestInit(t *testing.T) {
	initConfig(t)
	config := CozeConfig{
		BotID:     "ccccccccccc",
		PublicKey: "1234567890",
		AppID:     "1234567890",
	}
	version, err := SaveConfig("coze_auth_config", "coze_auth_config", config, "扣子Auth授权配置", "xiezhengdong")
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	} else {
		t.Logf("save config success")
	}
	t.Logf("version is %d", version)

	userRoleConfig := UserRoleConfig{
		LLM:    "gpt-4o",
		User:   "user1",
		Role:   "admin",
		Speech: "admin",
	}

	operator := "user1"
	_, err = SaveConfig("user_role_config", "user_role_config", userRoleConfig, "用户角色配置", operator)
	if err != nil {
		t.Fatalf("Failed to save user role config: %v", err)
	} else {
		t.Logf("save user role config success")
	}

	_ = EnableConfig("coze_auth_config", version)
	cloudConfig, err := GetConfig[CozeConfig]("coze_auth_config")
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}

	t.Logf("config: %+v\n", cloudConfig.AppID)

	//RemoveConfig("coze_auth_config")
}
