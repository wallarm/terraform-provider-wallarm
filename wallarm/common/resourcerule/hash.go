package resourcerule

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"

	wallarm "github.com/wallarm/wallarm-go"
)

// ConditionsHash computes a deterministic SHA256 hash of action conditions.
// Port of Ruby's Action.calculate_conditions_hash:
//
//	serialized = conditions.map { |c| c.to_serialized_array }
//	Digest::SHA2.hexdigest("not_null#{serialized.sort.join(',')}")
//
// Used for action matching, directory naming, and deduplication.
func ConditionsHash(conditions []wallarm.ActionDetails) string {
	serialized := make([]string, len(conditions))
	for i, c := range conditions {
		serialized[i] = serializeCondition(c)
	}
	sort.Strings(serialized)
	payload := "not_null" + strings.Join(serialized, ",")
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

// PointHash computes SHA256 of a raw_packed point array.
// Port of Ruby's HasPoint.calculate_point_hash:
//
//	Digest::SHA2.hexdigest(JSON.raw_pack(point_array))
func PointHash(point []interface{}) string {
	packed := rawPack(point)
	sum := sha256.Sum256([]byte(packed))
	return hex.EncodeToString(sum[:])
}

// serializeCondition replicates Ruby's Condition#to_serialized_array.
// Produces a sorted array of key-value pairs as a rawPack string:
//
//	[["point","<raw_packed_point>"],["type","<type>"],["value","<value_or_null>"]]
//
// Keys are always alphabetical: point < type < value.
func serializeCondition(c wallarm.ActionDetails) string {
	// Step 1: raw_pack the point array (inner encoding)
	pointPacked := rawPack(interfaceSlice(c.Point))

	// Step 2: determine value — nil becomes null, strings get quoted
	var valuePart string
	if c.Value == nil {
		valuePart = "null"
	} else {
		valuePart = rawPack(c.Value)
	}

	// Step 3: build sorted [["key","val"],...] and raw_pack the outer array
	// "point" < "type" < "value" — alphabetical order guaranteed
	return "[" +
		"[" + rawPack("point") + "," + rawPack(pointPacked) + "]," +
		"[" + rawPack("type") + "," + rawPack(c.Type) + "]," +
		"[" + rawPack("value") + "," + valuePart + "]" +
		"]"
}

// rawPack is a port of Ruby's JSON.raw_pack — a deterministic JSON serializer
// that produces compact output without whitespace.
//
//	Array  → [elem,elem,...]
//	String → "escaped_string"
//	int/float → number
//	nil    → null
func rawPack(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return "null"
	case string:
		b, _ := json.Marshal(val)
		return string(b)
	case int:
		return fmt.Sprintf("%d", val)
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%g", val)
	case json.Number:
		return val.String()
	case []interface{}:
		parts := make([]string, len(val))
		for i, elem := range val {
			parts[i] = rawPack(elem)
		}
		return "[" + strings.Join(parts, ",") + "]"
	default:
		// Unexpected type — log for debugging. This should not happen with
		// well-formed ActionDetails from the API or provider.
		log.Printf("[WARN] rawPack: unexpected type %T for value %v", val, val)
		return fmt.Sprintf("%v", val)
	}
}

// interfaceSlice converts []interface{} to itself (passthrough) or handles
// the case where Point might need conversion.
func interfaceSlice(s []interface{}) interface{} {
	if s == nil {
		return []interface{}{}
	}
	return s
}
