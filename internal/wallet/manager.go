package wallet

import (
    "context"
    "crypto/ecdsa"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
    "go.uber.org/zap"
)

type WalletManager struct {
    privateKey *ecdsa.PrivateKey
    publicKey  *ecdsa.PublicKey
    address    common.Address
    client     *ethclient.Client
    chainID    *big.Int
    logger     *zap.Logger
}

type WalletConfig struct {
    PrivateKey string
    RPCEndpoint string
    ChainID    int64
}

func NewWalletManager(cfg *WalletConfig, logger *zap.Logger) (*WalletManager, error) {
    privateKey, err := crypto.HexToECDSA(cfg.PrivateKey)
    if err != nil {
        return nil, fmt.Errorf("invalid private key: %w", err)
    }

    publicKey := privateKey.Public()
    publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
    if !ok {
        return nil, fmt.Errorf("error casting public key to ECDSA")
    }

    address := crypto.PubkeyToAddress(*publicKeyECDSA)

    client, err := ethclient.Dial(cfg.RPCEndpoint)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to Ethereum client: %w", err)
    }

    chainID := big.NewInt(cfg.ChainID)
    if cfg.ChainID == 0 {
        // 自动获取链ID
        chainID, err = client.ChainID(context.Background())
        if err != nil {
            return nil, fmt.Errorf("failed to get chain ID: %w", err)
        }
    }

    return &WalletManager{
        privateKey: privateKey,
        publicKey:  publicKeyECDSA,
        address:    address,
        client:     client,
        chainID:    chainID,
        logger:     logger,
    }, nil
}

func (wm *WalletManager) GetAddress() common.Address {
    return wm.address
}

func (wm *WalletManager) GetTransactor() (*bind.TransactOpts, error) {
    transactor, err := bind.NewKeyedTransactorWithChainID(wm.privateKey, wm.chainID)
    if err != nil {
        return nil, fmt.Errorf("failed to create transactor: %w", err)
    }

    // 获取当前 nonce
    nonce, err := wm.client.PendingNonceAt(context.Background(), wm.address)
    if err != nil {
        return nil, fmt.Errorf("failed to get nonce: %w", err)
    }
    transactor.Nonce = big.NewInt(int64(nonce))

    // get gas price
    gasPrice, err := wm.client.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, fmt.Errorf("failed to get gas price: %w", err)
    }
    transactor.GasPrice = gasPrice

    transactor.Context = context.Background()

    return transactor, nil
}

func (wm *WalletManager) GetCallOpts() *bind.CallOpts {
    return &bind.CallOpts{
        Context: context.Background(),
        From:    wm.address,
    }
}

func (wm *WalletManager) GetBalance(ctx context.Context) (*big.Int, error) {
    balance, err := wm.client.BalanceAt(ctx, wm.address, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to get balance: %w", err)
    }
    return balance, nil
}

func (wm *WalletManager) Close() {
    if wm.client != nil {
        wm.client.Close()
    }
}
