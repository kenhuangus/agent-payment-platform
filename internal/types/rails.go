package types

import (
	"fmt"
	"time"
)

// PaymentRail represents a payment rail with its characteristics
type PaymentRail string

const (
	RailACH   PaymentRail = "ach"
	RailCard  PaymentRail = "card"
	RailWire  PaymentRail = "wire"
	RailCheck PaymentRail = "check"
)

// RailCharacteristics defines the characteristics of each payment rail
type RailCharacteristics struct {
	Rail                 PaymentRail
	Name                 string
	Description          string
	MinAmount            float64
	MaxAmount            float64
	ProcessingTime       time.Duration
	SettlementTime       time.Duration
	FeeStructure         FeeStructure
	RiskLevel            string // "low", "medium", "high"
	Reversibility        bool
	InternationalSupport bool
	RequiresVerification bool
}

// FeeStructure defines the fee structure for a rail
type FeeStructure struct {
	FixedFee   float64
	PercentFee float64
	MinFee     float64
	MaxFee     float64
}

// RailSelector handles automatic rail selection
type RailSelector struct {
	Rails map[PaymentRail]*RailCharacteristics
}

// NewRailSelector creates a new rail selector with predefined rails
func NewRailSelector() *RailSelector {
	rs := &RailSelector{
		Rails: make(map[PaymentRail]*RailCharacteristics),
	}

	// Define ACH rail
	rs.Rails[RailACH] = &RailCharacteristics{
		Rail:           RailACH,
		Name:           "ACH",
		Description:    "Automated Clearing House - Low cost, batch processed",
		MinAmount:      0.01,
		MaxAmount:      100000.00,
		ProcessingTime: 1 * time.Hour,
		SettlementTime: 1 * 24 * time.Hour, // 1 business day
		FeeStructure: FeeStructure{
			FixedFee:   0.50,
			PercentFee: 0.001, // 0.1%
			MinFee:     0.50,
			MaxFee:     10.00,
		},
		RiskLevel:            "low",
		Reversibility:        true,
		InternationalSupport: false,
		RequiresVerification: false,
	}

	// Define Card rail
	rs.Rails[RailCard] = &RailCharacteristics{
		Rail:           RailCard,
		Name:           "Card Payment",
		Description:    "Credit/Debit card payment - Fast, secure",
		MinAmount:      0.50,
		MaxAmount:      10000.00,
		ProcessingTime: 5 * time.Minute,
		SettlementTime: 2 * 24 * time.Hour, // 2 business days
		FeeStructure: FeeStructure{
			FixedFee:   0.30,
			PercentFee: 0.029, // 2.9%
			MinFee:     0.30,
			MaxFee:     50.00,
		},
		RiskLevel:            "medium",
		Reversibility:        true,
		InternationalSupport: true,
		RequiresVerification: true,
	}

	// Define Wire rail
	rs.Rails[RailWire] = &RailCharacteristics{
		Rail:           RailWire,
		Name:           "Wire Transfer",
		Description:    "Real-time wire transfer - High value, immediate",
		MinAmount:      1.00,
		MaxAmount:      10000000.00,
		ProcessingTime: 30 * time.Minute,
		SettlementTime: 0, // Real-time
		FeeStructure: FeeStructure{
			FixedFee:   25.00,
			PercentFee: 0.001, // 0.1%
			MinFee:     25.00,
			MaxFee:     100.00,
		},
		RiskLevel:            "low",
		Reversibility:        false,
		InternationalSupport: true,
		RequiresVerification: true,
	}

	// Define Check rail
	rs.Rails[RailCheck] = &RailCharacteristics{
		Rail:           RailCheck,
		Name:           "Check Payment",
		Description:    "Physical check - Traditional, low tech",
		MinAmount:      1.00,
		MaxAmount:      100000.00,
		ProcessingTime: 24 * time.Hour,
		SettlementTime: 7 * 24 * time.Hour, // 1 week
		FeeStructure: FeeStructure{
			FixedFee:   1.00,
			PercentFee: 0.005, // 0.5%
			MinFee:     1.00,
			MaxFee:     25.00,
		},
		RiskLevel:            "medium",
		Reversibility:        true,
		InternationalSupport: false,
		RequiresVerification: false,
	}

	return rs
}

