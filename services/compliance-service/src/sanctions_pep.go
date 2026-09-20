package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
)

// ScreeningStatus represents the outcome of sanctions & PEP screening
type ScreeningStatus string

const (
	ScreeningStatusCleared             ScreeningStatus = "CLEARED"
	ScreeningStatusClearedWhitelisted  ScreeningStatus = "CLEARED_WHITELISTED"
	ScreeningStatusPotentialMatch      ScreeningStatus = "POTENTIAL_MATCH"
	ScreeningStatusConfirmedSanctions  ScreeningStatus = "CONFIRMED_SANCTIONS"
)

// PEPTier represents seniority of politically exposed persons
type PEPTier int

const (
	PEPTierNone PEPTier = 0
	PEPTier1    PEPTier = 1 // Heads of State, Union Ministers, High Judiciary, Armed Forces Chiefs
	PEPTier2    PEPTier = 2 // Members of Parliament, Senior Civil Servants, State Enterprise Executives
	PEPTier3    PEPTier = 3 // Municipal / Regional Political Figures, Close Associates & Family
)

// SanctionsEntry represents an individual or entity on an international or statutory watchlist
type SanctionsEntry struct {
	EntityID         string
	EntityName       string
	Aliases          []string
	Source           SanctionsSource
	Country          string // ISO-2 or ISO-3
	DOB              string // YYYY-MM-DD
	GovernmentIDHash string // SHA-256 hash of PAN, Passport, or National ID
	Designation      string
	ListReferenceID  string
}

// PEPEntry represents a registered Politically Exposed Person
type PEPEntry struct {
	EntryID     string
	Name        string
	Aliases     []string
	Category    PEPStatus
	Tier        PEPTier
	Country     string
	OfficeHeld  string
	DOB         string
}

// WhitelistException represents an analyst-approved false-positive override
type WhitelistException struct {
	InvestorUUID     string    `json:"investor_uuid"`
	MatchedEntityID  string    `json:"matched_entity_id"`
	Reason           string    `json:"reason"`
	ApprovedBy       string    `json:"approved_by"`
	ApprovedAt       time.Time `json:"approved_at"`
	ExpiresAt        time.Time `json:"expires_at"`
}

// ScreeningRequest encapsulates multi-attribute screening inputs
type ScreeningRequest struct {
	EntityID            string `json:"entity_id"`
	FullName            string `json:"full_name"`
	DateOfBirth         string `json:"date_of_birth"` // YYYY-MM-DD
	NationalityISO3     string `json:"nationality_iso3"`
	ResidentCountryISO3 string `json:"resident_country_iso3"`
	GovernmentIDHash    string `json:"government_id_hash"` // Hash of PAN/Passport
}

// ScreeningMatchDetail records granular match metadata
type ScreeningMatchDetail struct {
	MatchID             string          `json:"match_id"`
	WatchlistSource     SanctionsSource `json:"watchlist_source"`
	MatchedName         string          `json:"matched_name"`
	SimilarityScore     float64         `json:"similarity_score"` // 0.0 to 1.0
	NameScore           float64         `json:"name_score"`
	DOBScore            float64         `json:"dob_score"`
	CountryScore        float64         `json:"country_score"`
	IDScore             float64         `json:"id_score"`
	Designation         string          `json:"designation"`
	IsPEP               bool            `json:"is_pep"`
	PEPTier             PEPTier         `json:"pep_tier,omitempty"`
	AutoFreezeTriggered bool            `json:"auto_freeze_triggered"`
}

// ScreeningResult represents the comprehensive screening verdict
type ScreeningResult struct {
	RequestID          string                  `json:"request_id"`
	EntityID           string                  `json:"entity_id"`
	Status             ScreeningStatus         `json:"status"`
	HighestMatchScore  float64                 `json:"highest_match_score"`
	AutoFreezeRequired bool                    `json:"auto_freeze_required"`
	EDDRequired        bool                    `json:"edd_required"`
	Matches            []ScreeningMatchDetail  `json:"matches"`
	ScreenedAt         time.Time               `json:"screened_at"`
}

