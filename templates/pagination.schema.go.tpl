input Sort {
    field: String!
    direction: String
}

input Pagination {
    page: Int
    perPage: Int
    sort: Sort
}