// SelectRail automatically selects the best rail based on payment parameters
func (rs *RailSelector) SelectRail(amount float64, counterparty string, preferences *RailPreferences) (PaymentRail, *RailCharacteristics, error) {
	var candidates []*RailCharacteristics

	// Filter rails based on amount limits
	for _, rail := range rs.Rails {
		if amount >= rail.MinAmount && amount <= rail.MaxAmount {
			candidates = append(candidates, rail)
		}
	}

	if len(candidates) == 0 {
		return "", nil, fmt.Errorf("no suitable rail found for amount %.2f", amount)
	}

	// Apply preferences and scoring
	if preferences != nil {
		return rs.selectWithPreferences(candidates, amount, preferences)
	}

	// Default selection logic
	return rs.selectOptimalRail(candidates, amount)
}

// RailPreferences defines user preferences for rail selection
type RailPreferences struct {
	Priority          string // "speed", "cost", "security"
	MaxProcessingTime time.Duration
	MaxSettlementTime time.Duration
	PreferredRails    []PaymentRail
	ExcludeRails      []PaymentRail
	International     bool
}

// selectWithPreferences selects rail based on user preferences
func (rs *RailSelector) selectWithPreferences(candidates []*RailCharacteristics, amount float64, prefs *RailPreferences) (PaymentRail, *RailCharacteristics, error) {
	// Filter by preferred/excluded rails
	var filtered []*RailCharacteristics
	for _, candidate := range candidates {
		excluded := false
		for _, excludedRail := range prefs.ExcludeRails {
			if candidate.Rail == excludedRail {
				excluded = true
				break
			}
		}
		if !excluded {
			filtered = append(filtered, candidate)
		}
	}

	if len(filtered) == 0 {
		filtered = candidates // Fallback to all candidates
	}

	// Filter by preferred rails if specified
	if len(prefs.PreferredRails) > 0 {
		var preferred []*RailCharacteristics
		for _, candidate := range filtered {
			for _, preferredRail := range prefs.PreferredRails {
				if candidate.Rail == preferredRail {
					preferred = append(preferred, candidate)
					break
				}
			}
		}
		if len(preferred) > 0 {
			filtered = preferred
		}
	}

	// Filter by processing time
	if prefs.MaxProcessingTime > 0 {
		var timeFiltered []*RailCharacteristics
		for _, candidate := range filtered {
			if candidate.ProcessingTime <= prefs.MaxProcessingTime {
				timeFiltered = append(timeFiltered, candidate)
			}
		}
		if len(timeFiltered) > 0 {
			filtered = timeFiltered
		}
	}

	// Filter by settlement time
	if prefs.MaxSettlementTime > 0 {
		var settlementFiltered []*RailCharacteristics
		for _, candidate := range filtered {
			if candidate.SettlementTime <= prefs.MaxSettlementTime {
				settlementFiltered = append(settlementFiltered, candidate)
			}
		}
		if len(settlementFiltered) > 0 {
			filtered = settlementFiltered
		}
	}

	// Apply priority-based selection
	switch prefs.Priority {
	case "speed":
		return rs.selectFastestRail(filtered)
	case "cost":
		return rs.selectCheapestRail(filtered, amount)
	case "security":
		return rs.selectMostSecureRail(filtered)
	default:
		return rs.selectOptimalRail(filtered, amount)
	}
}

// selectOptimalRail selects the best overall rail
func (rs *RailSelector) selectOptimalRail(candidates []*RailCharacteristics, amount float64) (PaymentRail, *RailCharacteristics, error) {
	if len(candidates) == 0 {
		return "", nil, fmt.Errorf("no candidates available")
	}

	// For amounts under $100, prefer ACH for cost
	if amount < 100 {
		for _, candidate := range candidates {
			if candidate.Rail == RailACH {
				return candidate.Rail, candidate, nil
			}
		}
	}

	// For amounts $100-$1000, prefer cards for speed
	if amount >= 100 && amount <= 1000 {
		for _, candidate := range candidates {
			if candidate.Rail == RailCard {
				return candidate.Rail, candidate, nil
			}
		}
	}

	// For high amounts, prefer wires
	if amount > 10000 {
		for _, candidate := range candidates {
			if candidate.Rail == RailWire {
				return candidate.Rail, candidate, nil
			}
		}
	}

	// Default to first candidate
	return candidates[0].Rail, candidates[0], nil
}

