package common

type ApiConfig struct {
	ApiUrl         string `yaml:"api_url"`
	ApiTokenId     string `yaml:"api_token_id"`
	ApiTokenSecret string `yaml:"api_token_secret"`
}
