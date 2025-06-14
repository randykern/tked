package rope

import "testing"

func TestNewLenString(t *testing.T) {
	r := New("hello world")
	if r.Len() != len("hello world") {
		t.Fatalf("expected length %d got %d", len("hello world"), r.Len())
	}
	if got := r.String(); got != "hello world" {
		t.Fatalf("expected %q got %q", "hello world", got)
	}
}

func TestConcat(t *testing.T) {
	r1 := New("hello ")
	r2 := New("world")
	r := Concat(r1, r2)
	if got := r.String(); got != "hello world" {
		t.Fatalf("expected %q got %q", "hello world", got)
	}
	if r.Len() != len("hello world") {
		t.Fatalf("expected length %d got %d", len("hello world"), r.Len())
	}
}

func TestSplit(t *testing.T) {
	r := New("hello world")
	left, right := r.Split(5)
	if got := left.String(); got != "hello" {
		t.Fatalf("expected left %q got %q", "hello", got)
	}
	if got := right.String(); got != " world" {
		t.Fatalf("expected right %q got %q", " world", got)
	}
}

func TestInsert(t *testing.T) {
	r := New("helloworld")
	r = r.Insert(5, " ")
	if got := r.String(); got != "hello world" {
		t.Fatalf("expected %q got %q", "hello world", got)
	}
	// insert at beginning
	r = r.Insert(0, "start ")
	if got := r.String(); got != "start hello world" {
		t.Fatalf("expected %q got %q", "start hello world", got)
	}
	// insert at end
	r = r.Insert(r.Len(), " end")
	if got := r.String(); got != "start hello world end" {
		t.Fatalf("expected %q got %q", "start hello world end", got)
	}
}

func TestDelete(t *testing.T) {
	r := New("hello world")
	r = r.Delete(5, 6)
	if got := r.String(); got != "helloworld" {
		t.Fatalf("expected %q got %q", "helloworld", got)
	}
	// deleting with start>=end should be no-op
	r2 := r.Delete(3, 3)
	if got := r2.String(); got != "helloworld" {
		t.Fatalf("expected %q got %q", "helloworld", got)
	}
}

func TestIndex(t *testing.T) {
	data := "hello"
	r := New(data)
	for i := 0; i < len(data); i++ {
		b, ok := r.Index(i)
		if !ok {
			t.Fatalf("expected ok for index %d", i)
		}
		if byte(data[i]) != b {
			t.Fatalf("expected %c got %c", data[i], b)
		}
	}
	if _, ok := r.Index(-1); ok {
		t.Fatalf("expected false for negative index")
	}
	if _, ok := r.Index(len(data)); ok {
		t.Fatalf("expected false for out of range index")
	}
}