// BatchScreeningSummary encapsulates metrics from batch re-screening runs
type BatchScreeningSummary struct {
	TotalScreened      int           `json:"total_screened"`
	TotalCleared       int           `json:"total_cleared"`
	TotalPotential     int           `json:"total_potential_matches"`
	TotalConfirmed     int           `json:"total_confirmed_sanctions"`
	TotalAutoFrozen    int           `json:"total_auto_frozen"`
	ProcessedDuration  time.Duration `json:"processed_duration"`
}

// SanctionsPEPEngine coordinates multi-watchlist fuzzy & phonetic screening
type SanctionsPEPEngine struct {
	mu            sync.RWMutex
	sanctions     []SanctionsEntry
	pepList       []PEPEntry
	fatfBlacklist map[string]bool                 // ISO3 -> true
	whitelist     map[string]*WhitelistException  // investorUUID -> exception
}

// NewSanctionsPEPEngine instantiates the screening engine with statutory datasets
func NewSanctionsPEPEngine() *SanctionsPEPEngine {
	engine := &SanctionsPEPEngine{
		fatfBlacklist: map[string]bool{
			"PRK": true, // Democratic People's Republic of Korea (North Korea)
			"IRN": true, // Islamic Republic of Iran
			"MMR": true, // Myanmar
		},
		whitelist: make(map[string]*WhitelistException),
		sanctions: []SanctionsEntry{
			{EntityID: "UN-001", EntityName: "AL-QAIDA", Source: SanctionsUN, Designation: "Terrorist Organization", ListReferenceID: "UNSC-1267"},
			{EntityID: "MHA-001", EntityName: "LASHKAR-E-TAIBA", Source: SanctionsMHAUAPA, Designation: "Banned Terrorist Organization under UAPA", Country: "PAK"},
			{EntityID: "MHA-002", EntityName: "JAISH-E-MOHAMMED", Source: SanctionsMHAUAPA, Designation: "Banned Terrorist Organization under UAPA", Country: "PAK"},
			{EntityID: "MHA-003", EntityName: "HIZBUL MUJAHIDEEN", Source: SanctionsMHAUAPA, Designation: "Banned Terrorist Organization under UAPA", Country: "IND"},
			{EntityID: "MHA-004", EntityName: "BABBAR KHALSA INTERNATIONAL", Source: SanctionsMHAUAPA, Designation: "Banned Terrorist Organization", Country: "IND"},
			{EntityID: "UN-002", EntityName: "DAWOOD IBRAHIM KASKAR", Aliases: []string{"DAWOOD HASSAN", "SHEIKH DAWOOD", "DAWOOD IBRAHIM"}, Source: SanctionsUN, Country: "IND", DOB: "1955-12-26", Designation: "Designated Global Terrorist", ListReferenceID: "QDi.135"},
			{EntityID: "UN-003", EntityName: "HAFIZ MUHAMMAD SAEED", Aliases: []string{"HAFIZ SAEED", "MOHAMMAD SAEED"}, Source: SanctionsUN, Country: "PAK", DOB: "1950-06-05", Designation: "Designated Terrorist", ListReferenceID: "QDi.263"},
			{EntityID: "UN-004", EntityName: "MASOOD AZHAR", Aliases: []string{"MAULANA MASOOD AZHAR", "MOHAMMAD MASOOD AZHAR"}, Source: SanctionsUN, Country: "PAK", DOB: "1968-07-10", Designation: "Designated Terrorist", ListReferenceID: "QDi.422"},
			{EntityID: "OFAC-001", EntityName: "OFAC BLOCKED ENTITY LTD", Source: SanctionsOFAC, Country: "IRN", Designation: "OFAC SDN Listed", ListReferenceID: "SDN-10928"},
			{EntityID: "OFAC-002", EntityName: "EVIL CORP CYBER THREAT", Aliases: []string{"DRIDEX GANG"}, Source: SanctionsOFAC, Designation: "Cyber Sanctions SDN", ListReferenceID: "CYBER2-29831"},
			{EntityID: "OFAC-003", EntityName: "RUSSIAN ILLICIT TRANSIT BANK", Source: SanctionsOFAC, Country: "RUS", Designation: "Sectoral Sanctions", ListReferenceID: "RUSSIA-EO14024"},
			{EntityID: "EU-001", EntityName: "BELARUS WEAPONS CONGLOMERATE", Source: SanctionsEU, Country: "BLR", Designation: "EU Restrictive Measures", ListReferenceID: "EU-CFSP-7712"},
			{EntityID: "HMT-001", EntityName: "LONDON ILLICIT TRANSIT LTD", Source: SanctionsUKHMT, Country: "GBR", Designation: "UK Sanctions Act Listed", ListReferenceID: "UK-HMT-9941"},
		},
		pepList: []PEPEntry{
			{EntryID: "PEP-001", Name: "ARUN KUMAR MINISTERIAL", Aliases: []string{"ARUN K. MINISTER"}, Category: PEPDomestic, Tier: PEPTier1, Country: "IND", OfficeHeld: "Member of Parliament / Union Minister", DOB: "1962-04-14"},
			{EntryID: "PEP-002", Name: "VIKRAMADITYA SINGH POLITICIAN", Category: PEPDomestic, Tier: PEPTier2, Country: "IND", OfficeHeld: "Senior State Politician", DOB: "1975-08-20"},
			{EntryID: "PEP-003", Name: "FOREIGN DIPLOMAT JOHN DOE", Category: PEPForeign, Tier: PEPTier1, Country: "USA", OfficeHeld: "Ambassador Extraordinary", DOB: "1958-11-03"},
			{EntryID: "PEP-004", Name: "VLADIMIR HEAD OF STATE", Category: PEPForeign, Tier: PEPTier1, Country: "RUS", OfficeHeld: "Senior Foreign Government Official", DOB: "1952-10-07"},
			{EntryID: "PEP-005", Name: "POLITICAL SPOUSE ASSOCIATE", Category: PEPCloseAssociate, Tier: PEPTier3, Country: "IND", OfficeHeld: "Spouse of Cabinet Minister", DOB: "1965-02-18"},
		},
	}
	return engine
}

