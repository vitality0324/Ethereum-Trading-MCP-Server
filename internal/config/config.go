package config

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/spf13/viper"
)

type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Ethereum EthereumConfig `mapstructure:"ethereum"`
    Wallet   WalletConfig   `mapstructure:"wallet"`
    Logging  LoggingConfig  `mapstructure:"logging"`
}

type ServerConfig struct {
    Host string `mapstructure:"host"`
    Port int    `mapstructure:"port"`
}

type EthereumConfig struct {
    RPCEndpoint     string `mapstructure:"rpc_endpoint"`
    ChainID         int64  `mapstructure:"chain_id"`
    UniswapV2Router string `mapstructure:"uniswap_v2_router"`
    UniswapV3Router string `mapstructure:"uniswap_v3_router"`
    WETHAddress     string `mapstructure:"weth_address"`
}

type WalletConfig struct {
    PrivateKey string `mapstructure:"private_key"`
}

type LoggingConfig struct {
    Level string `mapstructure:"level"`
    File  string `mapstructure:"file"`
}

func Load() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")
    viper.AddConfigPath("./config")
    viper.AddConfigPath("/etc/ethereum-mcp/")

    // 设置默认值
    setDefaults()

    // 读取环境变量
    viper.AutomaticEnv()
    viper.SetEnvPrefix("MCP")

    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, fmt.Errorf("error reading config file: %w", err)
        }
    }

    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("error unmarshaling config: %w", err)
    }

    // 验证必需配置
    if err := validateConfig(&config); err != nil {
        return nil, err
    }

    return &config, nil
}

func setDefaults() {
    viper.SetDefault("server.host", "localhost")
    viper.SetDefault("server.port", 8080)
    viper.SetDefault("ethereum.chain_id", 1) // Mainnet
    viper.SetDefault("ethereum.uniswap_v2_router", "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D")
    viper.SetDefault("ethereum.uniswap_v3_router", "0xE592427A0AEce92De3Edee1F18E0157C05861564")
    viper.SetDefault("ethereum.weth_address", "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
    viper.SetDefault("logging.level", "info")
}

func validateConfig(config *Config) error {
    if config.Ethereum.RPCEndpoint == "" {
        return fmt.Errorf("ethereum rpc_endpoint is required")
    }
    if config.Wallet.PrivateKey == "" {
        return fmt.Errorf("wallet private_key is required")
    }
    return nil
}

func GetConfigPath() string {
    if path := os.Getenv("MCP_CONFIG_PATH"); path != "" {
        return path
    }
    return filepath.Join(".", "config.yaml")
}
