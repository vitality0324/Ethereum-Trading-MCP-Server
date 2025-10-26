package mcp

// MCP哥哥类型定义
type MCPMessage struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id,omitempty"`
    Method  string      `json:"method,omitempty"`
    Params  interface{} `json:"params,omitempty"`
    Result  interface{} `json:"result,omitempty"`
    Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

type InitializeParams struct {
    ProtocolVersion string      `json:"protocolVersion"`
    Capabilities    interface{} `json:"capabilities"`
    ClientInfo      *ClientInfo `json:"clientInfo,omitempty"`
}

type ClientInfo struct {
    Name    string `json:"name"`
    Version string `json:"version"`
}

type InitializeResult struct {
    ProtocolVersion string      `json:"protocolVersion"`
    Capabilities    interface{} `json:"capabilities"`
    ServerInfo      *ServerInfo `json:"serverInfo"`
}

type ServerInfo struct {
    Name    string `json:"name"`
    Version string `json:"version"`
}

type ToolDefinition struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"inputSchema"`
}

type CallToolParams struct {
    Name      string                 `json:"name"`
    Arguments map[string]interface{} `json:"arguments"`
}

type ToolResult struct {
    Content []ToolContent `json:"content"`
    IsError bool          `json:"isError,omitempty"`
}

type ToolContent struct {
    Type string                 `json:"type"`
    Text string                 `json:"text,omitempty"`
    Data map[string]interface{} `json:"data,omitempty"`
}
