// GraphQL
/**
input Sort {
    field: String!
    direction: String
}

input Pagination {
    page: Int
    perPage: Int
    sort: Sort
}
*/

type Pagination struct {
	Page    *int
	PerPage *int
	Sort    *Sort
}

type Sort struct {
	Field   string
	Direction *string
}