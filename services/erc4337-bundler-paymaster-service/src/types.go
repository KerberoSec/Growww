package main

import (
	"math/big"
	"time"
)

// UserOperation represents an ERC-4337 UserOperation struct
type UserOperation struct {
	Sender               string   `json:"sender"`
	Nonce                *big.Int `json:"nonce"`
	InitCode             string   `json:"initCode"`
	CallData             string   `json:"callData"`
	CallGasLimit         *big.Int `json:"callGasLimit"`
	VerificationGasLimit *big.Int `json:"verificationGasLimit"`
	PreVerificationGas   *big.Int `json:"preVerificationGas"`
	MaxFeePerGas         *big.Int `json:"maxFeePerGas"`
	MaxPriorityFeePerGas *big.Int `json:"maxPriorityFeePerGas"`
	PaymasterAndData     string   `json:"paymasterAndData"`
	Signature            string   `json:"signature"`
}

// UserOpReceipt encapsulates on-chain confirmation details
type UserOpReceipt struct {
	UserOpHash    string   `json:"userOpHash"`
	Sender        string   `json:"sender"`
	Nonce         *big.Int `json:"nonce"`
	Paymaster     string   `json:"paymaster,omitempty"`
	ActualGasCost *big.Int `json:"actualGasCost"`
	ActualGasUsed *big.Int `json:"actualGasUsed"`
	Success       bool     `json:"success"`
	Reason        string   `json:"reason,omitempty"`
	TxHash        string   `json:"transactionHash"`
	BlockNumber   uint64   `json:"blockNumber"`
}

// ExecutionBundle encapsulates a batch of UserOperations executed in EntryPoint.handleOps
type ExecutionBundle struct {
	BundleID       string           `json:"bundleId"`
	RelayerAddress string           `json:"relayerAddress"`
	UserOps        []*UserOperation `json:"userOps"`
	TxHash         string           `json:"transactionHash"`
	BlockNumber    uint64           `json:"blockNumber"`
	OpCount        int              `json:"opCount"`
	TotalGasUsed   *big.Int         `json:"totalGasUsed"`
	Status         string           `json:"status"` // SUBMITTED, MINED, REVERTED
	SubmittedAt    time.Time        `json:"submittedAt"`
	MinedAt        *time.Time       `json:"minedAt,omitempty"`
}

// AccountQuota tracks investor gas sponsorship allowances and KYC status
type AccountQuota struct {
	SenderAddress          string `json:"senderAddress"`
	UserID                 string `json:"userId"`
	KYCStatus              string `json:"kycStatus"` // VERIFIED, PENDING, REJECTED
	DailyGasQuotaLimit     uint64 `json:"dailyGasQuotaLimit"`
	DailyGasQuotaConsumed  uint64 `json:"dailyGasQuotaConsumed"`
	IsSponsoredEligible    bool   `json:"isSponsoredEligible"`
	LastQuotaResetDate     string `json:"lastQuotaResetDate"`
}

// EstimateGasResult returns ERC-4337 gas estimation fields
type EstimateGasResult struct {
	PreVerificationGas   *big.Int `json:"preVerificationGas"`
	VerificationGasLimit *big.Int `json:"verificationGasLimit"`
	CallGasLimit         *big.Int `json:"callGasLimit"`
	ValidUntil           uint64   `json:"validUntil,omitempty"`
	ValidAfter           uint64   `json:"validAfter,omitempty"`
}

// JSON-RPC 2.0 structures
type JSONRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
