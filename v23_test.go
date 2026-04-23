package iabtcfv2

import (
	"strings"
	"testing"
	"time"
)

// countSegments returns the number of dot-separated segments in a TC string.
func countSegments(tcString string) int {
	if tcString == "" {
		return 0
	}
	return len(strings.Split(tcString, "."))
}

// newTestCoreString creates a minimal CoreString for testing with the given TcfPolicyVersion.
// Uses a Created timestamp well before the v2.3 deadline to isolate policyVersion tests
// from the deadline-based validation.
func newTestCoreString(policyVersion int) *CoreString {
	return &CoreString{
		Version:                2,
		Created:                time.Date(2022, 1, 26, 0, 0, 0, 0, time.UTC),
		LastUpdated:            time.Date(2022, 1, 26, 0, 0, 0, 0, time.UTC),
		CmpId:                  92,
		CmpVersion:             1,
		ConsentScreen:          0,
		ConsentLanguage:        "EN",
		VendorListVersion:      32,
		TcfPolicyVersion:       policyVersion,
		IsServiceSpecific:      false,
		UseNonStandardTexts:    false,
		SpecialFeatureOptIns:   map[int]bool{},
		PurposesConsent:        map[int]bool{},
		PurposesLITransparency: map[int]bool{},
		PurposeOneTreatment:    false,
		PublisherCC:            "AA",
		MaxVendorId:            0,
		IsRangeEncoding:        false,
		VendorsConsent:         map[int]bool{},
		MaxVendorIdLI:          0,
		IsRangeEncodingLI:      false,
		VendorsLITransparency:  map[int]bool{},
		NumPubRestrictions:     0,
	}
}

// --- Test 1: Decode v2.3 mandatory DV ---
// Create a TC string with Version=2, TcfPolicyVersion=5, but WITHOUT DisclosedVendors segment.
// Verify Decode() returns an error about mandatory DV.
func TestDecodeV23MandatoryDisclosedVendors(t *testing.T) {
	core := newTestCoreString(TcfPolicyVersion23) // policyVersion=5
	tcString := core.Encode()

	_, err := Decode(tcString)
	if err == nil {
		t.Errorf("Decode() should return error for v2.3 TC string without DisclosedVendors segment")
		return
	}

	expected := "TCF v2.3: DisclosedVendors segment is mandatory for policyVersion >= 5"
	if err.Error() != expected {
		t.Errorf("Decode() error = %q, want %q", err.Error(), expected)
	}
}

// --- Test 2: Decode v2.2 optional DV ---
// Create a TC string with Version=2, TcfPolicyVersion=4 (pre-v2.3), without DV.
// Verify Decode() succeeds.
func TestDecodeV22OptionalDisclosedVendors(t *testing.T) {
	core := newTestCoreString(4) // policyVersion=4 (pre-v2.3)
	tcString := core.Encode()

	data, err := Decode(tcString)
	if err != nil {
		t.Errorf("Decode() should succeed for v2.2 TC string without DisclosedVendors: %s", err)
		return
	}

	if data.CoreString == nil {
		t.Errorf("Decode() should return non-nil CoreString")
	}

	if data.CoreString.TcfPolicyVersion != 4 {
		t.Errorf("CoreString.TcfPolicyVersion = %d, want 4", data.CoreString.TcfPolicyVersion)
	}
}

// --- Test 3: Encode v2.3 mandatory DV ---
// Create TCData with Version=2, TcfPolicyVersion=5, DisclosedVendors=nil.
// Verify ToTCString() auto-creates DV and produces 2+ segments.
func TestEncodeV23MandatoryDisclosedVendors(t *testing.T) {
	data := &TCData{
		CoreString:       newTestCoreString(TcfPolicyVersion23), // policyVersion=5
		DisclosedVendors: nil,                                    // nil - should be auto-created
	}

	result := data.ToTCString()
	segments := countSegments(result)

	if segments < 2 {
		t.Errorf("ToTCString() should produce at least 2 segments for v2.3 (Core + DV), got %d segments: %s", segments, result)
	}

	// Verify DisclosedVendors was auto-created on the TCData struct
	if data.DisclosedVendors == nil {
		t.Errorf("ToTCString() should auto-create DisclosedVendors for v2.3")
	}

	// Verify the second segment is DisclosedVendors
	segParts := strings.Split(result, ".")
	if len(segParts) >= 2 {
		segType, err := GetSegmentType(segParts[1])
		if err != nil {
			t.Errorf("GetSegmentType() error on second segment: %s", err)
		} else if segType != SegmentTypeDisclosedVendors {
			t.Errorf("Second segment should be DisclosedVendors (type %d), got type %d", SegmentTypeDisclosedVendors, segType)
		}
	}
}

