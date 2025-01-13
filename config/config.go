package config

type Config struct {
	AccessKeyID     string
	AccessKeySecret string
	Region          string
}

func LoadConfig() *Config {
	// 这里可以从环境变量、配置文件或命令行参数加载
	return &Config{
		AccessKeyID:     "your-access-key-id",
		AccessKeySecret: "your-access-key-secret",
		Region:          "cn-hangzhou",
	}
}
