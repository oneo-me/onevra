package components

import (
	"testing"

	"github.com/oneo-me/onevra/source/icons/lucide"
)

func TestFileIconKinds(t *testing.T) {
	for _, tc := range []struct {
		name string
		want fileKind
	}{
		{"main.go", fileKind{lucide.FileCode, fileGo}},
		{"source/view_test.go", fileKind{lucide.FileCheck, fileGo}},
		{`source\go.mod`, fileKind{lucide.Package, fileGo}},
		{"README.zh-CN.md", fileNames["readme.md"]},
		{"LICENSE.mygo", fileNames["license"]},
		{".env.local", fileNames[".env"]},
		{"App.TSX", fileKind{lucide.Atom, fileReact}},
		{"notes", fileKind{lucide.File, ""}},
		{"data.unknown", fileKind{lucide.File, ""}},
	} {
		got := fileKindOf(tc.name)
		if got.color != tc.want.color || got.icon() != tc.want.icon() {
			t.Errorf("fileKindOf(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}