// AddSanctionEntry adds custom or live updated sanction entries
func (e *SanctionsPEPEngine) AddSanctionEntry(entry SanctionsEntry) {
	e.mu.Lock()
	defer e.mu.Unlock()
	entry.EntityName = strings.ToUpper(strings.TrimSpace(entry.EntityName))
	e.sanctions = append(e.sanctions, entry)
}

// AddPEPEntry adds custom or live updated PEP entries
func (e *SanctionsPEPEngine) AddPEPEntry(entry PEPEntry) {
	e.mu.Lock()
	defer e.mu.Unlock()
	entry.Name = strings.ToUpper(strings.TrimSpace(entry.Name))
	e.pepList = append(e.pepList, entry)
}

// AddWhitelistException registers an analyst-verified false-positive exception
func (e *SanctionsPEPEngine) AddWhitelistException(ex WhitelistException) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.whitelist[ex.InvestorUUID] = &ex
}

// RemoveWhitelistException revokes a false-positive exception
func (e *SanctionsPEPEngine) RemoveWhitelistException(investorUUID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.whitelist, investorUUID)
}

// IsWhitelisted checks if an active unexpired whitelist entry exists
func (e *SanctionsPEPEngine) IsWhitelisted(investorUUID string) (*WhitelistException, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	ex, exists := e.whitelist[investorUUID]
	if !exists {
		return nil, false
	}
	if time.Now().UTC().After(ex.ExpiresAt) {
		return nil, false
	}
	return ex, true
}

// IsFATFBlacklisted checks if the country is subject to FATF Call for Action (Blacklist)
func (e *SanctionsPEPEngine) IsFATFBlacklisted(countryISO3 string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	iso := strings.ToUpper(strings.TrimSpace(countryISO3))
	return e.fatfBlacklist[iso]
}

// ScreenSanctions screens an individual or entity against all active watchlists (backwards compatible)
func (e *SanctionsPEPEngine) ScreenSanctions(name string) (hit bool, source SanctionsSource, designation string) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	normName := normalizeString(name)
	if len(normName) == 0 {
		return false, SanctionsNone, ""
	}

	for _, s := range e.sanctions {
		targetNorm := normalizeString(s.EntityName)
		score := calculateNameSimilarity(normName, targetNorm)
		if score >= 0.85 {
			return true, s.Source, s.Designation
		}
		for _, alias := range s.Aliases {
			aliasNorm := normalizeString(alias)
			if calculateNameSimilarity(normName, aliasNorm) >= 0.85 {
				return true, s.Source, s.Designation
			}
		}
	}

	return false, SanctionsNone, ""
}

