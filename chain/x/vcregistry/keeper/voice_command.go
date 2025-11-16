package keeper

import (
	"fmt"
	"strings"

	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
)

// ============================
// VOICE COMMAND PARSING
// ============================

// ParseVoiceCommand parses a voice command into attribute types
// Supported commands:
//   - "AURA show my age" -> [ATTRIBUTE_TYPE_AGE]
//   - "AURA show my age and address" -> [ATTRIBUTE_TYPE_AGE, ATTRIBUTE_TYPE_ADDRESS_FULL]
//   - "AURA show name address height" -> [ATTRIBUTE_TYPE_FULL_NAME, ATTRIBUTE_TYPE_ADDRESS_FULL, ATTRIBUTE_TYPE_HEIGHT]
//   - "AURA show everything" -> [ALL_ATTRIBUTES]
//   - "AURA show only that I'm over 21" -> [ATTRIBUTE_TYPE_AGE] with ZK proof flag
func (k *Keeper) ParseVoiceCommand(commandText string) ([]vcregistrypb.AttributeType, bool, error) {
	if commandText == "" {
		return nil, false, fmt.Errorf("command text cannot be empty")
	}

	// Normalize command text
	commandText = strings.ToLower(strings.TrimSpace(commandText))

	// Check if command starts with "aura"
	if !strings.HasPrefix(commandText, "aura") {
		return nil, false, fmt.Errorf("command must start with 'AURA'")
	}

	// Remove "aura" prefix and common words
	commandText = strings.TrimPrefix(commandText, "aura")
	commandText = strings.TrimSpace(commandText)
	commandText = strings.TrimPrefix(commandText, "show")
	commandText = strings.TrimSpace(commandText)
	commandText = strings.TrimPrefix(commandText, "my")
	commandText = strings.TrimSpace(commandText)

	// Check for special commands
	if strings.Contains(commandText, "everything") || strings.Contains(commandText, "all") {
		return k.getAllAttributeTypes(), false, nil
	}

	// Check for ZK proof keywords
	useZKProof := strings.Contains(commandText, "only") ||
		strings.Contains(commandText, "prove") ||
		strings.Contains(commandText, "without revealing")

	// Parse individual attributes
	attributes := []vcregistrypb.AttributeType{}

	// Split by common separators
	commandText = strings.ReplaceAll(commandText, " and ", ",")
	commandText = strings.ReplaceAll(commandText, "&", ",")
	tokens := strings.Split(commandText, ",")

	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		attrType, ok := k.parseAttributeToken(token)
		if ok {
			attributes = append(attributes, attrType)
		}
	}

	if len(attributes) == 0 {
		return nil, false, fmt.Errorf("no valid attributes found in command: %s", commandText)
	}

	return attributes, useZKProof, nil
}

