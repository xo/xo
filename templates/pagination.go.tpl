type Pagination struct {
	Page    *uint64
	PerPage *uint64
	Sort    *Sort
}

type Sort struct {
	Field   string
	Direction *string
}