// ScreenPEP checks if individual qualifies as Domestic PEP, Foreign PEP, or Close Associate (backwards compatible)
func (e *SanctionsPEPEngine) ScreenPEP(name string) (isPEP bool, pepType PEPStatus, office string) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	normName := normalizeString(name)
	if len(normName) == 0 {
		return false, PEPNone, ""
	}

	for _, p := range e.pepList {
		targetNorm := normalizeString(p.Name)
		if calculateNameSimilarity(normName, targetNorm) >= 0.85 {
			return true, p.Category, p.OfficeHeld
		}
		for _, alias := range p.Aliases {
			if calculateNameSimilarity(normName, normalizeString(alias)) >= 0.85 {
				return true, p.Category, p.OfficeHeld
			}
		}
	}

	return false, PEPNone, ""
}

// ComprehensiveScreening runs both Sanctions and PEP checks simultaneously (backwards compatible)
func (e *SanctionsPEPEngine) ComprehensiveScreening(name, countryISO3 string) (isSanctioned bool, sanctionsSource SanctionsSource, isPEP bool, pepCategory PEPStatus, eddRequired bool, accountFrozen bool) {
	// 1. Sanctions Check
	sanctioned, source, _ := e.ScreenSanctions(name)
	if sanctioned {
		return true, source, false, PEPNone, true, true // Immediate freeze on sanctions hit
	}

	// 2. Country Blacklist Check
	if e.IsFATFBlacklisted(countryISO3) {
		return true, SanctionsUN, false, PEPNone, true, true
	}

	// 3. PEP Check
	pepHit, pepCat, _ := e.ScreenPEP(name)
	if pepHit {
		return false, SanctionsNone, true, pepCat, true, false // PEP requires EDD, but not frozen
	}

	return false, SanctionsNone, false, PEPNone, false, false
}

// AdvancedScreen evaluates multi-attribute request against all watchlists with fuzzy, phonetic, and DOB matching
func (e *SanctionsPEPEngine) AdvancedScreen(req ScreeningRequest) ScreeningResult {
	e.mu.RLock()
	defer e.mu.RUnlock()

	now := time.Now().UTC()
	result := ScreeningResult{
		RequestID:  fmt.Sprintf("SCR-%d", now.UnixNano()),
		EntityID:   req.EntityID,
		Status:     ScreeningStatusCleared,
		Matches:    make([]ScreeningMatchDetail, 0),
		ScreenedAt: now,
	}

	// Check whitelist override
	if ex, ok := e.whitelist[req.EntityID]; ok && now.Before(ex.ExpiresAt) {
		result.Status = ScreeningStatusClearedWhitelisted
		return result
	}

	// Country FATF check
	country := strings.ToUpper(strings.TrimSpace(req.NationalityISO3))
	if country == "" {
		country = strings.ToUpper(strings.TrimSpace(req.ResidentCountryISO3))
	}
	if e.fatfBlacklist[country] {
		result.Status = ScreeningStatusConfirmedSanctions
		result.HighestMatchScore = 1.00
		result.AutoFreezeRequired = true
		result.EDDRequired = true
		result.Matches = append(result.Matches, ScreeningMatchDetail{
			MatchID:             fmt.Sprintf("FATF-%s", country),
			WatchlistSource:     SanctionsUN,
			MatchedName:         fmt.Sprintf("FATF Blacklisted Jurisdiction (%s)", country),
			SimilarityScore:     1.00,
			Designation:         "FATF High-Risk Call for Action Jurisdiction",
			AutoFreezeTriggered: true,
		})
		return result
	}

	normInputName := normalizeString(req.FullName)
	if normInputName == "" {
		return result
	}

	// 1. Screen Sanctions
	for _, s := range e.sanctions {
		match := e.evaluateSanctionMatch(normInputName, req, s)
		if match.SimilarityScore >= 0.70 {
			result.Matches = append(result.Matches, match)
			if match.SimilarityScore > result.HighestMatchScore {
				result.HighestMatchScore = match.SimilarityScore
			}
		}
	}

	// 2. Screen PEPs
	for _, p := range e.pepList {
		match := e.evaluatePEPMatch(normInputName, req, p)
		if match.SimilarityScore >= 0.70 {
			result.Matches = append(result.Matches, match)
			if match.SimilarityScore > result.HighestMatchScore {
				result.HighestMatchScore = match.SimilarityScore
			}
		}
	}

	// Sort matches by similarity score descending
	sort.Slice(result.Matches, func(i, j int) bool {
		return result.Matches[i].SimilarityScore > result.Matches[j].SimilarityScore
	})

	// Categorize Result Status
	if result.HighestMatchScore >= 0.95 {
		// Confirmed hit if top match is a sanction
		hasSanction := false
		for _, m := range result.Matches {
			if !m.IsPEP && m.SimilarityScore >= 0.95 {
				hasSanction = true
				break
			}
		}
		if hasSanction {
			result.Status = ScreeningStatusConfirmedSanctions
			result.AutoFreezeRequired = true
			result.EDDRequired = true
		} else {
			result.Status = ScreeningStatusPotentialMatch
			result.EDDRequired = true
		}
	} else if result.HighestMatchScore >= 0.85 {
		result.Status = ScreeningStatusPotentialMatch
		result.EDDRequired = true
	}

	return result
}

