package page

// Page holds pagination parameters and computes derived values.
type Page struct {
	Number int // 1-based page number
	Size   int // items per page
	Total  int // total item count
}

const defaultPageSize = 20
const maxPageSize = 100

// New creates a Page with sensible defaults and clamped values.
func New(number, size, total int) Page {
	if size <= 0 {
		size = defaultPageSize
	}
	if size > maxPageSize {
		size = maxPageSize
	}
	if number <= 0 {
		number = 1
	}
	return Page{Number: number, Size: size, Total: total}
}

// Offset returns the SQL-style row offset for this page.
func (p Page) Offset() int { return (p.Number - 1) * p.Size }

// Limit returns the page size (alias for clarity in queries).
func (p Page) Limit() int { return p.Size }

// TotalPages returns the total number of pages.
func (p Page) TotalPages() int {
	if p.Total <= 0 || p.Size <= 0 {
		return 0
	}
	return (p.Total + p.Size - 1) / p.Size
}

// HasNext reports whether a next page exists.
func (p Page) HasNext() bool { return p.Number < p.TotalPages() }

// HasPrev reports whether a previous page exists.
func (p Page) HasPrev() bool { return p.Number > 1 }

// Next returns a Page for the next page, clamped to TotalPages.
func (p Page) Next() Page {
	if p.HasNext() {
		return Page{Number: p.Number + 1, Size: p.Size, Total: p.Total}
	}
	return p
}

// Prev returns a Page for the previous page, clamped to 1.
func (p Page) Prev() Page {
	if p.HasPrev() {
		return Page{Number: p.Number - 1, Size: p.Size, Total: p.Total}
	}
	return p
}

// Response is a generic paginated response wrapper.
type Response[T any] struct {
	Items      []T  `json:"items"`
	Page       int  `json:"page"`
	Size       int  `json:"size"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// NewResponse wraps items with pagination metadata.
func NewResponse[T any](items []T, p Page) Response[T] {
	return Response[T]{
		Items:      items,
		Page:       p.Number,
		Size:       p.Size,
		Total:      p.Total,
		TotalPages: p.TotalPages(),
		HasNext:    p.HasNext(),
		HasPrev:    p.HasPrev(),
	}
}
