// This implementation for Patch is taken from go-get:
// https://github.com/go-git/go-git/blob/master/plumbing/object/patch.go

package diff

import (
	"bytes"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	fdiff "github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/diff"
	dmp "github.com/sergi/go-diff/diffmatchpatch"
)

// NewPatch - returns a new Patch intance
func NewPatch(filePatches []fdiff.FilePatch, message string) *Patch {
	return &Patch{
		filePatches: filePatches,
		message:     message,
	}
}

// CreateFilePatch - returns a new FilePatch for the given difference
func CreateFilePatch(wanted, got string) (fdiff.FilePatch, error) {
	diffs := diff.Do(wanted, got)
	var chunks []fdiff.Chunk
	for _, d := range diffs {

		var op fdiff.Operation
		switch d.Type {
		case dmp.DiffEqual:
			op = fdiff.Equal
		case dmp.DiffDelete:
			op = fdiff.Delete
		case dmp.DiffInsert:
			op = fdiff.Add
		}

		chunks = append(chunks, &textChunk{d.Text, op})
	}

	return &textFilePatch{
		chunks: chunks,
		from: object.ChangeEntry{
			Name: wanted,
		},
		to: object.ChangeEntry{
			Name: got,
		},
	}, nil
}

// textFilePatch is an implementation of fdiff.FilePatch interface
type textFilePatch struct {
	chunks   []fdiff.Chunk
	from, to object.ChangeEntry
}

func (tf *textFilePatch) Files() (from fdiff.File, to fdiff.File) {
	f := &changeEntryWrapper{tf.from}
	t := &changeEntryWrapper{tf.to}

	if !f.Empty() {
		from = f
	}

	if !t.Empty() {
		to = t
	}

	return
}

func (t *textFilePatch) IsBinary() bool {
	return len(t.chunks) == 0
}

func (t *textFilePatch) Chunks() []fdiff.Chunk {
	return t.chunks
}

// changeEntryWrapper is an implementation of fdiff.File interface
type changeEntryWrapper struct {
	ce object.ChangeEntry
}

func (f *changeEntryWrapper) Hash() plumbing.Hash {
	if !f.ce.TreeEntry.Mode.IsFile() {
		return plumbing.ZeroHash
	}

	return f.ce.TreeEntry.Hash
}

func (f *changeEntryWrapper) Mode() filemode.FileMode {
	return f.ce.TreeEntry.Mode
}
func (f *changeEntryWrapper) Path() string {
	if !f.ce.TreeEntry.Mode.IsFile() {
		return ""
	}

	return f.ce.Name
}

func (f *changeEntryWrapper) Empty() bool {
	return !f.ce.TreeEntry.Mode.IsFile()
}

type Patch struct {
	message     string
	filePatches []fdiff.FilePatch
}

func (t *Patch) FilePatches() []fdiff.FilePatch {
	return t.filePatches
}

func (t *Patch) Message() string {
	return t.message
}

func (t *Patch) Format() string {
	out := bytes.NewBuffer(nil)
	ue := fdiff.NewUnifiedEncoder(out, fdiff.DefaultContextLines)
	ue.Encode(t)
	return out.String()
}

type textChunk struct {
	content string
	op      fdiff.Operation
}

func (t *textChunk) Content() string {
	return t.content
}

func (t *textChunk) Type() fdiff.Operation {
	return t.op
}
