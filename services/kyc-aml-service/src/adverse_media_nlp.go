package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

// AdverseMediaDecision represents compliance screening outcome
type AdverseMediaDecision string

const (
	AdverseDecisionPass       AdverseMediaDecision = "PASS"
	AdverseDecisionEDDRequired AdverseMediaDecision = "EDD_REQUIRED"
	AdverseDecisionProhibited  AdverseMediaDecision = "PROHIBITED_MATCH"
)

// OffenseCategory represents FATF and statutory crime categories
type OffenseCategory string

const (
	OffenseFraudEmbezzlement OffenseCategory = "FRAUD_EMBEZZLEMENT"
	OffenseBriberyCorruption OffenseCategory = "BRIBERY_CORRUPTION"
	OffenseTaxEvasion        OffenseCategory = "TAX_EVASION"
	OffenseMoneyLaundering   OffenseCategory = "MONEY_LAUNDERING"
	OffenseNarcotics         OffenseCategory = "NARCOTICS_TRAFFICKING"
	OffenseTerrorFinancing   OffenseCategory = "TERRORIST_FINANCING"
	OffenseOrganizedCrime    OffenseCategory = "ORGANIZED_CRIME"
	OffenseSanctionsEvasion  OffenseCategory = "SANCTIONS_EVASION"
)

// Severity weights per crime category (0 to 100)
var offenseSeverity = map[OffenseCategory]float64{
	OffenseTerrorFinancing:   100.0,
	OffenseSanctionsEvasion:  100.0,
	OffenseMoneyLaundering:   90.0,
	OffenseOrganizedCrime:    85.0,
	OffenseNarcotics:         80.0,
	OffenseFraudEmbezzlement: 75.0,
	OffenseBriberyCorruption: 70.0,
	OffenseTaxEvasion:        60.0,
}

// MediaSourceTier represents credibility of news / enforcement source
type MediaSourceTier int

const (
	SourceTier1RegulatoryPress MediaSourceTier = 1 // Weight: 1.0 (SEBI, RBI, ED, CBI, OFAC, Reuters, Bloomberg)
	SourceTier2NationalPress   MediaSourceTier = 2 // Weight: 0.7 (Economic Times, Times of India, Mint, WSJ)
	SourceTier3RegionalBlog    MediaSourceTier = 3 // Weight: 0.3 (Local blogs, unverified outlets)
)

var sourceTierWeights = map[MediaSourceTier]float64{
	SourceTier1RegulatoryPress: 1.0,
	SourceTier2NationalPress:   0.7,
	SourceTier3RegionalBlog:    0.3,
}

// AdverseMediaArticle represents an ingested news or regulatory enforcement article
type AdverseMediaArticle struct {
	ArticleID      string          `json:"article_id"`
	Title          string          `json:"title"`
	Source         string          `json:"source"`
	SourceTier     MediaSourceTier `json:"source_tier"`
	PublishedAt    time.Time       `json:"published_at"`
	ExtractedText  string          `json:"extracted_text"`
	MatchedEntities []string       `json:"matched_entities"`
	OffenseCategory OffenseCategory `json:"offense_category"`
	SentimentScore float64         `json:"sentiment_score"` // -1.0 (very negative) to +1.0
}

// AdverseMediaScreeningRequest represents input to the NLP screening engine
type AdverseMediaScreeningRequest struct {
	RequestID    string   `json:"request_id"`
	EntityID     string   `json:"entity_id"`
	FullName     string   `json:"full_name"`
	Aliases      []string `json:"aliases"`
	DateOfBirth  string   `json:"date_of_birth,omitempty"`
	CountryCode  string   `json:"country_code"`
	IsPEP        bool     `json:"is_pep"`
}

// AdverseMediaScreeningReport represents screening result
type AdverseMediaScreeningReport struct {
	ReportID              string               `json:"report_id"`
	RequestID             string               `json:"request_id"`
	EntityID              string               `json:"entity_id"`
	FullName              string               `json:"full_name"`
	Decision              AdverseMediaDecision `json:"decision"`
	AdverseMediaRiskScore float64              `json:"adverse_media_risk_score"` // 0.0 to 100.0
	MatchedArticlesCount  int                  `json:"matched_articles_count"`
	HighestSeverityOffense OffenseCategory     `json:"highest_severity_offense"`
	MatchedArticles       []*AdverseMediaArticle `json:"matched_articles"`
	BesuAttestationHash   string               `json:"besu_attestation_hash"`
	ScreenedAt            time.Time            `json:"screened_at"`
}

// AdverseMediaScreeningNLPEngine processes multilingual adverse news & regulatory enforcement screening
type AdverseMediaScreeningNLPEngine struct {
	mu           sync.RWMutex
	articleDB    []*AdverseMediaArticle
	decayLambda  float64 // 0.0005 per day (half-life ~ 3.8 years)
}

// NewAdverseMediaScreeningNLPEngine creates a new NLP screening engine
func NewAdverseMediaScreeningNLPEngine() *AdverseMediaScreeningNLPEngine {
	return &AdverseMediaScreeningNLPEngine{
		articleDB:   make([]*AdverseMediaArticle, 0),
		decayLambda: 0.0005,
	}
}

// IngestArticle ingests a news or regulatory article into the screening index
func (e *AdverseMediaScreeningNLPEngine) IngestArticle(art *AdverseMediaArticle) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.articleDB = append(e.articleDB, art)
}