func (e *SanctionsPEPEngine) evaluateSanctionMatch(normInputName string, req ScreeningRequest, s SanctionsEntry) ScreeningMatchDetail {
	targetName := normalizeString(s.EntityName)
	bestNameScore := calculateNameSimilarity(normInputName, targetName)

	for _, alias := range s.Aliases {
		aScore := calculateNameSimilarity(normInputName, normalizeString(alias))
		if aScore > bestNameScore {
			bestNameScore = aScore
		}
	}

	dobScore := 0.0
	hasDOBCheck := req.DateOfBirth != "" && s.DOB != ""
	if hasDOBCheck {
		if req.DateOfBirth == s.DOB {
			dobScore = 1.0
		} else if len(req.DateOfBirth) >= 4 && len(s.DOB) >= 4 && req.DateOfBirth[:4] == s.DOB[:4] {
			dobScore = 0.5
		}
	}

	countryScore := 0.0
	hasCountryCheck := s.Country != "" && (req.NationalityISO3 != "" || req.ResidentCountryISO3 != "")
	if hasCountryCheck {
		if strings.EqualFold(req.NationalityISO3, s.Country) || strings.EqualFold(req.ResidentCountryISO3, s.Country) {
			countryScore = 1.0
		}
	}

	idScore := 0.0
	hasIDCheck := req.GovernmentIDHash != "" && s.GovernmentIDHash != ""
	if hasIDCheck {
		if strings.EqualFold(req.GovernmentIDHash, s.GovernmentIDHash) {
			idScore = 1.0
		}
	}

	// Calculate weighted composite score
	weightSum := 0.60
	scoreSum := bestNameScore * 0.60

	if hasDOBCheck {
		weightSum += 0.20
		scoreSum += dobScore * 0.20
	}
	if hasCountryCheck {
		weightSum += 0.10
		scoreSum += countryScore * 0.10
	}
	if hasIDCheck {
		weightSum += 0.10
		scoreSum += idScore * 0.10
	}

	composite := scoreSum / weightSum

	return ScreeningMatchDetail{
		MatchID:             s.EntityID,
		WatchlistSource:     s.Source,
		MatchedName:         s.EntityName,
		SimilarityScore:     math.Round(composite*1000) / 1000,
		NameScore:           bestNameScore,
		DOBScore:            dobScore,
		CountryScore:        countryScore,
		IDScore:             idScore,
		Designation:         s.Designation,
		IsPEP:               false,
		AutoFreezeTriggered: composite >= 0.95,
	}
}

