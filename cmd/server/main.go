package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"

    "go.uber.org/zap"

    "github.com/your-username/ethereum-trading-mcp/internal/config"
    "github.com/your-username/ethereum-trading-mcp/internal/ethereum"
    "github.com/your-username/ethereum-trading-mcp/internal/mcp"
    "github.com/your-username/ethereum-trading-mcp/internal/wallet"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    logger, err := initLogger(cfg)
    if err != nil {
        log.Fatalf("Failed to initialize logger: %v", err)
    }
    defer logger.Sync()

    logger.Info("Starting Ethereum Trading MCP Server")

    // 初始化钱包
    walletCfg := &wallet.WalletConfig{
        PrivateKey:  cfg.Wallet.PrivateKey,
        RPCEndpoint: cfg.Ethereum.RPCEndpoint,
        ChainID:     cfg.Ethereum.ChainID,
    }
    walletMgr, err := wallet.NewWalletManager(walletCfg, logger)
    if err != nil {
        logger.Fatal("Failed to initialize wallet manager", zap.Error(err))
    }
    defer walletMgr.Close()

    logger.Info("Wallet initialized", 
        zap.String("address", walletMgr.GetAddress().Hex()))

    // 初始化以太坊客户端
    ethCfg := &ethereum.EthereumConfig{
        RPCEndpoint:     cfg.Ethereum.RPCEndpoint,
        ChainID:         cfg.Ethereum.ChainID,
        UniswapV2Router: cfg.Ethereum.UniswapV2Router,
        UniswapV3Router: cfg.Ethereum.UniswapV3Router,
        WETHAddress:     cfg.Ethereum.WETHAddress,
    }
    ethClient, err := ethereum.NewEthereumClient(ethCfg, walletMgr, logger)
    if err != nil {
        logger.Fatal("Failed to initialize Ethereum client", zap.Error(err))
    }
    defer ethClient.Close()

    // 初始化 MCP 服务器
    mcpServer := mcp.NewMCPServer(ethClient, logger)

    // 设置信号处理
    setupSignalHandler(logger, ethClient)

    // 启动 MCP 服务器
    if err := mcpServer.Start(); err != nil {
        logger.Fatal("MCP server failed", zap.Error(err))
    }

    logger.Info("MCP server stopped")
}

func initLogger(cfg *config.Config) (*zap.Logger, error) {
    var zapConfig zap.Config
    
    if cfg.Logging.Level == "debug" {
        zapConfig = zap.NewDevelopmentConfig()
    } else {
        zapConfig = zap.NewProductionConfig()
    }

    if cfg.Logging.File != "" {
        zapConfig.OutputPaths = []string{cfg.Logging.File, "stdout"}
        zapConfig.ErrorOutputPaths = []string{cfg.Logging.File, "stderr"}
    }

    return zapConfig.Build()
}

func setupSignalHandler(logger *zap.Logger, ethClient *ethereum.EthereumClient) {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-c
        logger.Info("Received shutdown signal")
        ethClient.Close()
        os.Exit(0)
    }()
}
