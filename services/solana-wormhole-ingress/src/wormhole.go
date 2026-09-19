package main
import ("crypto/sha256";"encoding/hex";"fmt";"sync")
type VAA struct { Sequence uint64; EmitterChain uint16; EmitterAddr string; Payload []byte; Signatures []string; Verified bool }
type WormholeIngress struct { mu sync.Mutex; vaas map[uint64]*VAA; guardianKeys map[string]bool }
func NewWormholeIngress(guardianKeys []string) *WormholeIngress {
	keys := make(map[string]bool); for _, k := range guardianKeys { keys[k] = true }
	return &WormholeIngress{vaas: make(map[uint64]*VAA), guardianKeys: keys}
}
func (w *WormholeIngress) VerifyAndStore(vaa *VAA) error {
	w.mu.Lock(); defer w.mu.Unlock()
	if _, exists := w.vaas[vaa.Sequence]; exists { return fmt.Errorf("duplicate VAA sequence %d", vaa.Sequence) }
	quorum := (len(w.guardianKeys)*2)/3 + 1; validSigs := 0
	for _, sig := range vaa.Signatures { if w.guardianKeys[sig] { validSigs++ } }
	if validSigs < quorum { return fmt.Errorf("insufficient guardian signatures: %d/%d", validSigs, quorum) }
	vaa.Verified = true; w.vaas[vaa.Sequence] = vaa; return nil
}
func (w *WormholeIngress) GetVAA(seq uint64) (*VAA, error) {
	w.mu.Lock(); defer w.mu.Unlock()
	v, ok := w.vaas[seq]; if !ok { return nil, fmt.Errorf("not found") }; return v, nil
}
func ComputeVAADigest(payload []byte) string { h := sha256.Sum256(payload); return hex.EncodeToString(h[:]) }
