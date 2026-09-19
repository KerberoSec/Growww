package main
import ("sync";"time")
type SpeedBumpConfig struct { Symbol string; DelayMs int; BypassMarketMakers bool }
type SpeedBumpGuard struct { mu sync.RWMutex; configs map[string]*SpeedBumpConfig; marketMakers map[string]bool }
func NewSpeedBumpGuard() *SpeedBumpGuard { return &SpeedBumpGuard{configs: make(map[string]*SpeedBumpConfig), marketMakers: make(map[string]bool)} }
func (g *SpeedBumpGuard) SetConfig(symbol string, delayMs int, bypass bool) { g.mu.Lock(); defer g.mu.Unlock(); g.configs[symbol] = &SpeedBumpConfig{Symbol: symbol, DelayMs: delayMs, BypassMarketMakers: bypass} }
func (g *SpeedBumpGuard) RegisterMarketMaker(userID string) { g.mu.Lock(); defer g.mu.Unlock(); g.marketMakers[userID] = true }
func (g *SpeedBumpGuard) ApplyDelay(symbol, userID string) time.Duration {
	g.mu.RLock(); defer g.mu.RUnlock()
	cfg, ok := g.configs[symbol]; if !ok { return 0 }
	if cfg.BypassMarketMakers && g.marketMakers[userID] { return 0 }
	delay := time.Duration(cfg.DelayMs) * time.Millisecond
	time.Sleep(delay); return delay
}
