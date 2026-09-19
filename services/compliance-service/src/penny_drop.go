package main

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
)

var (
	ifscRegex = regexp.MustCompile(`^[A-Z]{4}0[A-Z0-9]{6}$`)
	honorifics = []string{"MR", "MRS", "MS", "DR", "SHRI", "SMT", "SH", "KUMAR", "KUMARI"}
)

type PennyDropValidator struct {
	minimumThreshold float64 // default 0.85
}

func NewPennyDropValidator(threshold float64) *PennyDropValidator {
	if threshold <= 0.0 || threshold > 1.0 {
		threshold = 0.85
	}
	return &PennyDropValidator{
		minimumThreshold: threshold,
	}
}

// NormalizeName strips honorifics, removes non-alphabetic chars, and condenses spaces
func (v *PennyDropValidator) NormalizeName(name string) string {
	upper := strings.ToUpper(strings.TrimSpace(name))
	// Replace punctuation with space
	reg := regexp.MustCompile(`[^A-Z\s]`)
	cleaned := reg.ReplaceAllString(upper, " ")
	tokens := strings.Fields(cleaned)
	filtered := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		isHonorific := false
		for _, h := range honorifics {
			if tok == h {
				isHonorific = true
				break
			}
		}
		if !isHonorific && len(tok) > 0 {
			filtered = append(filtered, tok)
		}
	}
	return strings.Join(filtered, " ")
}

// JaroDistance computes standard Jaro similarity between two strings
func JaroDistance(s1, s2 string) float64 {
	l1 := len(s1)
	l2 := len(s2)

	if l1 == 0 && l2 == 0 {
		return 1.0
	}
	if l1 == 0 || l2 == 0 {
		return 0.0
	}

	matchDistance := int(math.Max(float64(l1), float64(l2))/2.0) - 1
	if matchDistance < 0 {
		matchDistance = 0
	}

	s1Matches := make([]bool, l1)
	s2Matches := make([]bool, l2)

	matches := 0
	for i := 0; i < l1; i++ {
		start := int(math.Max(0, float64(i-matchDistance)))
		end := int(math.Min(float64(i+matchDistance+1), float64(l2)))

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
	for i := 0; i < l1; i++ {
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
	return (m/float64(l1) + m/float64(l2) + (m-float64(transpositions)/2.0)/m) / 3.0
}

// JaroWinklerDistance computes Jaro-Winkler similarity with prefix bonus
func JaroWinklerDistance(s1, s2 string) float64 {
	jaro := JaroDistance(s1, s2)
	if jaro < 0.7 {
		return jaro
	}

	prefixLen := 0
	maxPrefix := int(math.Min(4, math.Min(float64(len(s1)), float64(len(s2)))))
	for i := 0; i < maxPrefix; i++ {
		if s1[i] == s2[i] {
			prefixLen++
		} else {
			break
		}
	}

	scalingFactor := 0.1
	return jaro + float64(prefixLen)*scalingFactor*(1.0-jaro)
}

// ValidatePennyDrop verifies IFSC format, executes name match, and determines if account is accepted
func (v *PennyDropValidator) ValidatePennyDrop(ifsc, declaredName, beneficiaryName string) (score float64, err error) {
	ifscClean := strings.TrimSpace(strings.ToUpper(ifsc))
	if !ifscRegex.MatchString(ifscClean) {
		return 0.0, fmt.Errorf("invalid bank IFSC code '%s'; must match 4 letters, 0, 6 alphanumeric chars", ifsc)
	}

	normDeclared := v.NormalizeName(declaredName)
	normBeneficiary := v.NormalizeName(beneficiaryName)

	if len(normDeclared) == 0 || len(normBeneficiary) == 0 {
		return 0.0, errors.New("names cannot be empty after normalization")
	}

	score = JaroWinklerDistance(normDeclared, normBeneficiary)

	// Also check token set overlap for reordered names (e.g. "Kumar Rahul" vs "Rahul Kumar")
	tokensD := strings.Fields(normDeclared)
	tokensB := strings.Fields(normBeneficiary)
	if len(tokensD) == len(tokensB) && len(tokensD) > 1 {
		matchedAll := true
		for _, td := range tokensD {
			found := false
			for _, tb := range tokensB {
				if td == tb {
					found = true
					break
				}
			}
			if !found {
				matchedAll = false
				break
			}
		}
		if matchedAll && score < 0.95 {
			score = 0.95 // Direct permutation of tokens
		}
	}

	if score < v.minimumThreshold {
		return score, fmt.Errorf("bank penny drop name match failed: score %.3f below mandatory threshold %.2f (declared: '%s', beneficiary: '%s')",
			score, v.minimumThreshold, declaredName, beneficiaryName)
	}

	return score, nil
}