// ComputeJaroWinklerSimilarity computes string distance (0.0 to 1.0)
func ComputeJaroWinklerSimilarity(s1, s2 string) float64 {
	s1 = strings.ToLower(strings.TrimSpace(s1))
	s2 = strings.ToLower(strings.TrimSpace(s2))

	if s1 == s2 {
		return 1.0
	}
	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	matchDistance := int(math.Max(float64(len(s1)), float64(len(s2)))/2) - 1
	if matchDistance < 0 {
		matchDistance = 0
	}

	s1Matches := make([]bool, len(s1))
	s2Matches := make([]bool, len(s2))

	matches := 0
	for i := 0; i < len(s1); i++ {
		start := int(math.Max(0, float64(i-matchDistance)))
		end := int(math.Min(float64(i+matchDistance+1), float64(len(s2))))

		for j := start; j < end; j++ {
			if !s2Matches[j] && s1[i] == s2[j] {
				s1Matches[i] = true
				s2Matches[j] = true
				matches++
				break
			}
		}
	}

	if matches == 0 {
		return 0.0
	}

	transpositions := 0
	k := 0
	for i := 0; i < len(s1); i++ {
		if s1Matches[i] {
			for !s2Matches[k] {
				k++
			}
			if s1[i] != s2[k] {
				transpositions++
			}
			k++
		}
	}

	m := float64(matches)
	jaro := (m/float64(len(s1)) + m/float64(len(s2)) + (m-float64(transpositions)/2.0)/m) / 3.0

	// Winkler prefix adjustment (up to 4 chars)
	prefixLength := 0
	maxPrefix := int(math.Min(4, math.Min(float64(len(s1)), float64(len(s2)))))
	for i := 0; i < maxPrefix; i++ {
		if s1[i] == s2[i] {
			prefixLength++
		} else {
			break
		}
	}

	p := 0.1
	return jaro + float64(prefixLength)*p*(1.0-jaro)
}

// ScreenEntity evaluates candidate against all ingested adverse media records
func (e *AdverseMediaScreeningNLPEngine) ScreenEntity(req *AdverseMediaScreeningRequest) (*AdverseMediaScreeningReport, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if req == nil {
		return nil, errors.New("request cannot be nil")
	}
	if req.FullName == "" {
		return nil, errors.New("full_name is required")
	}

	namesToMatch := append([]string{req.FullName}, req.Aliases...)
	var matchedArticles []*AdverseMediaArticle
	maxRiskScore := 0.0
	var highestOffense OffenseCategory
	now := time.Now().UTC()

	for _, art := range e.articleDB {
		isMatch := false
		bestSimilarity := 0.0

		// Entity name matching in text / entity tags
		for _, name := range namesToMatch {
			// 1. Direct tag match or substring
			if strings.Contains(strings.ToLower(art.ExtractedText), strings.ToLower(name)) ||
				strings.Contains(strings.ToLower(art.Title), strings.ToLower(name)) {
				isMatch = true
				bestSimilarity = 1.0
				break
			}
			for _, taggedEntity := range art.MatchedEntities {
				sim := ComputeJaroWinklerSimilarity(name, taggedEntity)
				if sim >= 0.88 { // High-confidence fuzzy match
					isMatch = true
					if sim > bestSimilarity {
						bestSimilarity = sim
					}
				}
			}
		}

		if isMatch {
			matchedArticles = append(matchedArticles, art)

			// Calculate impact: BaseSeverity * SourceWeight * TimeDecay * FuzzyWeight * SentimentMultiplier
			baseSev := offenseSeverity[art.OffenseCategory]
			if baseSev == 0 {
				baseSev = 50.0
			}

			srcWeight := sourceTierWeights[art.SourceTier]
			if srcWeight == 0 {
				srcWeight = 0.5
			}

			// Time decay: e^(-lambda * days)
			daysOld := now.Sub(art.PublishedAt).Hours() / 24.0
			if daysOld < 0 {
				daysOld = 0
			}
			timeDecay := math.Exp(-e.decayLambda * daysOld)

			// Sentiment multiplier: if sentiment is -1.0, multiplier is 1.0; if neutral, 0.7
			sentimentMult := 0.7
			if art.SentimentScore < 0 {
				sentimentMult = 0.7 + 0.3*math.Abs(art.SentimentScore)
			}

			articleRisk := baseSev * srcWeight * timeDecay * bestSimilarity * sentimentMult

			if req.IsPEP {
				articleRisk *= 1.25 // PEP multiplier
			}

			if articleRisk > maxRiskScore {
				maxRiskScore = articleRisk
				highestOffense = art.OffenseCategory
			}
		}
	}

	if maxRiskScore > 100.0 {
		maxRiskScore = 100.0
	}

	// Decision Logic
	var decision AdverseMediaDecision
	if maxRiskScore >= 60.0 || highestOffense == OffenseTerrorFinancing || highestOffense == OffenseSanctionsEvasion {
		decision = AdverseDecisionProhibited
	} else if maxRiskScore >= 25.0 {
		decision = AdverseDecisionEDDRequired
	} else {
		decision = AdverseDecisionPass
	}

	// Besu Attestation Hash
	reportID := fmt.Sprintf("rep-media-%s", req.RequestID)
	besuPayload := fmt.Sprintf("%s:%s:%s:%.2f:%d", req.EntityID, req.FullName, decision, maxRiskScore, time.Now().Unix())
	besuHash := "0x" + hex.EncodeToString(sha256.New().Sum([]byte(besuPayload)))

	report := &AdverseMediaScreeningReport{
		ReportID:              reportID,
		RequestID:             req.RequestID,
		EntityID:              req.EntityID,
		FullName:              req.FullName,
		Decision:              decision,
		AdverseMediaRiskScore: maxRiskScore,
		MatchedArticlesCount:  len(matchedArticles),
		HighestSeverityOffense: highestOffense,
		MatchedArticles:       matchedArticles,
		BesuAttestationHash:   besuHash,
		ScreenedAt:            time.Now().UTC(),
	}

	return report, nil
}
