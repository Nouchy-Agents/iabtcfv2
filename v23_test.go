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

// newTestCoreStringAfterDeadline creates a CoreString with Created after the v2.3 deadline.
func newTestCoreStringAfterDeadline(policyVersion int) *CoreString {
	return &CoreString{
		Version:                2,
		Created:                time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), // After Feb 28, 2026
		LastUpdated:            time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
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

// --- Decode() deadline-based policy tests ---

// Test: Decode v2.3 TC string created BEFORE deadline without DV → ✅ should succeed
func TestDecodeV23BeforeDeadlineWithoutDV(t *testing.T) {
	core := newTestCoreString(TcfPolicyVersion23) // policyVersion=5, created=2022 (before deadline)
	tcString := core.Encode()

	data, err := Decode(tcString)
	if err != nil {
		t.Errorf("Decode() should succeed for v2.3 TC string before deadline without DV: %s", err)
		return
	}
	if data.CoreString == nil {
		t.Errorf("Decode() should return non-nil CoreString")
	}
	if data.CoreString.TcfPolicyVersion != TcfPolicyVersion23 {
		t.Errorf("CoreString.TcfPolicyVersion = %d, want %d", data.CoreString.TcfPolicyVersion, TcfPolicyVersion23)
	}
}

// Test: Decode v2.3 TC string created AFTER deadline without DV → ❌ should fail
func TestDecodeV23AfterDeadlineWithoutDV(t *testing.T) {
	core := newTestCoreStringAfterDeadline(TcfPolicyVersion23) // policyVersion=5, created=2026-03-01 (after deadline)
	tcString := core.Encode()

	_, err := Decode(tcString)
	if err == nil {
		t.Errorf("Decode() should return error for v2.3 TC string after deadline without DV")
		return
	}
	expected := "TCF v2.3: DisclosedVendors segment is mandatory for TC Strings created after 2026-02-28"
	if err.Error() != expected {
		t.Errorf("Decode() error = %q, want %q", err.Error(), expected)
	}
}

// Test: Decode v2.3 TC string created AFTER deadline with DV → ✅ should succeed
func TestDecodeV23AfterDeadlineWithDV(t *testing.T) {
	core := newTestCoreStringAfterDeadline(TcfPolicyVersion23)
	data := &TCData{
		CoreString: core,
		DisclosedVendors: &DisclosedVendors{
			SegmentType:      int(SegmentTypeDisclosedVendors),
			MaxVendorId:      0,
			IsRangeEncoding:  false,
			DisclosedVendors: map[int]bool{},
		},
	}
	tcString := data.ToTCString()

	decoded, err := Decode(tcString)
	if err != nil {
		t.Errorf("Decode() should succeed for v2.3 TC string after deadline with DV: %s", err)
		return
	}
	if decoded.DisclosedVendors == nil {
		t.Errorf("Decode() should return non-nil DisclosedVendors")
	}
}

// Test: Decode v2.2 TC string (policyVersion=4) created AFTER deadline → ❌ should fail
func TestDecodeV22AfterDeadline(t *testing.T) {
	core := newTestCoreStringAfterDeadline(4) // policyVersion=4, created after deadline
	tcString := core.Encode()

	_, err := Decode(tcString)
	if err == nil {
		t.Errorf("Decode() should return error for v2.2 TC string created after deadline")
		return
	}
	expected := "TCF v2.3: TC Strings created after 2026-02-28 must have policyVersion >= 5"
	if err.Error() != expected {
		t.Errorf("Decode() error = %q, want %q", err.Error(), expected)
	}
}

// Test: Decode v2.2 TC string created BEFORE deadline without DV → ✅ should succeed
func TestDecodeV22BeforeDeadlineWithoutDV(t *testing.T) {
	core := newTestCoreString(4) // policyVersion=4, created before deadline
	tcString := core.Encode()

	data, err := Decode(tcString)
	if err != nil {
		t.Errorf("Decode() should succeed for v2.2 TC string before deadline without DV: %s", err)
		return
	}
	if data.CoreString.TcfPolicyVersion != 4 {
		t.Errorf("CoreString.TcfPolicyVersion = %d, want 4", data.CoreString.TcfPolicyVersion)
	}
}

// --- IsVendorDisclosed tests ---

// Test: IsVendorDisclosed with bitfield
func TestIsVendorDisclosedBitfield(t *testing.T) {
	d := &DisclosedVendors{
		IsRangeEncoding:  false,
		MaxVendorId:      10,
		DisclosedVendors: map[int]bool{1: true, 3: true, 7: true},
	}
	if !d.IsVendorDisclosed(1) {
		t.Errorf("IsVendorDisclosed(1) should be true")
	}
	if !d.IsVendorDisclosed(3) {
		t.Errorf("IsVendorDisclosed(3) should be true")
	}
	if d.IsVendorDisclosed(2) {
		t.Errorf("IsVendorDisclosed(2) should be false")
	}
	if d.IsVendorDisclosed(5) {
		t.Errorf("IsVendorDisclosed(5) should be false")
	}
}

// Test: IsVendorDisclosed with range encoding
func TestIsVendorDisclosedRangeEncoding(t *testing.T) {
	d := &DisclosedVendors{
		IsRangeEncoding: true,
		RangeEntries: []*RangeEntry{
			{StartVendorID: 1, EndVendorID: 5},
			{StartVendorID: 10, EndVendorID: 10},
		},
	}
	if !d.IsVendorDisclosed(1) {
		t.Errorf("IsVendorDisclosed(1) should be true (in range 1-5)")
	}
	if !d.IsVendorDisclosed(5) {
		t.Errorf("IsVendorDisclosed(5) should be true (in range 1-5)")
	}
	if !d.IsVendorDisclosed(10) {
		t.Errorf("IsVendorDisclosed(10) should be true (single entry)")
	}
	if d.IsVendorDisclosed(6) {
		t.Errorf("IsVendorDisclosed(6) should be false (not in any range)")
	}
	if d.IsVendorDisclosed(9) {
		t.Errorf("IsVendorDisclosed(9) should be false (not in any range)")
	}
}

// Test: IsVendorDisclosed nil DisclosedVendors → false
func TestIsVendorDisclosedNil(t *testing.T) {
	data := &TCData{
		CoreString:       newTestCoreString(5),
		DisclosedVendors: nil,
	}
	if data.IsVendorDisclosed(1) {
		t.Errorf("IsVendorDisclosed(1) should be false when DisclosedVendors is nil")
	}
}

// --- IsPurposeLIAllowed(1) tests ---

// Test: IsPurposeLIAllowed(1) with v2.3 → false
func TestIsPurposeLIAllowed1WithV23(t *testing.T) {
	core := &CoreString{
		TcfPolicyVersion:       5,
		PurposesLITransparency: map[int]bool{1: true, 2: true},
	}
	if core.IsPurposeLIAllowed(1) {
		t.Errorf("IsPurposeLIAllowed(1) should be false for v2.3 (policyVersion=5), even if bit is set")
	}
}

// Test: IsPurposeLIAllowed(1) with v2.2 → PurposesLITransparency[1]
func TestIsPurposeLIAllowed1WithV22(t *testing.T) {
	core := &CoreString{
		TcfPolicyVersion:       4,
		PurposesLITransparency: map[int]bool{1: true, 2: true},
	}
	if !core.IsPurposeLIAllowed(1) {
		t.Errorf("IsPurposeLIAllowed(1) should be true for v2.2 (policyVersion=4) when bit is set")
	}
}

// Test: IsPurposeLIAllowed(2) with v2.3 → PurposesLITransparency[2]
func TestIsPurposeLIAllowed2WithV23(t *testing.T) {
	core := &CoreString{
		TcfPolicyVersion:       5,
		PurposesLITransparency: map[int]bool{1: true, 2: true},
	}
	if !core.IsPurposeLIAllowed(2) {
		t.Errorf("IsPurposeLIAllowed(2) should be true for v2.3 when bit is set")
	}
}

// --- IsV23() helper tests ---

// Test: IsV23() helper
func TestIsV23(t *testing.T) {
	dataV23 := &TCData{CoreString: &CoreString{TcfPolicyVersion: 5}}
	if !dataV23.IsV23() {
		t.Errorf("IsV23() should be true for policyVersion=5")
	}

	dataV22 := &TCData{CoreString: &CoreString{TcfPolicyVersion: 4}}
	if dataV22.IsV23() {
		t.Errorf("IsV23() should be false for policyVersion=4")
	}

	dataNil := &TCData{CoreString: nil}
	if dataNil.IsV23() {
		t.Errorf("IsV23() should be false for nil CoreString")
	}
}

// --- Validate() tests ---

// Test: Validate() matches Decode() policy
func TestValidateV23AfterDeadlineWithoutDV(t *testing.T) {
	data := &TCData{
		CoreString:       newTestCoreStringAfterDeadline(TcfPolicyVersion23),
		DisclosedVendors: nil,
	}
	err := data.Validate()
	if err == nil {
		t.Errorf("Validate() should return error for v2.3 after deadline without DV")
		return
	}
	expected := "TCF v2.3: DisclosedVendors segment is mandatory for TC Strings created after 2026-02-28"
	if err.Error() != expected {
		t.Errorf("Validate() error = %q, want %q", err.Error(), expected)
	}
}

func TestValidateV23AfterDeadlineWithDV(t *testing.T) {
	data := &TCData{
		CoreString: newTestCoreStringAfterDeadline(TcfPolicyVersion23),
		DisclosedVendors: &DisclosedVendors{
			SegmentType:      int(SegmentTypeDisclosedVendors),
			MaxVendorId:      0,
			IsRangeEncoding:  false,
			DisclosedVendors: map[int]bool{},
		},
	}
	err := data.Validate()
	if err != nil {
		t.Errorf("Validate() should return nil for v2.3 after deadline with DV, got: %s", err)
	}
}

func TestValidateV22AfterDeadline(t *testing.T) {
	data := &TCData{
		CoreString: newTestCoreStringAfterDeadline(4),
	}
	err := data.Validate()
	if err == nil {
		t.Errorf("Validate() should return error for v2.2 after deadline")
		return
	}
	expected := "TCF v2.3: TC Strings created after 2026-02-28 must have policyVersion >= 5"
	if err.Error() != expected {
		t.Errorf("Validate() error = %q, want %q", err.Error(), expected)
	}
}

func TestValidateV23BeforeDeadlineWithoutDV(t *testing.T) {
	data := &TCData{
		CoreString:       newTestCoreString(TcfPolicyVersion23), // before deadline
		DisclosedVendors: nil,
	}
	err := data.Validate()
	if err != nil {
		t.Errorf("Validate() should return nil for v2.3 before deadline without DV, got: %s", err)
	}
}

func TestValidateV22BeforeDeadlineWithoutDV(t *testing.T) {
	data := &TCData{
		CoreString:       newTestCoreString(4), // before deadline
		DisclosedVendors: nil,
	}
	err := data.Validate()
	if err != nil {
		t.Errorf("Validate() should return nil for v2.2 before deadline without DV, got: %s", err)
	}
}

func TestValidateNilCoreString(t *testing.T) {
	data := &TCData{CoreString: nil}
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

// --- DecodeLenient() tests ---

// Test: DecodeLenient with v2.3 after deadline without DV → succeeds
func TestDecodeLenientV23AfterDeadlineWithoutDV(t *testing.T) {
	core := newTestCoreStringAfterDeadline(TcfPolicyVersion23)
	tcString := core.Encode()

	data, err := DecodeLenient(tcString)
	if err != nil {
		t.Errorf("DecodeLenient() should succeed for v2.3 after deadline without DV: %s", err)
		return
	}
	if data.CoreString == nil {
		t.Errorf("DecodeLenient() should return non-nil CoreString")
	}
}

// Test: DecodeLenient invalid string → fails
func TestDecodeLenientInvalidString(t *testing.T) {
	_, err := DecodeLenient("!!!invalid!!!")
	if err == nil {
		t.Errorf("DecodeLenient() should return error for invalid string")
	}
}

// Test: DecodeLenient v2.2 before deadline without DV → succeeds
func TestDecodeLenientV22BeforeDeadlineWithoutDV(t *testing.T) {
	core := newTestCoreString(4)
	tcString := core.Encode()

	data, err := DecodeLenient(tcString)
	if err != nil {
		t.Errorf("DecodeLenient() should succeed for v2.2 before deadline without DV: %s", err)
		return
	}
	if data.CoreString.TcfPolicyVersion != 4 {
		t.Errorf("CoreString.TcfPolicyVersion = %d, want 4", data.CoreString.TcfPolicyVersion)
	}
}

// --- Round-trip encode/decode test ---

// Test: Round-trip encode/decode
func TestRoundTripEncodeDecode(t *testing.T) {
	core := newTestCoreString(TcfPolicyVersion23)
	dv := &DisclosedVendors{
		SegmentType:      int(SegmentTypeDisclosedVendors),
		MaxVendorId:      10,
		IsRangeEncoding:  false,
		DisclosedVendors: map[int]bool{1: true, 5: true, 10: true},
	}
	data := &TCData{
		CoreString:       core,
		DisclosedVendors: dv,
	}

	tcString := data.ToTCString()
	decoded, err := Decode(tcString)
	if err != nil {
		t.Errorf("Decode() should succeed for round-trip: %s", err)
		return
	}

	if decoded.CoreString.TcfPolicyVersion != TcfPolicyVersion23 {
		t.Errorf("Round-trip TcfPolicyVersion = %d, want %d", decoded.CoreString.TcfPolicyVersion, TcfPolicyVersion23)
	}
	if decoded.DisclosedVendors == nil {
		t.Errorf("Round-trip should have DisclosedVendors")
	}
}

// --- Encode tests (kept from original) ---

// Test: Encode v2.3 mandatory DV
func TestEncodeV23MandatoryDisclosedVendors(t *testing.T) {
	data := &TCData{
		CoreString:       newTestCoreString(TcfPolicyVersion23),
		DisclosedVendors: nil,
	}

	result := data.ToTCString()
	segments := countSegments(result)

	if segments < 2 {
		t.Errorf("ToTCString() should produce at least 2 segments for v2.3 (Core + DV), got %d segments: %s", segments, result)
	}

	if data.DisclosedVendors == nil {
		t.Errorf("ToTCString() should auto-create DisclosedVendors for v2.3")
	}

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

// Test: Encode v2.2 optional DV
func TestEncodeV22OptionalDisclosedVendors(t *testing.T) {
	data := &TCData{
		CoreString:       newTestCoreString(4),
		DisclosedVendors: nil,
	}

	result := data.ToTCString()
	segments := countSegments(result)

	if segments != 1 {
		t.Errorf("ToTCString() should produce exactly 1 segment for v2.2 without DV, got %d segments: %s", segments, result)
	}
}

// --- Constants test ---

// Test: Verify TcfV23Deadline constant value = 1772236800
func TestV23Constants(t *testing.T) {
	if TcfPolicyVersion23 != 5 {
		t.Errorf("TcfPolicyVersion23 = %d, want 5", TcfPolicyVersion23)
	}

	if TcfVersion23 != 2 {
		t.Errorf("TcfVersion23 = %d, want 2", TcfVersion23)
	}

	if TcfV23Deadline != 1772236800 {
		t.Errorf("TcfV23Deadline = %d, want 1772236800", TcfV23Deadline)
	}

	// Verify the constant corresponds to Feb 28, 2026 00:00:00 UTC
	deadlineTime := time.Unix(TcfV23Deadline, 0).UTC()
	expectedTime := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	if !deadlineTime.Equal(expectedTime) {
		t.Errorf("TcfV23Deadline time = %v, want %v", deadlineTime, expectedTime)
	}
}
