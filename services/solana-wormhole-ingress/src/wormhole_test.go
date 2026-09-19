package main
import "testing"
func TestVAAVerification(t *testing.T) {
	w := NewWormholeIngress([]string{"g1","g2","g3","g4","g5"})
	vaa := &VAA{Sequence: 1, EmitterChain: 1, Payload: []byte("transfer"), Signatures: []string{"g1","g2","g3","g4"}}
	if err := w.VerifyAndStore(vaa); err != nil { t.Fatal(err) }
	if !vaa.Verified { t.Error("expected verified") }
}
func TestInsufficientSigs(t *testing.T) {
	w := NewWormholeIngress([]string{"g1","g2","g3","g4","g5"})
	vaa := &VAA{Sequence: 2, Signatures: []string{"g1"}}
	if err := w.VerifyAndStore(vaa); err == nil { t.Error("expected quorum error") }
}
func TestDuplicateVAA(t *testing.T) {
	w := NewWormholeIngress([]string{"g1","g2","g3"})
	vaa := &VAA{Sequence: 3, Signatures: []string{"g1","g2","g3"}}
	w.VerifyAndStore(vaa)
	if err := w.VerifyAndStore(&VAA{Sequence: 3, Signatures: []string{"g1","g2","g3"}}); err == nil { t.Error("expected duplicate error") }
}