// parseAttributeToken parses a single token into an attribute type
func (k *Keeper) parseAttributeToken(token string) (vcregistrypb.AttributeType, bool) {
	token = strings.ToLower(strings.TrimSpace(token))

	// Remove common words
	token = strings.TrimPrefix(token, "my ")
	token = strings.TrimPrefix(token, "the ")
	token = strings.TrimPrefix(token, "i'm ")
	token = strings.TrimSpace(token)

	// Match against known attribute types
	attributeMap := map[string]vcregistrypb.AttributeType{
		// Names
		"name":       vcregistrypb.AttributeType_ATTRIBUTE_TYPE_FULL_NAME,
		"full name":  vcregistrypb.AttributeType_ATTRIBUTE_TYPE_FULL_NAME,
		"first name": vcregistrypb.AttributeType_ATTRIBUTE_TYPE_FIRST_NAME,
		"last name":  vcregistrypb.AttributeType_ATTRIBUTE_TYPE_LAST_NAME,
		"firstname":  vcregistrypb.AttributeType_ATTRIBUTE_TYPE_FIRST_NAME,
		"lastname":   vcregistrypb.AttributeType_ATTRIBUTE_TYPE_LAST_NAME,

		// Age
		"age":             vcregistrypb.AttributeType_ATTRIBUTE_TYPE_AGE,
		"birthdate":       vcregistrypb.AttributeType_ATTRIBUTE_TYPE_DATE_OF_BIRTH,
		"birthday":        vcregistrypb.AttributeType_ATTRIBUTE_TYPE_DATE_OF_BIRTH,
		"date of birth":   vcregistrypb.AttributeType_ATTRIBUTE_TYPE_DATE_OF_BIRTH,
		"over 18":         vcregistrypb.AttributeType_ATTRIBUTE_TYPE_AGE,
		"over 21":         vcregistrypb.AttributeType_ATTRIBUTE_TYPE_AGE,
		"over eighteen":   vcregistrypb.AttributeType_ATTRIBUTE_TYPE_AGE,
		"over twenty one": vcregistrypb.AttributeType_ATTRIBUTE_TYPE_AGE,

		// Address
		"address":  vcregistrypb.AttributeType_ATTRIBUTE_TYPE_ADDRESS_FULL,
		"location": vcregistrypb.AttributeType_ATTRIBUTE_TYPE_ADDRESS_FULL,
		"street":   vcregistrypb.AttributeType_ATTRIBUTE_TYPE_ADDRESS_STREET,
		"city":     vcregistrypb.AttributeType_ATTRIBUTE_TYPE_ADDRESS_CITY,
		"state":    vcregistrypb.AttributeType_ATTRIBUTE_TYPE_ADDRESS_STATE,
		"zip":      vcregistrypb.AttributeType_ATTRIBUTE_TYPE_ADDRESS_ZIP,
		"zipcode":  vcregistrypb.AttributeType_ATTRIBUTE_TYPE_ADDRESS_ZIP,
		"zip code": vcregistrypb.AttributeType_ATTRIBUTE_TYPE_ADDRESS_ZIP,
		"country":  vcregistrypb.AttributeType_ATTRIBUTE_TYPE_ADDRESS_COUNTRY,

		// Contact
		"email":        vcregistrypb.AttributeType_ATTRIBUTE_TYPE_EMAIL,
		"phone":        vcregistrypb.AttributeType_ATTRIBUTE_TYPE_PHONE,
		"phone number": vcregistrypb.AttributeType_ATTRIBUTE_TYPE_PHONE,

		// Physical
		"height":     vcregistrypb.AttributeType_ATTRIBUTE_TYPE_HEIGHT,
		"weight":     vcregistrypb.AttributeType_ATTRIBUTE_TYPE_WEIGHT,
		"eye color":  vcregistrypb.AttributeType_ATTRIBUTE_TYPE_EYE_COLOR,
		"hair color": vcregistrypb.AttributeType_ATTRIBUTE_TYPE_HAIR_COLOR,
		"eyes":       vcregistrypb.AttributeType_ATTRIBUTE_TYPE_EYE_COLOR,
		"hair":       vcregistrypb.AttributeType_ATTRIBUTE_TYPE_HAIR_COLOR,

		// Professional
		"occupation":           vcregistrypb.AttributeType_ATTRIBUTE_TYPE_OCCUPATION,
		"job":                  vcregistrypb.AttributeType_ATTRIBUTE_TYPE_OCCUPATION,
		"employer":             vcregistrypb.AttributeType_ATTRIBUTE_TYPE_EMPLOYER,
		"license":              vcregistrypb.AttributeType_ATTRIBUTE_TYPE_PROFESSIONAL_LICENSE,
		"professional license": vcregistrypb.AttributeType_ATTRIBUTE_TYPE_PROFESSIONAL_LICENSE,
		"education":            vcregistrypb.AttributeType_ATTRIBUTE_TYPE_EDUCATION_LEVEL,
		"degree":               vcregistrypb.AttributeType_ATTRIBUTE_TYPE_DEGREE,

		// Government IDs
		"passport":         vcregistrypb.AttributeType_ATTRIBUTE_TYPE_PASSPORT_NUMBER,
		"passport number":  vcregistrypb.AttributeType_ATTRIBUTE_TYPE_PASSPORT_NUMBER,
		"driver license":   vcregistrypb.AttributeType_ATTRIBUTE_TYPE_DRIVERS_LICENSE,
		"drivers license":  vcregistrypb.AttributeType_ATTRIBUTE_TYPE_DRIVERS_LICENSE,
		"driver's license": vcregistrypb.AttributeType_ATTRIBUTE_TYPE_DRIVERS_LICENSE,
		"ssn":              vcregistrypb.AttributeType_ATTRIBUTE_TYPE_SSN,
		"social security":  vcregistrypb.AttributeType_ATTRIBUTE_TYPE_SSN,

		// Special certifications
		"scuba":              vcregistrypb.AttributeType_ATTRIBUTE_TYPE_SCUBA_CERTIFIED,
		"scuba certified":    vcregistrypb.AttributeType_ATTRIBUTE_TYPE_SCUBA_CERTIFIED,
		"pilot":              vcregistrypb.AttributeType_ATTRIBUTE_TYPE_PILOTS_LICENSE,
		"pilots license":     vcregistrypb.AttributeType_ATTRIBUTE_TYPE_PILOTS_LICENSE,
		"pilot's license":    vcregistrypb.AttributeType_ATTRIBUTE_TYPE_PILOTS_LICENSE,
		"security clearance": vcregistrypb.AttributeType_ATTRIBUTE_TYPE_SECURITY_CLEARANCE,
		"clearance":          vcregistrypb.AttributeType_ATTRIBUTE_TYPE_SECURITY_CLEARANCE,
	}

	// Try exact match first
	if attrType, ok := attributeMap[token]; ok {
		return attrType, true
	}

	// Try partial match
	for key, attrType := range attributeMap {
		if strings.Contains(token, key) || strings.Contains(key, token) {
			return attrType, true
		}
	}

	return vcregistrypb.AttributeType_ATTRIBUTE_TYPE_UNSPECIFIED, false
}

