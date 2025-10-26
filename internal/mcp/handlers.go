package mcp

import (
    "context"
    "encoding/json"
    "fmt"
    "strconv"

    "go.uber.org/zap"

    "github.com/your-username/ethereum-trading-mcp/internal/ethereum"
    "github.com/your-username/ethereum-trading-mcp/pkg/decimal"
)

type MCPHandler struct {
    ethClient *ethereum.EthereumClient
    logger    *zap.Logger
    tools     []ToolDefinition
}

func NewMCPHandler(ethClient *ethereum.EthereumClient, logger *zap.Logger) *MCPHandler {
    handler := &MCPHandler{
        ethClient: ethClient,
        logger:    logger,
    }
    handler.registerTools()
    return handler
}

func (h *MCPHandler) registerTools() {
    h.tools = []ToolDefinition{
        {
            Name:        "get_balance",
            Description: "Query ETH and ERC20 token balances for a wallet address",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "address": map[string]interface{}{
                        "type":        "string",
                        "description": "Ethereum wallet address",
                    },
                    "token_address": map[string]interface{}{
                        "type":        "string",
                        "description": "Optional ERC20 token contract address",
                    },
                },
                "required": []string{"address"},
            },
        },
        {
            Name:        "get_token_price",
            Description: "Get current token price in USD or ETH",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "token_identifier": map[string]interface{}{
                        "type":        "string",
                        "description": "Token address or symbol (e.g., 'ETH', 'USDC', '0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48')",
                    },
                },
                "required": []string{"token_identifier"},
            },
        },
        {
            Name:        "swap_tokens",
            Description: "Simulate a token swap on Uniswap V2 or V3",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "from_token": map[string]interface{}{
                        "type":        "string",
                        "description": "Source token address or symbol",
                    },
                    "to_token": map[string]interface{}{
                        "type":        "string",
                        "description": "Destination token address or symbol",
                    },
                    "amount": map[string]interface{}{
                        "type":        "string",
                        "description": "Amount to swap (as string to preserve precision)",
                    },
                    "slippage_tolerance": map[string]interface{}{
                        "type":        "string",
                        "description": "Slippage tolerance as decimal (e.g., '0.01' for 1%)",
                    },
                    "use_v3": map[string]interface{}{
                        "type":        "boolean",
                        "description": "Whether to use Uniswap V3 (defaults to V2)",
                    },
                },
                "required": []string{"from_token", "to_token", "amount"},
            },
        },
    }
}

func (h *MCPHandler) HandleInitialize(params *InitializeParams) *InitializeResult {
    h.logger.Info("MCP client initialized",
        zap.String("client", params.ClientInfo.Name),
        zap.String("version", params.ClientInfo.Version),
    )

    return &InitializeResult{
        ProtocolVersion: "2024-11-05",
        Capabilities: map[string]interface{}{
            "tools": map[string]interface{}{
                "listChanged": true,
            },
        },
        ServerInfo: &ServerInfo{
            Name:    "Ethereum Trading MCP Server",
            Version: "1.0.0",
        },
    }
}

func (h *MCPHandler) HandleListTools() []ToolDefinition {
    return h.tools
}

func (h *MCPHandler) HandleCallTool(params *CallToolParams) (*ToolResult, error) {
    h.logger.Debug("Tool called",
        zap.String("name", params.Name),
        zap.Any("arguments", params.Arguments),
    )

    switch params.Name {
    case "get_balance":
        return h.handleGetBalance(params.Arguments)
    case "get_token_price":
        return h.handleGetTokenPrice(params.Arguments)
    case "swap_tokens":
        return h.handleSwapTokens(params.Arguments)
    default:
        return nil, fmt.Errorf("unknown tool: %s", params.Name)
    }
}

