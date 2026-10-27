package postgres

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Acceptance criterion (#566): "nil and empty slice behaviour is documented
// and stable." Stability is pinned by the integration test; this pins that
// the public methods' godoc states both cases for EncryptionKey.
func TestUploadSessionMethods_DocumentEncryptionKeyNilVersusEmpty(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "upload_session.go", nil, parser.ParseComments)
	require.NoError(t, err)

	docs := map[string]string{}
	for _, d := range f.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok && fn.Doc != nil {
			docs[fn.Name.Name] = strings.ToLower(fn.Doc.Text())
		}
	}
	for _, method := range []string{"CreateUploadSession", "GetUploadSession"} {
		doc, ok := docs[method]
		require.True(t, ok, "%s must have a doc comment", method)
		assert.Contains(t, doc, "encryptionkey", "%s godoc must mention EncryptionKey", method)
		assert.Contains(t, doc, "nil", "%s godoc must describe nil EncryptionKey behaviour", method)
		assert.Contains(t, doc, "empty", "%s godoc must describe empty EncryptionKey behaviour", method)
	}
}