// selectFastestRail selects the rail with fastest processing
func (rs *RailSelector) selectFastestRail(candidates []*RailCharacteristics) (PaymentRail, *RailCharacteristics, error) {
	if len(candidates) == 0 {
		return "", nil, fmt.Errorf("no candidates available")
	}

	fastest := candidates[0]
	for _, candidate := range candidates[1:] {
		if candidate.ProcessingTime < fastest.ProcessingTime {
			fastest = candidate
		}
	}

	return fastest.Rail, fastest, nil
}

// selectCheapestRail selects the rail with lowest fees
func (rs *RailSelector) selectCheapestRail(candidates []*RailCharacteristics, amount float64) (PaymentRail, *RailCharacteristics, error) {
	if len(candidates) == 0 {
		return "", nil, fmt.Errorf("no candidates available")
	}

	cheapest := candidates[0]
	cheapestFee := rs.calculateFee(cheapest, amount)

	for _, candidate := range candidates[1:] {
		fee := rs.calculateFee(candidate, amount)
		if fee < cheapestFee {
			cheapest = candidate
			cheapestFee = fee
		}
	}

	return cheapest.Rail, cheapest, nil
}

// selectMostSecureRail selects the rail with lowest risk
func (rs *RailSelector) selectMostSecureRail(candidates []*RailCharacteristics) (PaymentRail, *RailCharacteristics, error) {
	if len(candidates) == 0 {
		return "", nil, fmt.Errorf("no candidates available")
	}

	// Risk level priority: low > medium > high
	riskPriority := map[string]int{
		"low":    3,
		"medium": 2,
		"high":   1,
	}

	mostSecure := candidates[0]
	highestPriority := riskPriority[mostSecure.RiskLevel]

	for _, candidate := range candidates[1:] {
		priority := riskPriority[candidate.RiskLevel]
		if priority > highestPriority {
			mostSecure = candidate
			highestPriority = priority
		}
	}

	return mostSecure.Rail, mostSecure, nil
}

// calculateFee calculates the total fee for a rail
func (rs *RailSelector) calculateFee(rail *RailCharacteristics, amount float64) float64 {
	fee := rail.FeeStructure.FixedFee + (amount * rail.FeeStructure.PercentFee)

	// Apply min/max bounds
	if fee < rail.FeeStructure.MinFee {
		fee = rail.FeeStructure.MinFee
	}
	if fee > rail.FeeStructure.MaxFee {
		fee = rail.FeeStructure.MaxFee
	}

	return fee
}

// GetRailCharacteristics returns characteristics for a specific rail
func (rs *RailSelector) GetRailCharacteristics(rail PaymentRail) (*RailCharacteristics, error) {
	characteristics, exists := rs.Rails[rail]
	if !exists {
		return nil, fmt.Errorf("rail %s not found", rail)
	}
	return characteristics, nil
}

// GetAvailableRails returns all available rails
func (rs *RailSelector) GetAvailableRails() map[PaymentRail]*RailCharacteristics {
	return rs.Rails
}

// ValidateRail checks if a rail is valid for the given amount
func (rs *RailSelector) ValidateRail(rail PaymentRail, amount float64) error {
	characteristics, exists := rs.Rails[rail]
	if !exists {
		return fmt.Errorf("rail %s not supported", rail)
	}

	if amount < characteristics.MinAmount {
		return fmt.Errorf("amount %.2f below minimum %.2f for rail %s",
			amount, characteristics.MinAmount, rail)
	}

	if amount > characteristics.MaxAmount {
		return fmt.Errorf("amount %.2f above maximum %.2f for rail %s",
			amount, characteristics.MaxAmount, rail)
	}

	return nil
}
