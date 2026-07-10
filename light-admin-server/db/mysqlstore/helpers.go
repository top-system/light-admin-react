package mysqlstore

import "math"

// Hand-written conversion helpers used by the generated adapter (store.go).
// They bridge the engine-neutral store types onto the types sqlc's mysql
// engine infers (pointer named args, non-null int32 LIMIT/OFFSET).

// csvList serializes a nullable id-list filter for the FIND_IN_SET membership
// test used by the mysql queries: FIND_IN_SET(col, list). nil disables the
// filter (SQL NULL); a non-nil slice becomes "a,b,c". An empty non-nil slice
// yields "", which FIND_IN_SET never matches — the same semantics as the
// postgres `col = ANY('{}')`.
func csvList(ids []string) *string {
	if ids == nil {
		return nil
	}
	out := ""
	for i, id := range ids {
		if i > 0 {
			out += ","
		}
		out += id
	}
	return &out
}

// limitOrAll maps the store's nullable limit onto MySQL's non-null LIMIT
// parameter: nil means "no limit", which MySQL has no spelling for, so the
// largest representable row count serves as the sentinel.
func limitOrAll(limit *int32) int32 {
	if limit == nil {
		return math.MaxInt32
	}
	return *limit
}

// offsetOrZero maps the store's nullable offset onto MySQL's non-null OFFSET
// parameter: nil means "start at the first row".
func offsetOrZero(offset *int32) int32 {
	if offset == nil {
		return 0
	}
	return *offset
}
