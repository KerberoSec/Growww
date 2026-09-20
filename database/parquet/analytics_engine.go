package parquet

import (
	"errors"
	"math/big"
)

// AnalyticsResult contains summary metrics for a query.
type AnalyticsResult struct {
	TotalVolumeE8   uint64  `json:"total_volume_e8"`
	TotalQuoteVolE8 uint64  `json:"total_quote_volume_e8"`
	VWAPE8          uint64  `json:"vwap_e8"`
	HighPriceE8     uint64  `json:"high_price_e8"`
	LowPriceE8      uint64  `json:"low_price_e8"`
	TradeCount      int     `json:"trade_count"`
	BuyVolumeE8     uint64  `json:"buy_volume_e8"`
	SellVolumeE8    uint64  `json:"sell_volume_e8"`
	BuyTakerRatio   float64 `json:"buy_taker_ratio"`
}

// ParquetAnalyticsEngine executes vectorized computations over columnar batches.
type ParquetAnalyticsEngine struct{}

// NewParquetAnalyticsEngine creates a new analytics engine.
func NewParquetAnalyticsEngine() *ParquetAnalyticsEngine {
	return &ParquetAnalyticsEngine{}
}

// ComputeTradeAnalytics calculates institutional metrics (OHLCV bounds, VWAP, buy/sell distribution).
func (a *ParquetAnalyticsEngine) ComputeTradeAnalytics(batch *ColumnarTradeBatch) (*AnalyticsResult, error) {
	if batch.Len() == 0 {
		return nil, errors.New("cannot compute analytics on empty batch")
	}

	res := &AnalyticsResult{
		TradeCount: batch.Len(),
		LowPriceE8: ^uint64(0), // max uint64
	}

	totQuoteBig := big.NewInt(0)
	totVolBig := big.NewInt(0)
	e8Big := big.NewInt(100_000_000)

	for i := 0; i < batch.Len(); i++ {
		p := batch.PricesE8[i]
		q := batch.QuantitiesE8[i]
		s := batch.Sides[i]

		if p > res.HighPriceE8 {
			res.HighPriceE8 = p
		}
		if p < res.LowPriceE8 {
			res.LowPriceE8 = p
		}

		res.TotalVolumeE8 += q
		if s == "BUY" {
			res.BuyVolumeE8 += q
		} else if s == "SELL" {
			res.SellVolumeE8 += q
		}

		pBig := new(big.Int).SetUint64(p)
		qBig := new(big.Int).SetUint64(q)
		prod := new(big.Int).Mul(pBig, qBig)
		quoteVol := new(big.Int).Div(prod, e8Big)

		totQuoteBig.Add(totQuoteBig, quoteVol)
		totVolBig.Add(totVolBig, qBig)
	}

	res.TotalQuoteVolE8 = totQuoteBig.Uint64()

	// Compute VWAP: (TotalQuoteVol * 1e8) / TotalVolume
	if res.TotalVolumeE8 > 0 {
		num := new(big.Int).Mul(totQuoteBig, e8Big)
		vwapBig := new(big.Int).Div(num, totVolBig)
		res.VWAPE8 = vwapBig.Uint64()

		res.BuyTakerRatio = float64(res.BuyVolumeE8) / float64(res.TotalVolumeE8)
	}

	return res, nil
}
