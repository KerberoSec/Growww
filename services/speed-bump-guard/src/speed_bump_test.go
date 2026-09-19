package main
import ("testing";"time")
func TestSpeedBumpApplied(t *testing.T) {
	g := NewSpeedBumpGuard(); g.SetConfig("BTC/USDT", 5, true)
	start := time.Now(); g.ApplyDelay("BTC/USDT", "usr_retail")
	if time.Since(start) < 4*time.Millisecond { t.Error("delay should be ~5ms for retail") }
}
func TestMarketMakerBypass(t *testing.T) {
	g := NewSpeedBumpGuard(); g.SetConfig("BTC/USDT", 50, true); g.RegisterMarketMaker("mm_1")
	start := time.Now(); g.ApplyDelay("BTC/USDT", "mm_1")
	if time.Since(start) > 5*time.Millisecond { t.Error("market maker should bypass") }
}
func TestNoConfig(t *testing.T) {
	g := NewSpeedBumpGuard(); d := g.ApplyDelay("UNKNOWN", "usr_1")
	if d != 0 { t.Error("no config should mean no delay") }
}
