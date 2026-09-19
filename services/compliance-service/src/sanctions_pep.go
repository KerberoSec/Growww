package main

import (
	"strings"
	"sync"
)

type SanctionsEntry struct {
	EntityName  string
	Aliases     []string
	Source      SanctionsSource
	Country     string
	Designation string
}

type PEPEntry struct {
	Name        string
	Category    PEPStatus
	Country     string
	OfficeHeld  string
}

type SanctionsPEPEngine struct {
	mu           sync.RWMutex
	sanctions    []SanctionsEntry
	pepList      []PEPEntry
	fatfBlacklist map[string]bool // ISO3 -> true
}

func NewSanctionsPEPEngine() *SanctionsPEPEngine {
	engine := &SanctionsPEPEngine{
		fatfBlacklist: map[string]bool{
			"PRK": true, // Democratic People's Republic of Korea (North Korea)
			"IRN": true, // Islamic Republic of Iran
			"MMR": true, // Myanmar
		},
		sanctions: []SanctionsEntry{
			{EntityName: "AL-QAIDA", Source: SanctionsUN, Designation: "Terrorist Organization"},
			{EntityName: "LASHKAR-E-TAIBA", Source: SanctionsMHAUAPA, Designation: "Banned Terrorist Organization under UAPA"},
			{EntityName: "JAISH-E-MOHAMMED", Source: SanctionsMHAUAPA, Designation: "Banned Terrorist Organization under UAPA"},
			{EntityName: "HIZBUL MUJAHIDEEN", Source: SanctionsMHAUAPA, Designation: "Banned Terrorist Organization under UAPA"},
			{EntityName: "BABBAR KHALSA INTERNATIONAL", Source: SanctionsMHAUAPA, Designation: "Banned Terrorist Organization"},
			{EntityName: "DAWOOD IBRAHIM KASKAR", Source: SanctionsUN, Aliases: []string{"DAWOOD HASSAN", "SHEIKH DAWOOD"}, Designation: "Designated Global Terrorist"},
			{EntityName: "HAFIZ MUHAMMAD SAEED", Source: SanctionsUN, Aliases: []string{"HAFIZ SAEED"}, Designation: "Designated Terrorist"},
			{EntityName: "MASOOD AZHAR", Source: SanctionsUN, Aliases: []string{"MAULANA MASOOD AZHAR"}, Designation: "Designated Terrorist"},
			{EntityName: "OFAC BLOCKED ENTITY LTD", Source: SanctionsOFAC, Country: "IRN", Designation: "OFAC SDN Listed"},
			{EntityName: "EVIL CORP CYBER THREAT", Source: SanctionsOFAC, Designation: "Cyber Sanctions SDN"},
			{EntityName: "RUSSIAN ILLICIT TRANSIT BANK", Source: SanctionsOFAC, Country: "RUS", Designation: "Sectoral Sanctions"},
		},
		pepList: []PEPEntry{
			{Name: "ARUN KUMAR MINISTERIAL", Category: PEPDomestic, Country: "IND", OfficeHeld: "Member of Parliament / Union Minister"},
			{Name: "VIKRAMADITYA SINGH POLITICIAN", Category: PEPDomestic, Country: "IND", OfficeHeld: "Senior State Politician"},
			{Name: "FOREIGN DIPLOMAT JOHN DOE", Category: PEPForeign, Country: "USA", OfficeHeld: "Ambassador Extraordinary"},
			{Name: "VLADIMIR HEAD OF STATE", Category: PEPForeign, Country: "RUS", OfficeHeld: "Senior Foreign Government Official"},
			{Name: "POLITICAL SPOUSE ASSOCIATE", Category: PEPCloseAssociate, Country: "IND", OfficeHeld: "Spouse of Cabinet Minister"},
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

// IsFATFBlacklisted checks if the country is subject to FATF Call for Action (Blacklist)
func (e *SanctionsPEPEngine) IsFATFBlacklisted(countryISO3 string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	iso := strings.ToUpper(strings.TrimSpace(countryISO3))
	return e.fatfBlacklist[iso]
}

// ScreenSanctions screens an individual or entity against all active watchlists
func (e *SanctionsPEPEngine) ScreenSanctions(name string) (hit bool, source SanctionsSource, designation string) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	normName := strings.ToUpper(strings.TrimSpace(name))
	if len(normName) == 0 {
		return false, SanctionsNone, ""
	}

	for _, s := range e.sanctions {
		if normName == s.EntityName || strings.Contains(normName, s.EntityName) || strings.Contains(s.EntityName, normName) {
			return true, s.Source, s.Designation
		}
		for _, alias := range s.Aliases {
			normAlias := strings.ToUpper(alias)
			if normName == normAlias || strings.Contains(normName, normAlias) || strings.Contains(normAlias, normName) {
				return true, s.Source, s.Designation
			}
		}
	}

	return false, SanctionsNone, ""
}

// ScreenPEP checks if individual qualifies as Domestic PEP, Foreign PEP, or Close Associate
func (e *SanctionsPEPEngine) ScreenPEP(name string) (isPEP bool, pepType PEPStatus, office string) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	normName := strings.ToUpper(strings.TrimSpace(name))
	if len(normName) == 0 {
		return false, PEPNone, ""
	}

	for _, p := range e.pepList {
		if normName == p.Name || strings.Contains(normName, p.Name) || strings.Contains(p.Name, normName) {
			return true, p.Category, p.OfficeHeld
		}
	}

	return false, PEPNone, ""
}

// ComprehensiveScreening runs both Sanctions and PEP checks simultaneously
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
