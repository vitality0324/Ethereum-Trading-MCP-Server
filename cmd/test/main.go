package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "go.uber.org/zap"

    "github.com/your-username/ethereum-trading-mcp/internal/config"
    "github.com/your-username/ethereum-trading-mcp/internal/ethereum"
    "github.com/your-username/ethereum-trading-mcp/internal/wallet"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    logger, err := zap.NewDevelopment()
    if err != nil {
        log.Fatalf("Failed to initialize logger: %v", err)
    }
    defer logger.Sync()

    logger.Info("Starting integration test")

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

    // init以太坊客户端
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

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // test余额查询
    testBalanceQuery(ctx, ethClient, walletMgr.GetAddress().Hex())

    // test价格查询
    testPriceQuery(ctx, ethClient)

    // test交换模拟
    testSwapSimulation(ctx, ethClient)

    logger.Info("Integration test completed successfully")
}

func testBalanceQuery(ctx context.Context, ethClient *ethereum.EthereumClient, address string) {
    fmt.Printf("\n=== Testing Balance Query ===\n")
    
    // 查询ETH余额
    balance, err := ethClient.GetBalance(ctx, address, nil)
    if err != nil {
        fmt.Printf("Error getting ETH balance: %v\n", err)
        return
    }
    
    balanceJSON, _ := json.MarshalIndent(balance, "", "  ")
    fmt.Printf("ETH Balance: %s\n", string(balanceJSON))
}

func testPriceQuery(ctx context.Context, ethClient *ethereum.EthereumClient) {
    fmt.Printf("\n=== Testing Price Query ===\n")
    
    // 查询ETH价格
    price, err := ethClient.GetTokenPrice(ctx, "ETH")
    if err != nil {
        fmt.Printf("Error getting ETH price: %v\n", err)
        return
    }
    
    priceJSON, _ := json.MarshalIndent(price, "", "  ")
    fmt.Printf("ETH Price: %s\n", string(priceJSON))
    
    // 查询USDC 价格
    price, err = ethClient.GetTokenPrice(ctx, "USDC")
    if err != nil {
        fmt.Printf("Error getting USDC price: %v\n", err)
        return
    }
    
    priceJSON, _ = json.MarshalIndent(price, "", "  ")
    fmt.Printf("USDC Price: %s\n", string(priceJSON))
}

func testSwapSimulation(ctx context.Context, ethClient *ethereum.EthereumClient) {
    fmt.Printf("\n=== Testing Swap Simulation ===\n")
    
    req := &ethereum.SwapRequest{
        FromToken:       "ETH",
        ToToken:         "USDC",
        Amount:          decimal.NewFromFloat(0.1), 
        SlippageTolerance: decimal.NewFromFloat(0.01), 
        UseV3:           false,
    }
    
    result, err := ethClient.SwapTokens(ctx, req)
    if err != nil {
        fmt.Printf("Error simulating swap: %v\n", err)
        return
    }
    
    resultJSON, _ := json.MarshalIndent(result, "", "  ")
    fmt.Printf("Swap Result: %s\n", string(resultJSON))
}
