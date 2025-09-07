package common

import (
	"regexp"
	"strings"
)

// ValidateEmail validates email format
func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidatePhone validates phone number format (basic validation)
func ValidatePhone(phone string) bool {
	// Remove all non-digit characters
	digits := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")

	// Check if it's a valid length (10-15 digits)
	return len(digits) >= 10 && len(digits) <= 15
}

// ValidateUUID validates UUID format
func ValidateUUID(uuid string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return uuidRegex.MatchString(strings.ToLower(uuid))
}

// ValidateRequired validates that a string is not empty
func ValidateRequired(value string) bool {
	return strings.TrimSpace(value) != ""
}

// ValidateLength validates string length
func ValidateLength(value string, min, max int) bool {
	length := len(strings.TrimSpace(value))
	return length >= min && length <= max
}

// ValidateEnum validates that a value is in a list of allowed values
func ValidateEnum(value string, allowed []string) bool {
	for _, allowedValue := range allowed {
		if value == allowedValue {
			return true
		}
	}
	return false
}

// ValidatePartyType validates party type
func ValidatePartyType(partyType string) bool {
	return ValidateEnum(partyType, []string{"individual", "organization"})
}

// ValidateIdentityMode validates identity mode
func ValidateIdentityMode(identityMode string) bool {
	return ValidateEnum(identityMode, []string{"did", "oauth"})
}

// ValidateStruct performs comprehensive validation and returns validation errors
func ValidateStruct(validators map[string]func() (bool, string)) []ValidationError {
	var errors []ValidationError

	for field, validator := range validators {
		if valid, message := validator(); !valid {
			errors = append(errors, ValidationError{
				Field:   field,
				Message: message,
			})
		}
	}

	return errors
}

// ValidateParty validates party data
func ValidateParty(name, partyType string) []ValidationError {
	return ValidateStruct(map[string]func() (bool, string){
		"name": func() (bool, string) {
			if !ValidateRequired(name) {
				return false, "name is required"
			}
			if !ValidateLength(name, 1, 255) {
				return false, "name must be between 1 and 255 characters"
			}
			return true, ""
		},
		"type": func() (bool, string) {
			if !ValidateRequired(partyType) {
				return false, "type is required"
			}
			if !ValidatePartyType(partyType) {
				return false, "type must be 'individual' or 'organization'"
			}
			return true, ""
		},
	})
}

// ValidateAgent validates agent data
func ValidateAgent(displayName, ownerPartyID, identityMode string) []ValidationError {
	return ValidateStruct(map[string]func() (bool, string){
		"displayName": func() (bool, string) {
			if !ValidateRequired(displayName) {
				return false, "displayName is required"
			}
			if !ValidateLength(displayName, 1, 255) {
				return false, "displayName must be between 1 and 255 characters"
			}
			return true, ""
		},
		"ownerPartyId": func() (bool, string) {
			if !ValidateRequired(ownerPartyID) {
				return false, "ownerPartyId is required"
			}
			if !ValidateUUID(ownerPartyID) {
				return false, "ownerPartyId must be a valid UUID"
			}
			return true, ""
		},
		"identityMode": func() (bool, string) {
			if !ValidateRequired(identityMode) {
				return false, "identityMode is required"
			}
			if !ValidateIdentityMode(identityMode) {
				return false, "identityMode must be 'did' or 'oauth'"
			}
			return true, ""
		},
	})
}
