package ethereum

import (
    "context"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "go.uber.org/zap"

    "github.com/your-username/ethereum-trading-mcp/internal/wallet"
    "github.com/your-username/ethereum-trading-mcp/pkg/decimal"
)

type EthereumClient struct {
    client       *ethclient.Client
    walletMgr    *wallet.WalletManager
    logger       *zap.Logger
    config       *EthereumConfig
}

type EthereumConfig struct {
    RPCEndpoint     string
    ChainID         int64
    UniswapV2Router string
    UniswapV3Router string
    WETHAddress     string
}

func NewEthereumClient(cfg *EthereumConfig, walletMgr *wallet.WalletManager, logger *zap.Logger) (*EthereumClient, error) {
    client, err := ethclient.Dial(cfg.RPCEndpoint)
    if err != nil {
        return nil, err
    }

    return &EthereumClient{
        client:    client,
        walletMgr: walletMgr,
        logger:    logger,
        config:    cfg,
    }, nil
}

func (ec *EthereumClient) GetClient() *ethclient.Client {
    return ec.client
}

func (ec *EthereumClient) GetWalletManager() *wallet.WalletManager {
    return ec.walletMgr
}

func (ec *EthereumClient) GetChainID() *big.Int {
    return ec.walletMgr.ChainID
}

func (ec *EthereumClient) ValidateAddress(address string) (common.Address, error) {
    if !common.IsHexAddress(address) {
        return common.Address{}, fmt.Errorf("invalid Ethereum address: %s", address)
    }
    return common.HexToAddress(address), nil
}

func (ec *EthereumClient) GetGasPrice(ctx context.Context) (*big.Int, error) {
    return ec.client.SuggestGasPrice(ctx)
}

func (ec *EthereumClient) EstimateGas(ctx context.Context, from common.Address, to *common.Address, value *big.Int, data []byte) (uint64, error) {
    msg := ethereum.CallMsg{
        From:  from,
        To:    to,
        Value: value,
        Data:  data,
    }
    return ec.client.EstimateGas(ctx, msg)
}

func (ec *EthereumClient) Close() {
    if ec.client != nil {
        ec.client.Close()
    }
    if ec.walletMgr != nil {
        ec.walletMgr.Close()
    }
}
