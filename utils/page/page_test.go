package page_test

import (
	"testing"

	"github.com/azcov/gokit/utils/page"
)

func TestNew_Defaults(t *testing.T) {
	p := page.New(0, 0, 100)
	if p.Number != 1 {
		t.Errorf("expected page 1, got %d", p.Number)
	}
	if p.Size != 20 {
		t.Errorf("expected size 20, got %d", p.Size)
	}
}

func TestNew_ClampsMaxSize(t *testing.T) {
	p := page.New(1, 500, 1000)
	if p.Size != 100 {
		t.Errorf("expected size clamped to 100, got %d", p.Size)
	}
}

func TestOffset(t *testing.T) {
	cases := []struct {
		page, size, want int
	}{
		{1, 10, 0},
		{2, 10, 10},
		{3, 25, 50},
	}
	for _, tc := range cases {
		p := page.New(tc.page, tc.size, 1000)
		if got := p.Offset(); got != tc.want {
			t.Errorf("page=%d size=%d: expected offset %d, got %d", tc.page, tc.size, tc.want, got)
		}
	}
}

func TestTotalPages(t *testing.T) {
	cases := []struct {
		total, size, want int
	}{
		{100, 10, 10},
		{101, 10, 11},
		{0, 10, 0},
		{10, 10, 1},
	}
	for _, tc := range cases {
		p := page.New(1, tc.size, tc.total)
		if got := p.TotalPages(); got != tc.want {
			t.Errorf("total=%d size=%d: expected %d pages, got %d", tc.total, tc.size, tc.want, got)
		}
	}
}

func TestHasNext(t *testing.T) {
	p := page.New(1, 10, 25)
	if !p.HasNext() {
		t.Error("expected HasNext=true on page 1 of 3")
	}
	last := page.New(3, 10, 25)
	if last.HasNext() {
		t.Error("expected HasNext=false on last page")
	}
}

func TestHasPrev(t *testing.T) {
	p := page.New(1, 10, 50)
	if p.HasPrev() {
		t.Error("expected HasPrev=false on first page")
	}
	p2 := page.New(2, 10, 50)
	if !p2.HasPrev() {
		t.Error("expected HasPrev=true on page 2")
	}
}

func TestNext_Clamp(t *testing.T) {
	last := page.New(5, 10, 50)
	next := last.Next()
	if next.Number != 5 {
		t.Errorf("expected clamped to 5, got %d", next.Number)
	}
}

func TestPrev_Clamp(t *testing.T) {
	first := page.New(1, 10, 50)
	prev := first.Prev()
	if prev.Number != 1 {
		t.Errorf("expected clamped to 1, got %d", prev.Number)
	}
}

func TestNewResponse(t *testing.T) {
	items := []string{"a", "b", "c"}
	p := page.New(2, 3, 9)
	resp := page.NewResponse(items, p)

	if resp.Page != 2 {
		t.Errorf("expected page 2, got %d", resp.Page)
	}
	if resp.TotalPages != 3 {
		t.Errorf("expected 3 total pages, got %d", resp.TotalPages)
	}
	if !resp.HasPrev {
		t.Error("expected HasPrev=true")
	}
	if !resp.HasNext {
		t.Error("expected HasNext=true")
	}
	if len(resp.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(resp.Items))
	}
}