func (e *SanctionsPEPEngine) evaluatePEPMatch(normInputName string, req ScreeningRequest, p PEPEntry) ScreeningMatchDetail {
	targetName := normalizeString(p.Name)
	bestNameScore := calculateNameSimilarity(normInputName, targetName)

	for _, alias := range p.Aliases {
		aScore := calculateNameSimilarity(normInputName, normalizeString(alias))
		if aScore > bestNameScore {
			bestNameScore = aScore
		}
	}

	dobScore := 0.0
	hasDOBCheck := req.DateOfBirth != "" && p.DOB != ""
	if hasDOBCheck {
		if req.DateOfBirth == p.DOB {
			dobScore = 1.0
		}
	}

	countryScore := 0.0
	hasCountryCheck := p.Country != "" && (req.NationalityISO3 != "" || req.ResidentCountryISO3 != "")
	if hasCountryCheck {
		if strings.EqualFold(req.NationalityISO3, p.Country) || strings.EqualFold(req.ResidentCountryISO3, p.Country) {
			countryScore = 1.0
		}
	}

	weightSum := 0.70
	scoreSum := bestNameScore * 0.70

	if hasDOBCheck {
		weightSum += 0.20
		scoreSum += dobScore * 0.20
	}
	if hasCountryCheck {
		weightSum += 0.10
		scoreSum += countryScore * 0.10
	}

	composite := scoreSum / weightSum

	return ScreeningMatchDetail{
		MatchID:             p.EntryID,
		WatchlistSource:     SanctionsNone,
		MatchedName:         p.Name,
		SimilarityScore:     math.Round(composite*1000) / 1000,
		NameScore:           bestNameScore,
		DOBScore:            dobScore,
		CountryScore:        countryScore,
		Designation:         p.OfficeHeld,
		IsPEP:               true,
		PEPTier:             p.Tier,
		AutoFreezeTriggered: false, // PEP does not auto-freeze; requires EDD
	}
}

// BatchScreenInvestors runs high-throughput screening over an active investor roster
func (e *SanctionsPEPEngine) BatchScreenInvestors(requests []ScreeningRequest) (BatchScreeningSummary, []ScreeningResult) {
	start := time.Now()
	summary := BatchScreeningSummary{
		TotalScreened: len(requests),
	}
	results := make([]ScreeningResult, len(requests))

	for i, req := range requests {
		res := e.AdvancedScreen(req)
		results[i] = res

		switch res.Status {
		case ScreeningStatusCleared, ScreeningStatusClearedWhitelisted:
			summary.TotalCleared++
		case ScreeningStatusPotentialMatch:
			summary.TotalPotential++
		case ScreeningStatusConfirmedSanctions:
			summary.TotalConfirmed++
		}

		if res.AutoFreezeRequired {
			summary.TotalAutoFrozen++
		}
	}

	summary.ProcessedDuration = time.Since(start)
	return summary, results
}

// ─────────────────────────────────────────────────────────────
// Fuzzy & Phonetic Matching Utilities
// ─────────────────────────────────────────────────────────────

func normalizeString(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	var sb strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			sb.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(sb.String()), " ")
}

func calculateNameSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}
	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	// 1. Direct containment
	if strings.Contains(s1, s2) || strings.Contains(s2, s1) {
		minLen := math.Min(float64(len(s1)), float64(len(s2)))
		maxLen := math.Max(float64(len(s1)), float64(len(s2)))
		if minLen/maxLen >= 0.80 {
			return 0.95
		}
		return 0.88
	}

	// 2. Token overlap & token sort
	tokenScore := tokenSimilarity(s1, s2)
	if tokenScore >= 0.90 {
		return tokenScore
	}

	// 3. Jaro-Winkler distance
	jwScore := jaroWinkler(s1, s2)

	// 4. Phonetic Soundex match
	if soundex(s1) == soundex(s2) {
		jwScore = math.Min(1.0, jwScore+0.15)
	}

	return math.Max(tokenScore, jwScore)
}