// --- Test 4: Encode v2.2 optional DV ---
// Create TCData with Version=2, TcfPolicyVersion=4, DisclosedVendors=nil.
// Verify ToTCString() produces only 1 segment (Core only).
func TestEncodeV22OptionalDisclosedVendors(t *testing.T) {
	data := &TCData{
		CoreString:       newTestCoreString(4), // policyVersion=4 (pre-v2.3)
		DisclosedVendors: nil,                  // nil - should NOT be auto-created
	}

	result := data.ToTCString()
	segments := countSegments(result)

	if segments != 1 {
		t.Errorf("ToTCString() should produce exactly 1 segment for v2.2 without DV, got %d segments: %s", segments, result)
	}
}

// --- Test 5: Validate() v2.3 missing DV ---
// Test Validate() returns error for v2.3 without DV.
func TestValidateV23MissingDisclosedVendors(t *testing.T) {
	data := &TCData{
		CoreString:       newTestCoreString(TcfPolicyVersion23),
		DisclosedVendors: nil,
	}

	err := data.Validate()
	if err == nil {
		t.Errorf("Validate() should return error for v2.3 without DisclosedVendors")
		return
	}

	expected := "TCF v2.3: DisclosedVendors segment is mandatory for policyVersion >= 5"
	if err.Error() != expected {
		t.Errorf("Validate() error = %q, want %q", err.Error(), expected)
	}
}

// --- Test 6: Validate() v2.3 with DV ---
// Test Validate() returns nil for v2.3 with DV.
func TestValidateV23WithDisclosedVendors(t *testing.T) {
	data := &TCData{
		CoreString: newTestCoreString(TcfPolicyVersion23),
		DisclosedVendors: &DisclosedVendors{
			SegmentType:      int(SegmentTypeDisclosedVendors),
			MaxVendorId:      0,
			IsRangeEncoding:  false,
			DisclosedVendors: map[int]bool{},
		},
	}

	err := data.Validate()
	if err != nil {
		t.Errorf("Validate() should return nil for v2.3 with DisclosedVendors, got: %s", err)
	}
}

// --- Test 7: Validate() v2.2 without DV ---
// Test Validate() returns nil for v2.2 without DV.
func TestValidateV22WithoutDisclosedVendors(t *testing.T) {
	data := &TCData{
		CoreString:       newTestCoreString(4), // policyVersion=4
		DisclosedVendors: nil,
	}

	err := data.Validate()
	if err != nil {
		t.Errorf("Validate() should return nil for v2.2 without DisclosedVendors, got: %s", err)
	}
}

// --- Test 8: Validate() nil CoreString ---
// Test Validate() returns error for nil CoreString.
func TestValidateNilCoreString(t *testing.T) {
	data := &TCData{
		CoreString: nil,
	}

	err := data.Validate()
	if err == nil {
		t.Errorf("Validate() should return error for nil CoreString")
		return
	}

	expected := "core string is required"
	if err.Error() != expected {
		t.Errorf("Validate() error = %q, want %q", err.Error(), expected)
	}
}

// --- Test 9: Constants ---
// Verify TcfPolicyVersion23=5 and TcfVersion23=2.
func TestV23Constants(t *testing.T) {
	if TcfPolicyVersion23 != 5 {
		t.Errorf("TcfPolicyVersion23 = %d, want 5", TcfPolicyVersion23)
	}

	if TcfVersion23 != 2 {
		t.Errorf("TcfVersion23 = %d, want 2", TcfVersion23)
	}
}
