package rope

// Node represents a node in the rope tree. A node is either an internal
// node with left/right children, or a leaf node that holds a substring.
type Node struct {
	// weight is the number of bytes in the left subtree. For leaf nodes it is
	// simply len(value).
	weight int
	left   *Node
	right  *Node
	value  string
}

// leaf creates a new leaf node containing the provided string.
func leaf(s string) *Node {
	return &Node{weight: len(s), value: s}
}

// len returns the number of bytes in the subtree rooted at n.
func (n *Node) len() int {
	if n == nil {
		return 0
	}
	if n.left == nil && n.right == nil {
		return len(n.value)
	}
	return n.weight + n.right.len()
}

// concat concatenates two nodes into a new parent node.
func concat(left, right *Node) *Node {
	if left == nil {
		return right
	}
	if right == nil {
		return left
	}
	return &Node{weight: left.len(), left: left, right: right}
}

// splitNode splits the node at the provided index and returns two new nodes.
func splitNode(n *Node, idx int) (*Node, *Node) {
	if n == nil {
		return nil, nil
	}
	if n.left == nil && n.right == nil {
		if idx <= 0 {
			return nil, n
		}
		if idx >= len(n.value) {
			return n, nil
		}
		return leaf(n.value[:idx]), leaf(n.value[idx:])
	}
	if idx < n.weight {
		l, r := splitNode(n.left, idx)
		return l, concat(r, n.right)
	}
	if idx > n.weight {
		l, r := splitNode(n.right, idx-n.weight)
		return concat(n.left, l), r
	}
	return n.left, n.right
}

// Rope describes the operations supported by a rope implementation.
type Rope interface {
	// Len returns the number of bytes stored in the rope.
	Len() int
	// Split splits the rope at the provided index and returns two new ropes.
	Split(idx int) (Rope, Rope)
	// Insert returns a new rope with s inserted at idx.
	Insert(idx int, s string) Rope
	// Delete returns a new rope with the range [start,end) removed.
	Delete(start, end int) Rope
	// String returns the contents of the rope as a string.
	String() string
	// Index returns the byte at position idx. ok will be false if idx is out of range.
	Index(idx int) (byte, bool)
}

// binaryRope is a binary rope structure for efficiently storing and editing large
// strings. It implements the Rope interface.
type binaryRope struct {
	root *Node
}

// New creates a new Rope containing the provided string.
func New(s string) Rope {
	return &binaryRope{root: leaf(s)}
}

// Len returns the number of bytes stored in the rope.
func (r *binaryRope) Len() int {
	if r == nil || r.root == nil {
		return 0
	}
	return r.root.len()
}

// Concat returns a new rope that is the concatenation of r1 and r2.
func Concat(r1, r2 Rope) Rope {
	br1, _ := r1.(*binaryRope)
	br2, _ := r2.(*binaryRope)
	if br1 == nil || br1.root == nil {
		return r2
	}
	if br2 == nil || br2.root == nil {
		return r1
	}
	return &binaryRope{root: concat(br1.root, br2.root)}
}

// Split splits the rope at the provided index and returns two new ropes.
func (r *binaryRope) Split(idx int) (Rope, Rope) {
	if r == nil {
		return nil, nil
	}
	l, rgt := splitNode(r.root, idx)
	return &binaryRope{root: l}, &binaryRope{root: rgt}
}

// Insert returns a new rope with s inserted at idx.
func (r *binaryRope) Insert(idx int, s string) Rope {
	left, right := r.Split(idx)
	return Concat(Concat(left, New(s)), right)
}

// Delete returns a new rope with the range [start,end) removed.
func (r *binaryRope) Delete(start, end int) Rope {
	if start >= end {
		return r
	}
	left, rest := r.Split(start)
	_, right := rest.Split(end - start)
	return Concat(left, right)
}

// String returns the full contents of the rope as a string.
func (r *binaryRope) String() string {
	var b []byte
	var traverse func(n *Node)
	traverse = func(n *Node) {
		if n == nil {
			return
		}
		if n.left == nil && n.right == nil {
			b = append(b, n.value...)
			return
		}
		traverse(n.left)
		traverse(n.right)
	}
	traverse(r.root)
	return string(b)
}

// Index returns the byte at position idx. It returns ok=false if the index is
// out of range.
func (r *binaryRope) Index(idx int) (byte, bool) {
	if r == nil || idx < 0 || idx >= r.Len() {
		return 0, false
	}
	n := r.root
	for n.left != nil || n.right != nil {
		if idx < n.weight {
			n = n.left
		} else {
			idx -= n.weight
			n = n.right
		}
	}
	if idx < len(n.value) {
		return n.value[idx], true
	}
	return 0, false
}
