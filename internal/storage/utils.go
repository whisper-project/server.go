package storage

func SetIfMissing[T int64 | float64 | string | bool](loc *T, val T) {
	if *loc == *new(T) {
		*loc = val
	}
}
