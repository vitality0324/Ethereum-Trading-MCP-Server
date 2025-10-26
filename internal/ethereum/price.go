//代币查询相关
package ethereum

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/ethereum/go-ethereum/common"
    "go.uber.org/zap"

    "github.com/your-username/ethereum-trading-mcp/pkg/decimal"
)

type PriceResponse struct {
    TokenAddress string          `json:"token_address"`
    Symbol       string          `json:"symbol"`
    PriceUSD     decimal.Decimal `json:"price_usd"`
    PriceETH     decimal.Decimal `json:"price_eth"`
    LastUpdated  time.Time       `json:"last_updated"`
    Source       string          `json:"source"`
}

type CoinGeckoResponse struct {
    Ethereum struct {
        USD float64 `json:"usd"`
    } `json:"ethereum"`
}

func (ec *EthereumClient) GetTokenPrice(ctx context.Context, tokenIdentifier string) (*PriceResponse, error) {

    var tokenAddress common.Address
    var symbol string

    // 检查是否是 ETH
    if tokenIdentifier == "ETH" || tokenIdentifier == "0x0000000000000000000000000000000000000000" {
        tokenAddress = common.HexToAddress("0x0000000000000000000000000000000000000000")
        symbol = "ETH"
    } else if common.IsHexAddress(tokenIdentifier) {
        tokenAddress = common.HexToAddress(tokenIdentifier)
        symbol = "UNKNOWN"
    } else {
        symbol = tokenIdentifier
        tokenAddress = getTokenAddressBySymbol(symbol)
    }

    priceUSD, priceETH, err := ec.fetchPriceFromCoinGecko(ctx, symbol)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch price: %w", err)
    }

    return &PriceResponse{
        TokenAddress: tokenAddress.Hex(),
        Symbol:       symbol,
        PriceUSD:     priceUSD,
        PriceETH:     priceETH,
        LastUpdated:  time.Now(),
        Source:       "coingecko",
    }, nil
}

func (ec *EthereumClient) fetchPriceFromCoinGecko(ctx context.Context, symbol string) (decimal.Decimal, decimal.Decimal, error) {
    // 将符号转换为CoinGecko 的id
    coinID := getCoinGeckoID(symbol)

    url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd,eth", coinID)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return decimal.Zero, decimal.Zero, err
    }

    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return decimal.Zero, decimal.Zero, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return decimal.Zero, decimal.Zero, err
    }

    var result map[string]map[string]float64
    if err := json.Unmarshal(body, &result); err != nil {
        return decimal.Zero, decimal.Zero, err
    }

    coinData, exists := result[coinID]
    if !exists {
        return decimal.Zero, decimal.Zero, fmt.Errorf("coin %s not found", symbol)
    }

    usdPrice, usdExists := coinData["usd"]
    ethPrice, ethExists := coinData["eth"]

    if !usdExists || !ethExists {
        return decimal.Zero, decimal.Zero, fmt.Errorf("price data incomplete for %s", symbol)
    }

    return decimal.NewFromFloat(usdPrice), decimal.NewFromFloat(ethPrice), nil
}

func getCoinGeckoID(symbol string) string {
    // 符号到 CoinGecko ID 的映射
    mapping := map[string]string{
        "ETH":  "ethereum",
        "USDC": "usd-coin",
        "USDT": "tether",
        "DAI":  "dai",
        "WBTC": "wrapped-bitcoin",
        "UNI":  "uniswap",
        "LINK": "chainlink",
        "AAVE": "aave",
    }

    if id, exists := mapping[symbol]; exists {
        return id
    }
    return symbol // 回退到使用原始符号
}

func getTokenAddressBySymbol(symbol string) common.Address {
    // 符号到地址的映射（主网）
    mapping := map[string]string{
        "USDC": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
        "USDT": "0xdAC17F958D2ee523a2206206994597C13D831ec7",
        "DAI":  "0x6B175474E89094C44Da98b954EedeAC495271d0F",
        "WBTC": "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599",
        "UNI":  "0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984",
        "LINK": "0x514910771AF9Ca656af840dff83E8264EcF986CA",
        "AAVE": "0x7Fc66500c84A76Ad7e9c93437bFc5Ac33E2DDaE9",
    }

    if addr, exists := mapping[symbol]; exists {
        return common.HexToAddress(addr)
    }
    return common.HexToAddress("0x0000000000000000000000000000000000000000")
}
