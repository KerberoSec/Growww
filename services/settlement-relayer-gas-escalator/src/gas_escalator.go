package main
import ("fmt";"sync";"time")
type PendingTx struct { TxHash string; Nonce uint64; GasPrice uint64; SubmittedAt time.Time; Retries int; Status string }
type GasEscalator struct { mu sync.Mutex; pending map[uint64]*PendingTx; baseGas uint64; maxGas uint64; escalationPct float64 }
func NewGasEscalator(baseGas, maxGas uint64, escalationPct float64) *GasEscalator {
	return &GasEscalator{pending: make(map[uint64]*PendingTx), baseGas: baseGas, maxGas: maxGas, escalationPct: escalationPct}
}
func (e *GasEscalator) Submit(nonce uint64, txHash string) { e.mu.Lock(); defer e.mu.Unlock(); e.pending[nonce] = &PendingTx{TxHash: txHash, Nonce: nonce, GasPrice: e.baseGas, SubmittedAt: time.Now(), Status: "PENDING"} }
func (e *GasEscalator) EscalateStuck(staleThreshold time.Duration) []PendingTx {
	e.mu.Lock(); defer e.mu.Unlock(); var escalated []PendingTx
	for _, tx := range e.pending {
		if tx.Status == "PENDING" && time.Since(tx.SubmittedAt) > staleThreshold {
			newGas := uint64(float64(tx.GasPrice) * (1 + e.escalationPct/100))
			if newGas > e.maxGas { newGas = e.maxGas }
			tx.GasPrice = newGas; tx.Retries++; tx.SubmittedAt = time.Now()
			escalated = append(escalated, *tx)
		}
	}; return escalated
}
func (e *GasEscalator) Confirm(nonce uint64) error {
	e.mu.Lock(); defer e.mu.Unlock()
	tx, ok := e.pending[nonce]; if !ok { return fmt.Errorf("nonce not found") }
	tx.Status = "CONFIRMED"; return nil
}
