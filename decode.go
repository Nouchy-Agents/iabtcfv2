package iabtcfv2

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// Decodes a string and returns the TcfVersion
// It can also decode version from a TCF V1.1 consent string
// - TcfVersionUndefined = -1
// - TcfVersion1 = 1
// - TcfVersion2 = 2
func GetVersion(s string) (version TcfVersion, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	segments := strings.Split(s, ".")
	if len(segments) == 0 {
		return TcfVersionUndefined, err
	}

	b, err := base64.RawURLEncoding.DecodeString(segments[0])
	if err != nil {
		return TcfVersionUndefined, err
	}

	var e = NewTCEncoder(b)
	return TcfVersion(e.ReadInt(bitsVersion)), nil
}

// Decodes a segment value and returns the SegmentType
// - SegmentTypeUndefined = -1
// - SegmentTypeCoreString = 0
// - SegmentTypeDisclosedVendors = 1
// - SegmentTypePublisherTC = 3
func GetSegmentType(segment string) (segmentType SegmentType, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	b, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		return SegmentTypeUndefined, err
	}

	var e = NewTCEncoder(b)
	return SegmentType(e.ReadInt(bitsSegmentType)), nil
}

// Decode a TC String and returns it as a TCData structure
// A valid TC String must start with a Core String segment.
// Subsequent segments (DisclosedVendors, PublisherTC) can appear in any order,
// identified by their SegmentType field.
// TCF v2.3 (policyVersion >= 5): DisclosedVendors segment is MANDATORY.
func Decode(tcString string) (t *TCData, err error) {
	t = &TCData{}
	mapSegments := map[SegmentType]bool{}
	segmentOrder := []SegmentType{}

	for i, v := range strings.Split(tcString, ".") {
		segmentType, err := GetSegmentType(v)
		if err != nil {
			return nil, err
		}

		switch segmentType {
		case SegmentTypeDisclosedVendors:
			if mapSegments[SegmentTypeDisclosedVendors] == true {
				return nil, fmt.Errorf("duplicate Disclosed Vendors segment")
			}
			segment, err := DecodeDisclosedVendors(v)
			if err == nil {
				t.DisclosedVendors = segment
				mapSegments[SegmentTypeDisclosedVendors] = true
				segmentOrder = append(segmentOrder, SegmentTypeDisclosedVendors)
			}
			break
		case SegmentTypePublisherTC:
			if mapSegments[SegmentTypePublisherTC] == true {
				return nil, fmt.Errorf("duplicate Publisher TC segment")
			}
			segment, err := DecodePublisherTC(v)
			if err == nil {
				t.PublisherTC = segment
				mapSegments[SegmentTypePublisherTC] = true
				segmentOrder = append(segmentOrder, SegmentTypePublisherTC)
			}
			break
		default:
			if mapSegments[SegmentTypeCoreString] == true {
				return nil, fmt.Errorf("duplicate Core String segment")
			}
			segment, err := DecodeCoreString(v)
			if err == nil {
				t.CoreString = segment
				if i == 0 {
					mapSegments[SegmentTypeCoreString] = true
					segmentOrder = append(segmentOrder, SegmentTypeCoreString)
				}
			}
			break
		}
	}

	if mapSegments[SegmentTypeCoreString] == false {
		return nil, fmt.Errorf("invalid TC string")
	}

	// TCF v2.3: After Feb 28, 2026, TC strings must have TcfPolicyVersion >= 5
	// AND a DisclosedVendors segment. Before March 1, 2026, DV is optional.
	if t.CoreString.Created.After(v23Deadline) {
		if t.CoreString.TcfPolicyVersion < TcfPolicyVersion23 {
			return nil, fmt.Errorf("TCF v2.3: TC Strings created after %s must have policyVersion >= %d", v23Deadline.Format("2006-01-02"), TcfPolicyVersion23)
		}
		if !mapSegments[SegmentTypeDisclosedVendors] {
			return nil, fmt.Errorf("TCF v2.3: DisclosedVendors segment is mandatory for TC Strings created after %s", v23Deadline.Format("2006-01-02"))
		}
		// Core String must be the first segment
		if segmentOrder[0] != SegmentTypeCoreString {
			return nil, fmt.Errorf("TCF v2.3: Core String must be the first segment")
		}
	}

	return t, nil
}

// DecodeLenient decodes a TC String without enforcing TCF v2.3 deadline/policyVersion checks.
// Accepts any well-formed TC string regardless of creation date or policy version.
func DecodeLenient(tcString string) (t *TCData, err error) {
	t = &TCData{}
	mapSegments := map[SegmentType]bool{}
	segmentOrder := []SegmentType{}

	for i, v := range strings.Split(tcString, ".") {
		segmentType, err := GetSegmentType(v)
		if err != nil {
			return nil, err
		}

		switch segmentType {
		case SegmentTypeDisclosedVendors:
			if mapSegments[SegmentTypeDisclosedVendors] {
				return nil, fmt.Errorf("duplicate Disclosed Vendors segment")
			}
			segment, err := DecodeDisclosedVendors(v)
			if err == nil {
				t.DisclosedVendors = segment
				mapSegments[SegmentTypeDisclosedVendors] = true
				segmentOrder = append(segmentOrder, SegmentTypeDisclosedVendors)
			}
		case SegmentTypePublisherTC:
			if mapSegments[SegmentTypePublisherTC] {
				return nil, fmt.Errorf("duplicate Publisher TC segment")
			}
			segment, err := DecodePublisherTC(v)
			if err == nil {
				t.PublisherTC = segment
				mapSegments[SegmentTypePublisherTC] = true
				segmentOrder = append(segmentOrder, SegmentTypePublisherTC)
			}
		default:
			if mapSegments[SegmentTypeCoreString] {
				return nil, fmt.Errorf("duplicate Core String segment")
			}
			segment, err := DecodeCoreString(v)
			if err == nil {
				t.CoreString = segment
				if i == 0 {
					mapSegments[SegmentTypeCoreString] = true
					segmentOrder = append(segmentOrder, SegmentTypeCoreString)
				}
			}
		}
	}

	if !mapSegments[SegmentTypeCoreString] {
		return nil, fmt.Errorf("invalid TC string")
	}

	// NO deadline/policyVersion checks — lenient mode
	_ = segmentOrder
	return t, nil
}