func (h *MCPHandler) handleGetBalance(args map[string]interface{}) (*ToolResult, error) {
    address, ok := args["address"].(string)
    if !ok {
        return nil, fmt.Errorf("address is required and must be a string")
    }

    var tokenAddress *string
    if tokenAddr, ok := args["token_address"].(string); ok {
        tokenAddress = &tokenAddr
    }

    ctx := context.Background()
    balance, err := h.ethClient.GetBalance(ctx, address, tokenAddress)
    if err != nil {
        return &ToolResult{
            Content: []ToolContent{
                {
                    Type: "text",
                    Text: fmt.Sprintf("Error getting balance: %v", err),
                },
            },
            IsError: true,
        }, nil
    }

    balanceJSON, err := json.MarshalIndent(balance, "", "  ")
    if err != nil {
        return nil, fmt.Errorf("failed to marshal balance: %w", err)
    }

    return &ToolResult{
        Content: []ToolContent{
            {
                Type: "text",
                Text: fmt.Sprintf("Balance for %s:", address),
            },
            {
                Type: "text",
                Text: string(balanceJSON),
            },
        },
    }, nil
}

func (h *MCPHandler) handleGetTokenPrice(args map[string]interface{}) (*ToolResult, error) {
    tokenIdentifier, ok := args["token_identifier"].(string)
    if !ok {
        return nil, fmt.Errorf("token_identifier is required and must be a string")
    }

    ctx := context.Background()
    price, err := h.ethClient.GetTokenPrice(ctx, tokenIdentifier)
    if err != nil {
        return &ToolResult{
            Content: []ToolContent{
                {
                    Type: "text",
                    Text: fmt.Sprintf("Error getting token price: %v", err),
                },
            },
            IsError: true,
        }, nil
    }

    priceJSON, err := json.MarshalIndent(price, "", "  ")
    if err != nil {
        return nil, fmt.Errorf("failed to marshal price: %w", err)
    }

    return &ToolResult{
        Content: []ToolContent{
            {
                Type: "text",
                Text: fmt.Sprintf("Price for %s:", tokenIdentifier),
            },
            {
                Type: "text",
                Text: string(priceJSON),
            },
        },
    }, nil
}

func (h *MCPHandler) handleSwapTokens(args map[string]interface{}) (*ToolResult, error) {
    fromToken, ok := args["from_token"].(string)
    if !ok {
        return nil, fmt.Errorf("from_token is required and must be a string")
    }

    toToken, ok := args["to_token"].(string)
    if !ok {
        return nil, fmt.Errorf("to_token is required and must be a string")
    }

    amountStr, ok := args["amount"].(string)
    if !ok {
        return nil, fmt.Errorf("amount is required and must be a string")
    }

    amount, err := decimal.ParseDecimal(amountStr)
    if err != nil {
        return nil, fmt.Errorf("invalid amount format: %w", err)
    }

    slippageStr := "0.01" // 默认 1% 滑点
    if s, ok := args["slippage_tolerance"].(string); ok {
        slippageStr = s
    }

    slippage, err := decimal.ParseDecimal(slippageStr)
    if err != nil {
        return nil, fmt.Errorf("invalid slippage_tolerance format: %w", err)
    }

    useV3 := false
    if v3, ok := args["use_v3"].(bool); ok {
        useV3 = v3
    }

    req := &ethereum.SwapRequest{
        FromToken:       fromToken,
        ToToken:         toToken,
        Amount:          amount,
        SlippageTolerance: slippage,
        UseV3:           useV3,
    }

    ctx := context.Background()
    result, err := h.ethClient.SwapTokens(ctx, req)
    if err != nil {
        return &ToolResult{
            Content: []ToolContent{
                {
                    Type: "text",
                    Text: fmt.Sprintf("Error simulating swap: %v", err),
                },
            },
            IsError: true,
        }, nil
    }

    resultJSON, err := json.MarshalIndent(result, "", "  ")
    if err != nil {
        return nil, fmt.Errorf("failed to marshal swap result: %w", err)
    }

    var text string
    if result.Success {
        text = fmt.Sprintf("Swap simulation successful:\nInput: %s %s\nEstimated Output: %s %s\nGas Cost: ~$%s",
            result.InputAmount.String(), result.FromToken,
            result.EstimatedOutput.String(), result.ToToken,
            result.GasCostUSD.StringFixed(2))
    } else {
        text = fmt.Sprintf("Swap simulation failed: %s", *result.Error)
    }

    return &ToolResult{
        Content: []ToolContent{
            {
                Type: "text",
                Text: text,
            },
            {
                Type: "text",
                Text: string(resultJSON),
            },
        },
    }, nil
}
