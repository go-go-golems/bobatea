package inputhistory

import "testing"

func TestHistory(t *testing.T) {
	history := NewHistory(5)

	history.Add("input1", "output1", false)
	history.Add("input2", "output2", true)

	entries := history.GetEntries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Input != "input1" || entries[0].Output != "output1" || entries[0].IsErr {
		t.Fatalf("unexpected first entry: %#v", entries[0])
	}
	if entries[1].Input != "input2" || entries[1].Output != "output2" || !entries[1].IsErr {
		t.Fatalf("unexpected second entry: %#v", entries[1])
	}

	if history.IsNavigating() {
		t.Fatalf("expected not navigating")
	}
	if up := history.NavigateUp(); up != "input2" {
		t.Fatalf("expected input2, got %q", up)
	}
	if !history.IsNavigating() {
		t.Fatalf("expected navigating")
	}
	if up := history.NavigateUp(); up != "input1" {
		t.Fatalf("expected input1, got %q", up)
	}
	if down := history.NavigateDown(); down != "input2" {
		t.Fatalf("expected input2, got %q", down)
	}
	if down := history.NavigateDown(); down != "" {
		t.Fatalf("expected empty string, got %q", down)
	}
	if history.IsNavigating() {
		t.Fatalf("expected not navigating")
	}

	history.Clear()
	if len(history.GetEntries()) != 0 {
		t.Fatalf("expected no entries after clear")
	}
}
