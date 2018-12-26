type Pagination struct {
	Page    *int
	PerPage *int
	Sort    []string
}

type ListMetadata struct {
    Count int `db:"count"`
}