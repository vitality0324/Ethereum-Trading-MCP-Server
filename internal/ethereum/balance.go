//余额查询相关
package ethereum

import (
    "context"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/accounts/abi/bind"
    "go.uber.org/zap"

    "github.com/your-username/ethereum-trading-mcp/pkg/decimal"
)

// ERC20 只包含 balanceOf
const erc20BalanceOfABI = `[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"}]`

type BalanceResponse struct {
    Address      string          `json:"address"`
    TokenAddress *string         `json:"token_address,omitempty"`
    Balance      decimal.Decimal `json:"balance"`
    Decimals     int             `json:"decimals"`
    Symbol       *string         `json:"symbol,omitempty"`
    Name         *string         `json:"name,omitempty"`
    IsETH        bool            `json:"is_eth"`
}

func (ec *EthereumClient) GetBalance(ctx context.Context, addressStr string, tokenAddressStr *string) (*BalanceResponse, error) {
    address, err := ec.ValidateAddress(addressStr)
    if err != nil {
        return nil, err
    }

    var tokenAddress common.Address
    isETH := tokenAddressStr == nil

    if !isETH {
        tokenAddress, err = ec.ValidateAddress(*tokenAddressStr)
        if err != nil {
            return nil, err
        }
    }

    if isETH {
        // 查询 ETH余额
        balance, err := ec.client.BalanceAt(ctx, address, nil)
        if err != nil {
            return nil, fmt.Errorf("failed to get ETH balance: %w", err)
        }

        return &BalanceResponse{
            Address:  address.Hex(),
            Balance:  decimal.FromWei(balance),
            Decimals: 18,
            Symbol:   stringPtr("ETH"),
            Name:     stringPtr("Ethereum"),
            IsETH:    true,
        }, nil
    } else {
        // 查询ERC20 代币余额
        balance, decimals, symbol, name, err := ec.getERC20Balance(ctx, address, tokenAddress)
        if err != nil {
            return nil, err
        }

        return &BalanceResponse{
            Address:      address.Hex(),
            TokenAddress: stringPtr(tokenAddress.Hex()),
            Balance:      decimal.FormatBalance(balance, decimals),
            Decimals:     decimals,
            Symbol:       symbol,
            Name:         name,
            IsETH:        false,
        }, nil
    }
}

func (ec *EthereumClient) getERC20Balance(ctx context.Context, address, tokenAddress common.Address) (*big.Int, int, *string, *string, error) {
    callOpts := &bind.CallOpts{
        Context: ctx,
    }

    // 合约调用
    token, err := NewERC20Caller(tokenAddress, ec.client)
    if err != nil {
        return nil, 0, nil, nil, fmt.Errorf("failed to create token caller: %w", err)
    }

    balance, err := token.BalanceOf(callOpts, address)
    if err != nil {
        return nil, 0, nil, nil, fmt.Errorf("failed to get token balance: %w", err)
    }

    decimals, err := token.Decimals(callOpts)
    if err != nil {
        decimals = 18
    }

    symbol, err := token.Symbol(callOpts)
    var symbolPtr *string
    if err == nil {
        symbolPtr = &symbol
    }

    name, err := token.Name(callOpts)
    var namePtr *string
    if err == nil {
        namePtr = &name
    }

    return balance, int(decimals), symbolPtr, namePtr, nil
}

// ERC20
type ERC20Caller struct {

}

func NewERC20Caller(address common.Address, client *ethclient.Client) (*ERC20Caller, error) {
    return &ERC20Caller{}, nil
}

func (e *ERC20Caller) BalanceOf(opts *bind.CallOpts, address common.Address) (*big.Int, error) {
    // 实现balanceOf 调用
    return big.NewInt(0), nil
}

func (e *ERC20Caller) Decimals(opts *bind.CallOpts) (uint8, error) {
    // 实现 decimals调用
    return 18, nil
}

func (e *ERC20Caller) Symbol(opts *bind.CallOpts) (string, error) {
    // 实现 symbol 调用
    return "", nil
}

func (e *ERC20Caller) Name(opts *bind.CallOpts) (string, error) {
    // 实现 name 调用
    return "", nil
}

func stringPtr(s string) *string {
    return &s
}
