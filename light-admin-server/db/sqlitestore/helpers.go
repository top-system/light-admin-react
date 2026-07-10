package sqlitestore

// Hand-written conversion helpers used by the generated adapter (store.go).
// They bridge the engine-neutral store types onto the looser types sqlc's
// sqlite engine infers (interface{} for nullable named args, non-null int64
// for LIMIT/OFFSET).

// csvList serializes a nullable id-list filter for the comma-wrapped instr()
// membership test used by the sqlite queries: instr(list, ',' || col || ',').
// nil disables the filter (SQL NULL); a non-nil slice becomes ",a,b,c," so
// every element is matched with exact comma boundaries. An empty non-nil
// slice yields "," which matches nothing — the same semantics as the postgres
// `col = ANY('{}')`.
func csvList(ids []string) interface{} {
	if ids == nil {
		return nil
	}
	out := ","
	for _, id := range ids {
		out += id + ","
	}
	return out
}

// limitOrAll maps the store's nullable limit onto SQLite's non-null LIMIT
// parameter: nil means "no limit", which SQLite spells LIMIT -1.
func limitOrAll(limit *int32) int64 {
	if limit == nil {
		return -1
	}
	return int64(*limit)
}

// offsetOrZero maps the store's nullable offset onto SQLite's non-null OFFSET
// parameter: nil means "start at the first row".
func offsetOrZero(offset *int32) int64 {
	if offset == nil {
		return 0
	}
	return int64(*offset)
}
