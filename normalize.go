package igbinary

import "strconv"

// NormalizeArrays recursively walks a decoded igbinary value and converts
// map[string]any whose keys are sequential integers ("0","1","2",…) into
// []any slices, matching PHP's JSON encoding of sequential arrays.
//
// PHP does not distinguish between indexed arrays and associative arrays at
// the igbinary level — both are encoded as key-value maps. When decoded by
// [Decode], a PHP indexed array like ["a","b","c"] becomes
// map[string]any{"0":"a","1":"b","2":"c"}. This function converts such
// maps back into Go slices so that JSON serialization produces JSON arrays.
//
// Rules:
//   - An empty map is treated as an empty PHP array and is converted to []any{}.
//   - A map whose keys are exactly "0","1",…,"N-1" is converted to a []any of
//     length N with values in index order, recursively normalized.
//   - All other maps are left as-is but their values are recursively normalized.
//   - Slices have their elements recursively normalized.
//   - Non-map, non-slice values are returned unchanged.
//
// This is a pure function — it does not modify the original value when the
// top-level type is a slice or scalar value. When the top-level type is a map
// that is NOT sequential, values are normalized in-place for efficiency.
func NormalizeArrays(v any) any {
	switch val := v.(type) {
	case map[string]any:
		return normalizeMap(val)
	case []any:
		return normalizeSlice(val)
	default:
		return v
	}
}

// WithNormalizeArrays enables automatic array normalization after decoding.
// When enabled, the decoder applies [NormalizeArrays] to the decoded result
// before returning it, converting PHP sequential arrays (decoded as maps with
// "0","1",… keys) into Go slices.
//
// This is useful when the decoded data will be serialized to JSON, where the
// distinction between arrays and objects matters.
func WithNormalizeArrays() Option {
	return func(d *Decoder) {
		d.normalize = true
	}
}

// normalizeMap checks whether the map has sequential integer keys (0..n-1).
// If so, it returns a []any slice with the values in order. Otherwise, it
// returns the same map with values recursively normalized.
func normalizeMap(m map[string]any) any {
	n := len(m)

	// Empty map → empty slice (PHP empty array [] serialized as igbinary array
	// with zero entries decodes as map[string]any{}).
	if n == 0 {
		return []any{}
	}

	// Check whether every key is a sequential integer from 0 to n-1.
	isSequential := true
	for i := 0; i < n; i++ {
		if _, ok := m[strconv.Itoa(i)]; !ok {
			isSequential = false
			break
		}
	}

	if isSequential {
		slice := make([]any, n)
		for i := 0; i < n; i++ {
			slice[i] = NormalizeArrays(m[strconv.Itoa(i)])
		}
		return slice
	}

	// Not sequential — normalize values in-place.
	for k, v := range m {
		m[k] = NormalizeArrays(v)
	}
	return m
}

// normalizeSlice recursively normalizes each element of a slice.
func normalizeSlice(s []any) []any {
	for i, v := range s {
		s[i] = NormalizeArrays(v)
	}
	return s
}
