type Pagination struct {
	Page    *int
	PerPage *int
	Sort    *Sort
}

type Sort struct {
	Field *string
	Order *string
}