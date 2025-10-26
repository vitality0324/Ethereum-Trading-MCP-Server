package mcp

import (
    "bufio"
    "encoding/json"
    "fmt"
    "io"
    "os"
    "sync"

    "go.uber.org/zap"

    "github.com/your-username/ethereum-trading-mcp/internal/ethereum"
)

type MCPServer struct {
    handler *MCPHandler
    logger  *zap.Logger
    input   *json.Decoder
    output  *json.Encoder
    mu      sync.Mutex
}

func NewMCPServer(ethClient *ethereum.EthereumClient, logger *zap.Logger) *MCPServer {
    handler := NewMCPHandler(ethClient, logger)
    return &MCPServer{
        handler: handler,
        logger:  logger,
        input:   json.NewDecoder(os.Stdin),
        output:  json.NewEncoder(os.Stdout),
    }
}

func (s *MCPServer) Start() error {
    s.logger.Info("Starting MCP server")


    s.output.SetIndent("", "  ")

    for {
        var msg MCPMessage
        if err := s.input.Decode(&msg); err != nil {
            if err == io.EOF {
                s.logger.Info("Input stream closed")
                return nil
            }
            s.logger.Error("Failed to decode message", zap.Error(err))
            continue
        }

        go s.handleMessage(&msg)
    }
}

func (s *MCPServer) handleMessage(msg *MCPMessage) {
    s.logger.Debug("Received message",
        zap.String("method", msg.Method),
        zap.Any("id", msg.ID),
    )

    var response *MCPMessage
    var err error

    switch msg.Method {
    case "initialize":
        response = s.handleInitialize(msg)
    case "tools/list":
        response = s.handleListTools(msg)
    case "tools/call":
        response = s.handleCallTool(msg)
    case "notifications/canceled":
        // 忽略取消通知
        return
    default:
        s.logger.Warn("Unknown method", zap.String("method", msg.Method))
        response = &MCPMessage{
            JSONRPC: "2.0",
            ID:      msg.ID,
            Error: &MCPError{
                Code:    -32601,
                Message: "Method not found",
            },
        }
    }

    if err := s.sendMessage(response); err != nil {
        s.logger.Error("Failed to send response", zap.Error(err))
    }
}

func (s *MCPServer) handleInitialize(msg *MCPMessage) *MCPMessage {
    var params InitializeParams
    if err := json.Unmarshal(msg.Params.(json.RawMessage), &params); err != nil {
        return &MCPMessage{
            JSONRPC: "2.0",
            ID:      msg.ID,
            Error: &MCPError{
                Code:    -32602,
                Message: "Invalid params",
                Data:    err.Error(),
            },
        }
    }

    result := s.handler.HandleInitialize(&params)

    return &MCPMessage{
        JSONRPC: "2.0",
        ID:      msg.ID,
        Result:  result,
    }
}

func (s *MCPServer) handleListTools(msg *MCPMessage) *MCPMessage {
    tools := s.handler.HandleListTools()

    return &MCPMessage{
        JSONRPC: "2.0",
        ID:      msg.ID,
        Result: map[string]interface{}{
            "tools": tools,
        },
    }
}

func (s *MCPServer) handleCallTool(msg *MCPMessage) *MCPMessage {
    var params CallToolParams
    if err := json.Unmarshal(msg.Params.(json.RawMessage), &params); err != nil {
        return &MCPMessage{
            JSONRPC: "2.0",
            ID:      msg.ID,
            Error: &MCPError{
                Code:    -32602,
                Message: "Invalid params",
                Data:    err.Error(),
            },
        }
    }

    result, err := s.handler.HandleCallTool(&params)
    if err != nil {
        return &MCPMessage{
            JSONRPC: "2.0",
            ID:      msg.ID,
            Error: &MCPError{
                Code:    -32603,
                Message: "Internal error",
                Data:    err.Error(),
            },
        }
    }

    return &MCPMessage{
        JSONRPC: "2.0",
        ID:      msg.ID,
        Result:  result,
    }
}

func (s *MCPServer) sendMessage(msg *MCPMessage) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    s.logger.Debug("Sending message",
        zap.Any("id", msg.ID),
        zap.String("method", msg.Method),
    )

    if err := s.output.Encode(msg); err != nil {
        return fmt.Errorf("failed to encode message: %w", err)
    }

    // 刷新输出
    if w, ok := s.output.(*json.Encoder); ok {
        if f, ok := w.(interface{ Flush() error }); ok {
            return f.Flush()
        }
    }

    return nil
}

func (s *MCPServer) sendNotification(method string, params interface{}) error {
    msg := &MCPMessage{
        JSONRPC: "2.0",
        Method:  method,
        Params:  params,
    }
    return s.sendMessage(msg)
}