func tokenSimilarity(s1, s2 string) float64 {
	t1 := strings.Fields(s1)
	t2 := strings.Fields(s2)
	if len(t1) == 0 || len(t2) == 0 {
		return 0.0
	}

	// Count matching tokens
	set2 := make(map[string]bool)
	for _, tok := range t2 {
		set2[tok] = true
	}

	matchCount := 0
	for _, tok := range t1 {
		if set2[tok] {
			matchCount++
		} else {
			// Check fuzzy token match or soundex
			for tok2 := range set2 {
				if soundex(tok) == soundex(tok2) || jaroWinkler(tok, tok2) >= 0.88 {
					matchCount++
					break
				}
			}
		}
	}

	// If all tokens of the shorter query match tokens in the candidate
	minTokens := math.Min(float64(len(t1)), float64(len(t2)))
	maxTokens := math.Max(float64(len(t1)), float64(len(t2)))
	if float64(matchCount) >= minTokens && minTokens >= 2 {
		// High overlap (e.g. "Kaskar Dawood" vs "Dawood Ibrahim Kaskar")
		return 0.85 + 0.15*(minTokens/maxTokens)
	}

	sort.Strings(t1)
	sort.Strings(t2)
	sorted1 := strings.Join(t1, " ")
	sorted2 := strings.Join(t2, " ")
	if sorted1 == sorted2 {
		return 1.0
	}
	return jaroWinkler(sorted1, sorted2)
}

func jaroWinkler(s1, s2 string) float64 {
	jaro := jaroDistance(s1, s2)
	if jaro < 0.7 {
		return jaro
	}

	// Prefix length up to 4 characters
	prefixLen := 0
	maxPrefix := int(math.Min(4, math.Min(float64(len(s1)), float64(len(s2)))))
	for i := 0; i < maxPrefix; i++ {
		if s1[i] == s2[i] {
			prefixLen++
		} else {
			break
		}
	}

	return jaro + float64(prefixLen)*0.1*(1.0-jaro)
}

func jaroDistance(s1, s2 string) float64 {
	len1 := len(s1)
	len2 := len(s2)
	if len1 == 0 && len2 == 0 {
		return 1.0
	}
	if len1 == 0 || len2 == 0 {
		return 0.0
	}

	matchDistance := int(math.Max(float64(len1), float64(len2))/2) - 1
	if matchDistance < 0 {
		matchDistance = 0
	}

	s1Matches := make([]bool, len1)
	s2Matches := make([]bool, len2)
	matches := 0

	for i := 0; i < len1; i++ {
		start := int(math.Max(0, float64(i-matchDistance)))
		end := int(math.Min(float64(i+matchDistance+1), float64(len2)))

		for j := start; j < end; j++ {
			if s2Matches[j] || s1[i] != s2[j] {
				continue
			}
			s1Matches[i] = true
			s2Matches[j] = true
			matches++
			break
		}
	}

	if matches == 0 {
		return 0.0
	}

	transpositions := 0
	k := 0
	for i := 0; i < len1; i++ {
		if !s1Matches[i] {
			continue
		}
		for !s2Matches[k] {
			k++
		}
		if s1[i] != s2[k] {
			transpositions++
		}
		k++
	}

	m := float64(matches)
	return ((m / float64(len1)) + (m / float64(len2)) + ((m - float64(transpositions/2)) / m)) / 3.0
}

func soundex(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	if len(s) == 0 {
		return "0000"
	}

	var sb strings.Builder
	first := rune(s[0])
	sb.WriteRune(first)

	var lastCode byte
	switch first {
	case 'B', 'F', 'P', 'V':
		lastCode = '1'
	case 'C', 'G', 'J', 'K', 'Q', 'S', 'X', 'Z':
		lastCode = '2'
	case 'D', 'T':
		lastCode = '3'
	case 'L':
		lastCode = '4'
	case 'M', 'N':
		lastCode = '5'
	case 'R':
		lastCode = '6'
	default:
		lastCode = '0'
	}

	for i := 1; i < len(s) && sb.Len() < 4; i++ {
		var code byte
		switch s[i] {
		case 'B', 'F', 'P', 'V':
			code = '1'
		case 'C', 'G', 'J', 'K', 'Q', 'S', 'X', 'Z':
			code = '2'
		case 'D', 'T':
			code = '3'
		case 'L':
			code = '4'
		case 'M', 'N':
			code = '5'
		case 'R':
			code = '6'
		default:
			code = '0'
		}

		if code != '0' && code != lastCode {
			sb.WriteByte(code)
		}
		lastCode = code
	}

	for sb.Len() < 4 {
		sb.WriteByte('0')
	}

	return sb.String()
}
