package main
import ("fmt";"sync";"time")
type DutchAuction struct { ID string; AssetID string; StartPrice float64; ReservePrice float64; PriceStep float64; StepInterval time.Duration; StartTime time.Time; Status string; WinnerID string; FinalPrice float64 }
type AuctionEngine struct { mu sync.Mutex; auctions map[string]*DutchAuction }
func NewAuctionEngine() *AuctionEngine { return &AuctionEngine{auctions: make(map[string]*DutchAuction)} }
func (e *AuctionEngine) CreateAuction(id, assetID string, startPrice, reservePrice, priceStep float64, stepInterval time.Duration) (*DutchAuction, error) {
	if startPrice <= reservePrice { return nil, fmt.Errorf("start must exceed reserve") }
	e.mu.Lock(); defer e.mu.Unlock()
	a := &DutchAuction{ID: id, AssetID: assetID, StartPrice: startPrice, ReservePrice: reservePrice, PriceStep: priceStep, StepInterval: stepInterval, StartTime: time.Now(), Status: "ACTIVE"}
	e.auctions[id] = a; return a, nil
}
func (e *AuctionEngine) CurrentPrice(id string) (float64, error) {
	e.mu.Lock(); defer e.mu.Unlock()
	a, ok := e.auctions[id]; if !ok { return 0, fmt.Errorf("not found") }
	elapsed := time.Since(a.StartTime); steps := int(elapsed / a.StepInterval)
	price := a.StartPrice - float64(steps)*a.PriceStep
	if price < a.ReservePrice { price = a.ReservePrice }; return price, nil
}
func (e *AuctionEngine) PlaceBid(auctionID, userID string) error {
	e.mu.Lock(); defer e.mu.Unlock()
	a, ok := e.auctions[auctionID]; if !ok { return fmt.Errorf("not found") }
	if a.Status != "ACTIVE" { return fmt.Errorf("auction not active") }
	price, _ := e.currentPriceLocked(a); a.WinnerID = userID; a.FinalPrice = price; a.Status = "SOLD"; return nil
}
func (e *AuctionEngine) currentPriceLocked(a *DutchAuction) (float64, error) {
	elapsed := time.Since(a.StartTime); steps := int(elapsed / a.StepInterval)
	price := a.StartPrice - float64(steps)*a.PriceStep
	if price < a.ReservePrice { price = a.ReservePrice }; return price, nil
}
