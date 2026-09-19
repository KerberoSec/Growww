package main

import (
	"fmt"
	"math"
	"sync"
	"time"
)

type LiquidityPool struct {
	Symbol       string
	BaseReserve  float64
	QuoteReserve float64
	K            float64 // x * y = k invariant
	SpreadBps    int
	mu           sync.Mutex
}

type LiquidityEngine struct {
	mu    sync.RWMutex
	pools map[string]*LiquidityPool
}

func NewLiquidityEngine() *LiquidityEngine {
	return &LiquidityEngine{pools: make(map[string]*LiquidityPool)}
}

func (e *LiquidityEngine) CreatePool(symbol string, baseReserve, quoteReserve float64, spreadBps int) (*LiquidityPool, error) {
	if baseReserve <= 0 || quoteReserve <= 0 {
		return nil, fmt.Errorf("reserves must be positive")
	}
	if spreadBps < 0 || spreadBps > 1000 {
		return nil, fmt.Errorf("spread must be 0-1000 bps")
	}
	pool := &LiquidityPool{
		Symbol: symbol, BaseReserve: baseReserve, QuoteReserve: quoteReserve,
		K: baseReserve * quoteReserve, SpreadBps: spreadBps,
	}
	e.mu.Lock()
	e.pools[symbol] = pool
	e.mu.Unlock()
	return pool, nil
}

func (p *LiquidityPool) GetQuote(side string, amount float64) (float64, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if amount <= 0 {
		return 0, fmt.Errorf("amount must be positive")
	}
	switch side {
	case "BUY":
		newBase := p.BaseReserve - amount
		if newBase <= 0 {
			return 0, fmt.Errorf("insufficient base liquidity")
		}
		newQuote := p.K / newBase
		cost := newQuote - p.QuoteReserve
		spreadAdj := cost * float64(p.SpreadBps) / 10000.0
		return cost + spreadAdj, nil
	case "SELL":
		newBase := p.BaseReserve + amount
		newQuote := p.K / newBase
		proceeds := p.QuoteReserve - newQuote
		spreadAdj := proceeds * float64(p.SpreadBps) / 10000.0
		return proceeds - spreadAdj, nil
	default:
		return 0, fmt.Errorf("side must be BUY or SELL")
	}
}

func (p *LiquidityPool) MidPrice() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.QuoteReserve / p.BaseReserve
}

func (p *LiquidityPool) Depth(pctFromMid float64) float64 {
	return math.Abs(p.BaseReserve - math.Sqrt(p.K/(p.MidPrice()*(1+pctFromMid))))
}

var _ = time.Now
