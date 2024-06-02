package initialization

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	FeishuAppId             string `mapstructure:"FEISHU_APP_ID"`
	FeishuAppSecret         string `mapstructure:"FEISHU_APP_SECRET"`
	FeishuEncryptKey        string `mapstructure:"FEISHU_ENCRYPT_KEY"`
	FeishuVerificationToken string `mapstructure:"FEISHU_VERIFICATION_TOKEN"`
	OpenaiApiKey            string `mapstructure:"OPENAI_KEY"`
	OpenaiModel             string `mapstructure:"OPENAI_MODEL"`
	OpenaiMaxTokens         int    `mapstructure:"OPENAI_MAX_TOKENS"`
	OpenaiApiUrl            string `mapstructure:"OPENAI_API_URL"`
	HttpProxy               string `mapstructure:"HTTP_PROXY"`
	AzureOn                 bool   `mapstructure:"AZURE_ON"`
	AzureDeploymentName     string `mapstructure:"AZURE_DEPLOYMENT_NAME"`
	AzureEndpoint           string `mapstructure:"AZURE_ENDPOINT"`
	AzureOpenaiToken        string `mapstructure:"AZURE_OPENAI_TOKEN"`
}

var (
	cfg    = pflag.StringP("config", "c", "./config.yaml", "apiserver config file path.")
	config *Config
	once   sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		config = LoadConfig(*cfg)
	})

	return config
}

func LoadConfig(cfg string) *Config {
	viper.SetConfigFile(cfg)
	viper.ReadInConfig()
	viper.AutomaticEnv()
	//content, err := ioutil.ReadFile("config.yaml")
	//if err != nil {
	//	fmt.Println("Error reading file:", err)
	//}
	//fmt.Println(string(content))

	config := &Config{
		FeishuAppId:             getViperStringValue("APP_ID", ""),
		FeishuAppSecret:         getViperStringValue("APP_SECRET", ""),
		FeishuEncryptKey:        getViperStringValue("APP_ENCRYPT_KEY", ""),
		FeishuVerificationToken: getViperStringValue("APP_VERIFICATION_TOKEN", ""),
		OpenaiApiKey:            getViperStringValue("OPENAI_KEY", ""),
		OpenaiModel:             getViperStringValue("OPENAI_MODEL", "gpt-4o"),
		OpenaiMaxTokens:         getViperIntValue("OPENAI_MAX_TOKENS", 2000),
		OpenaiApiUrl:            getViperStringValue("API_URL", "https://api.openai.com"),
		HttpProxy:               getViperStringValue("HTTP_PROXY", ""),
		AzureOn:                 getViperBoolValue("AZURE_ON", false),
		AzureDeploymentName:     getViperStringValue("AZURE_DEPLOYMENT_NAME", ""),
		AzureEndpoint:           getViperStringValue("AZURE_ENDPOINT", ""),
		AzureOpenaiToken:        getViperStringValue("AZURE_OPENAI_TOKEN", ""),
	}

	return config
}

func getViperStringValue(key string, defaultValue string) string {
	value := viper.GetString(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getViperIntValue(key string, defaultValue int) int {
	value := viper.GetString(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		fmt.Printf("Invalid value for %s, using default value %d\n", key, defaultValue)
		return defaultValue
	}
	return intValue
}

func getViperBoolValue(key string, defaultValue bool) bool {
	value := viper.GetString(key)
	if value == "" {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		fmt.Printf("Invalid value for %s, using default value %v\n", key, defaultValue)
		return defaultValue
	}
	return boolValue
}
