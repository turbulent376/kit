package common

type SortRequest struct {
	Field     string
	Direction string
}

type PagingRequest struct {
	Size   int
	Index  int
	SortBy []*SortRequest
}

type PagingResponse struct {
	Total int
	Index int
}