// getAllAttributeTypes returns all available attribute types
func (k *Keeper) getAllAttributeTypes() []vcregistrypb.AttributeType {
	return []vcregistrypb.AttributeType{
		vcregistrypb.AttributeType_ATTRIBUTE_TYPE_FULL_NAME,
		vcregistrypb.AttributeType_ATTRIBUTE_TYPE_AGE,
		vcregistrypb.AttributeType_ATTRIBUTE_TYPE_ADDRESS_FULL,
		vcregistrypb.AttributeType_ATTRIBUTE_TYPE_EMAIL,
		vcregistrypb.AttributeType_ATTRIBUTE_TYPE_PHONE,
		vcregistrypb.AttributeType_ATTRIBUTE_TYPE_HEIGHT,
		vcregistrypb.AttributeType_ATTRIBUTE_TYPE_WEIGHT,
		vcregistrypb.AttributeType_ATTRIBUTE_TYPE_EYE_COLOR,
		vcregistrypb.AttributeType_ATTRIBUTE_TYPE_OCCUPATION,
	}
}

// ValidateVoiceCommand validates a voice command without parsing
func (k *Keeper) ValidateVoiceCommand(commandText string) (bool, string) {
	if commandText == "" {
		return false, "command text cannot be empty"
	}

	commandText = strings.ToLower(strings.TrimSpace(commandText))
	if !strings.HasPrefix(commandText, "aura") {
		return false, "command must start with 'AURA'"
	}

	if !strings.Contains(commandText, "show") {
		return false, "command must contain 'show'"
	}

	return true, ""
}

// GenerateVoiceCommandSuggestions generates suggested voice commands for a user
func (k *Keeper) GenerateVoiceCommandSuggestions(holderAddress string) []string {
	// Get user's attribute VCs
	// Get user VCs for attribute suggestions
	attributeVCs := k.ListUserVCs(holderAddress, vcregistrypb.VCStatus_VC_STATUS_ACTIVE, vcregistrypb.VCType_VC_TYPE_UNSPECIFIED)

	suggestions := []string{
		"AURA show everything",
	}

	// Generate suggestions based on available attributes
	attributeNames := make(map[vcregistrypb.AttributeType]string)
	attributeNames[vcregistrypb.AttributeType_ATTRIBUTE_TYPE_FULL_NAME] = "name"
	attributeNames[vcregistrypb.AttributeType_ATTRIBUTE_TYPE_AGE] = "age"
	attributeNames[vcregistrypb.AttributeType_ATTRIBUTE_TYPE_ADDRESS_FULL] = "address"
	attributeNames[vcregistrypb.AttributeType_ATTRIBUTE_TYPE_EMAIL] = "email"
	attributeNames[vcregistrypb.AttributeType_ATTRIBUTE_TYPE_PHONE] = "phone"
	attributeNames[vcregistrypb.AttributeType_ATTRIBUTE_TYPE_HEIGHT] = "height"

	for _, vc := range attributeVCs {
		if name, ok := attributeNames[vc.AttributeType]; ok {
			suggestions = append(suggestions, fmt.Sprintf("AURA show my %s", name))
		}
	}

	// Add common combinations
	if len(attributeVCs) >= 2 {
		suggestions = append(suggestions, "AURA show my name and address")
		suggestions = append(suggestions, "AURA show my age and height")
	}

	// Add ZK proof suggestions
	suggestions = append(suggestions, "AURA show only that I'm over 21")
	suggestions = append(suggestions, "AURA show only that I'm over 18")

	return suggestions
}
