package decimal

import (
    "math/big"

    "github.com/shopspring/decimal"
)

var (
    WeiPerETH = decimal.NewFromBigInt(big.NewInt(1e18), 0)
)

// 将 wei 转换为 ETH
func FromWei(wei *big.Int) decimal.Decimal {
    weiDec := decimal.NewFromBigInt(wei, 0)
    return weiDec.Div(WeiPerETH)
}

// 将 ETH 转换为 wei
func ToWei(eth decimal.Decimal) *big.Int {
    wei := eth.Mul(WeiPerETH)
    return wei.BigInt()
}

// 格式化
func FormatBalance(balance *big.Int, decimals int) decimal.Decimal {
    divisor := decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(decimals)))
    balanceDec := decimal.NewFromBigInt(balance, 0)
    return balanceDec.Div(divisor)
}

func ParseDecimal(value string) (decimal.Decimal, error) {
    return decimal.NewFromString(value)
}
