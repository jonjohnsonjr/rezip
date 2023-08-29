package rezip

import (
	"bytes"
	"io"
)

type Rewriter struct {
	zr io.Reader

	w      io.Writer
	window []byte // last 32KB uncompressed written to w

	z       io.Reader
	zwindow []byte // last 32KB uncompressed from z

	paused bool
}

// NewRewriter is like a gzip.Writer that also takes an already compressed version.
// It can use the already compressed version as a cheatsheet for bytes that get written
// to it.
func NewRewriter(w io.Writer, z io.Reader) (*Rewriter, error) {
	return &Rewriter{
		w: w,
		z: z,
	}, nil
}

func (r *Rewriter) writeFast(p []byte) (int, error) {
	// TODO: We may need to buffer more than one block depending on len(p).
	block, err := r.nextBlock()
	if err != nil {
		return 0, err
	}

	if !bytes.Equal(p, block.b) {
		return r.writeSlow(p)
	}

	if block.btype == 0 {
		// We don't need to worry about pointers for literal blocks.
		return r.w.Write(block.zb)
	}

	// TODO: We actually don't have to look at the entire window,
	// just the intersection of all the pointers in block.
	if !bytes.Equal(r.window, r.zwindow) {
		return r.writeSlow(p)
	}

	if block.btype == 1 {
		// We don't need to worry about trees for fixed blocks.
		return r.w.Write(block.zb)
	}

	// TODO: What do we do with trees?
	return r.w.Write(block.zb)
}

func (r *Rewriter) Write(p []byte) (int, error) {
	if !r.paused {
		return r.writeFast(p)
	}

	// This is very slow, do our best to realign.
	return r.writeSlow(p)
}

func (r *Rewriter) Reset(z io.Reader) error {
	// TODO: How do we reset zwindow?
	return nil
}

// Pause stops using the cheatsheet.
func (r *Rewriter) Pause() {
	r.paused = true
}

// Resume resumes using the cheatsheet.
func (r *Rewriter) Resume() {
	r.paused = false
}

// Discard decompresses n bytes from z and discards them.
func (r *Rewriter) Discard(n int64) error {
	return nil
}

func (r *Rewriter) Close() error {
	// TODO: Same as gzip.Writer.Close, we need to update isize and checksum.
	return nil
}

func (r *Rewriter) nextBlock() (block, error) {
	return block{}, nil
}

func (r *Rewriter) writeSlow(p []byte) (int, error) {
	return 0, nil
}

// TODO: This section is just pseudocode notes.
type tree struct {
	hlit  byte
	hdist byte
	hclen byte
}

type block struct {
	btype byte
	final bool

	trees tree

	b  []byte // compressed
	zb []byte // uncompressed
}
