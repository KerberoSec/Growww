package main
import ("testing";"time")
func TestDutchAuction(t *testing.T) {
	eng := NewAuctionEngine()
	a, err := eng.CreateAuction("a1", "RWA_BOND_001", 1000, 500, 50, 1*time.Second); if err != nil { t.Fatal(err) }
	if a.Status != "ACTIVE" { t.Error("expected active") }
	price, _ := eng.CurrentPrice("a1")
	if price > 1000 || price < 500 { t.Errorf("price out of range: %f", price) }
	eng.PlaceBid("a1", "usr_1")
	if a.Status != "SOLD" { t.Error("expected sold") }
}
func TestReserveFloor(t *testing.T) {
	eng := NewAuctionEngine()
	eng.CreateAuction("a2", "ASSET", 100, 80, 50, 1*time.Millisecond)
	time.Sleep(10 * time.Millisecond)
	price, _ := eng.CurrentPrice("a2")
	if price < 80 { t.Errorf("price %f below reserve 80", price) }
}