// Decodes a Core String value and returns it as a CoreString structure
func DecodeCoreString(coreString string) (c *CoreString, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	b, err := base64.RawURLEncoding.DecodeString(coreString)
	if err != nil {
		return nil, err
	}

	var e = NewTCEncoder(b)

	c = &CoreString{}
	c.Version = e.ReadInt(bitsVersion)
	c.Created = e.ReadTime()
	c.LastUpdated = e.ReadTime()
	c.CmpId = e.ReadInt(bitsCmpId)
	c.CmpVersion = e.ReadInt(bitsCmpVersion)
	c.ConsentScreen = e.ReadInt(bitsConsentScreen)
	c.ConsentLanguage = e.ReadChars(bitsConsentLanguage)
	c.VendorListVersion = e.ReadInt(bitsVendorListVersion)
	c.TcfPolicyVersion = e.ReadInt(bitsTcfPolicyVersion)
	c.IsServiceSpecific = e.ReadBool()
	c.UseNonStandardTexts = e.ReadBool()
	c.SpecialFeatureOptIns = e.ReadBitField(bitsSpecialFeatureOptIns)
	c.PurposesConsent = e.ReadBitField(bitsPurposesConsent)
	c.PurposesLITransparency = e.ReadBitField(bitsPurposesLITransparency)
	c.PurposeOneTreatment = e.ReadBool()
	c.PublisherCC = e.ReadChars(bitsPublisherCC)

	c.MaxVendorId = e.ReadInt(bitsMaxVendorId)
	c.IsRangeEncoding = e.ReadBool()
	if c.IsRangeEncoding {
		c.NumEntries, c.RangeEntries = e.ReadRangeEntries()
	} else {
		c.VendorsConsent = e.ReadBitField(uint(c.MaxVendorId))
	}

	c.MaxVendorIdLI = e.ReadInt(bitsMaxVendorId)
	c.IsRangeEncodingLI = e.ReadBool()
	if c.IsRangeEncodingLI {
		c.NumEntriesLI, c.RangeEntriesLI = e.ReadRangeEntries()
	} else {
		c.VendorsLITransparency = e.ReadBitField(uint(c.MaxVendorIdLI))
	}

	c.NumPubRestrictions, c.PubRestrictions = e.ReadPubRestrictions()

	return c, nil
}

// Decodes a Disclosed Vendors value and returns it as a DisclosedVendors structure
func DecodeDisclosedVendors(disclosedVendors string) (d *DisclosedVendors, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	b, err := base64.RawURLEncoding.DecodeString(disclosedVendors)
	if err != nil {
		return nil, err
	}

	var e = NewTCEncoder(b)

	d = &DisclosedVendors{}
	d.SegmentType = e.ReadInt(bitsSegmentType)
	d.MaxVendorId = e.ReadInt(bitsMaxVendorId)
	d.IsRangeEncoding = e.ReadBool()
	if d.IsRangeEncoding {
		d.NumEntries, d.RangeEntries = e.ReadRangeEntries()
	} else {
		d.DisclosedVendors = e.ReadBitField(uint(d.MaxVendorId))
	}

	if d.SegmentType != int(SegmentTypeDisclosedVendors) {
		err = fmt.Errorf("disclosed vendors segment type must be %d", SegmentTypeDisclosedVendors)
		return nil, err
	}

	return d, nil
}

// Decodes a Publisher TC value and returns it as a PublisherTC structure
func DecodePublisherTC(publisherTC string) (p *PublisherTC, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	b, err := base64.RawURLEncoding.DecodeString(publisherTC)
	if err != nil {
		return nil, err
	}

	var e = NewTCEncoder(b)

	p = &PublisherTC{}
	p.SegmentType = e.ReadInt(bitsSegmentType)
	p.PubPurposesConsent = e.ReadBitField(bitsPubPurposesConsent)
	p.PubPurposesLITransparency = e.ReadBitField(bitsPubPurposesLITransparency)
	p.NumCustomPurposes = e.ReadInt(bitsNumCustomPurposes)
	p.CustomPurposesConsent = e.ReadBitField(uint(p.NumCustomPurposes))
	p.CustomPurposesLITransparency = e.ReadBitField(uint(p.NumCustomPurposes))

	if p.SegmentType != int(SegmentTypePublisherTC) {
		return nil, fmt.Errorf("publisher TC segment type must be %d", SegmentTypePublisherTC)
	}

	return p, nil
}
