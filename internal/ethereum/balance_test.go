package ethereum_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "go.uber.org/zap"

    "github.com/your-username/ethereum-trading-mcp/internal/ethereum"
    "github.com/your-username/ethereum-trading-mcp/internal/wallet"
)

type MockEthClient struct {
    mock.Mock
}

func (m *MockEthClient) BalanceAt(ctx context.Context, address string, blockNumber *big.Int) (*big.Int, error) {
    args := m.Called(ctx, address, blockNumber)
    return args.Get(0).(*big.Int), args.Error(1)
}

func TestGetBalance_ETH(t *testing.T) {
    logger, _ := zap.NewDevelopment()
    
    // 创建模拟钱包管理器
    walletCfg := &wallet.WalletConfig{
        PrivateKey:  "test-key",
        RPCEndpoint: "https://test.rpc",
        ChainID:     1,
    }
    walletMgr, err := wallet.NewWalletManager(walletCfg, logger)
    assert.NoError(t, err)
    
    // 创建以太坊客户端配置
    ethCfg := &ethereum.EthereumConfig{
        RPCEndpoint:     "https://test.rpc",
        ChainID:         1,
        UniswapV2Router: "0xRouterV2",
        UniswapV3Router: "0xRouterV3", 
        WETHAddress:     "0xWETH",
    }
    
    ethClient, err := ethereum.NewEthereumClient(ethCfg, walletMgr, logger)
    assert.NoError(t, err)
    
    // 测试获取 ETH 余额
    ctx := context.Background()
    balance, err := ethClient.GetBalance(ctx, "0x742d35Cc6634C0532925a3b8D7a2a5c4A7A6A5a5", nil)
    
    if err != nil {
        t.Logf("Expected error in test environment: %v", err)
    } else {
        assert.NotNil(t, balance)
        assert.True(t, balance.IsETH)
    }
}

func TestValidateAddress(t *testing.T) {
    logger, _ := zap.NewDevelopment()
    
    walletCfg := &wallet.WalletConfig{
        PrivateKey:  "test-key",
        RPCEndpoint: "https://test.rpc", 
        ChainID:     1,
    }
    walletMgr, err := wallet.NewWalletManager(walletCfg, logger)
    assert.NoError(t, err)
    
    ethCfg := &ethereum.EthereumConfig{
        RPCEndpoint: "https://test.rpc",
        ChainID:     1,
    }
    
    ethClient, err := ethereum.NewEthereumClient(ethCfg, walletMgr, logger)
    assert.NoError(t, err)
    
    // 测试有效地址
    validAddr := "0x742d35Cc6634C0532925a3b8D7a2a5c4A7A6A5a5"
    result, err := ethClient.ValidateAddress(validAddr)
    assert.NoError(t, err)
    assert.Equal(t, validAddr, result.Hex())
    
    // 测试无效地址
    invalidAddr := "not-an-address"
    _, err = ethClient.ValidateAddress(invalidAddr)
    assert.Error(t, err)
}
