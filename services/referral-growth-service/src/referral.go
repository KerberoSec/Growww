package main
import ("crypto/rand";"encoding/hex";"fmt";"sync")
type ReferralNode struct { UserID string; Code string; ReferrerID string; DirectRefs []string; TotalEarnings float64 }
type ReferralService struct { mu sync.Mutex; users map[string]*ReferralNode; codes map[string]string }
func NewReferralService() *ReferralService { return &ReferralService{users: make(map[string]*ReferralNode), codes: make(map[string]string)} }
func (s *ReferralService) GenerateCode(userID string) (string, error) {
	s.mu.Lock(); defer s.mu.Unlock()
	if _, exists := s.users[userID]; exists { return s.users[userID].Code, nil }
	b := make([]byte, 4); rand.Read(b); code := "GRW" + hex.EncodeToString(b)
	node := &ReferralNode{UserID: userID, Code: code}
	s.users[userID] = node; s.codes[code] = userID; return code, nil
}
func (s *ReferralService) ApplyReferral(newUserID, code string) error {
	s.mu.Lock(); defer s.mu.Unlock()
	referrerID, ok := s.codes[code]; if !ok { return fmt.Errorf("invalid code") }
	if referrerID == newUserID { return fmt.Errorf("self-referral not allowed") }
	s.users[referrerID].DirectRefs = append(s.users[referrerID].DirectRefs, newUserID)
	s.users[newUserID] = &ReferralNode{UserID: newUserID, ReferrerID: referrerID}; return nil
}
func (s *ReferralService) CreditCommission(referrerID string, amount float64) {
	s.mu.Lock(); defer s.mu.Unlock()
	if node, ok := s.users[referrerID]; ok { node.TotalEarnings += amount }
}
