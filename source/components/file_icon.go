package components

import (
	"path"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/icons/lucide"
)

// fileKind is a file type's icon and color; a zero color inherits the text
// color.
type fileKind struct {
	icon  func() *ui.SVG
	color string
}

// File type colors follow each ecosystem's own, darkened where needed to
// read on both light and dark backgrounds.
const (
	fileGo     = "#00ADD8"
	fileReact  = "#149ECA"
	fileTS     = "#3178C6"
	fileJS     = "#D4A72C"
	fileGit    = "#F05032"
	fileDoc    = "#519ABA"
	fileNPM    = "#CB3837"
	fileImage  = "#A074C4"
	fileShell  = "#4EAA25"
	fileConfig = "#8D8D8D"
	fileStyle  = "#C6538C"
	fileRust   = "#CE7E4E"
	filePython = "#3776AB"
	fileSwift  = "#F05138"
)

// fileNames match whole names, case-insensitively, before extensions.
var fileNames = map[string]fileKind{
	".gitignore":        {lucide.GitBranch, fileGit},
	".gitattributes":    {lucide.GitBranch, fileGit},
	".gitmodules":       {lucide.GitBranch, fileGit},
	".gitkeep":          {lucide.GitBranch, fileGit},
	"license":           {lucide.Scale, fileJS},
	"licence":           {lucide.Scale, fileJS},
	"copying":           {lucide.Scale, fileJS},
	"readme.md":         {lucide.BookOpen, fileDoc},
	"changelog.md":      {lucide.FileClock, fileDoc},
	"go.mod":            {lucide.Package, fileGo},
	"go.sum":            {lucide.Lock, fileGo},
	"go.work":           {lucide.Package, fileGo},
	"go.work.sum":       {lucide.Lock, fileGo},
	"package.json":      {lucide.Package, fileNPM},
	"package-lock.json": {lucide.Lock, fileNPM},
	"pnpm-lock.yaml":    {lucide.Lock, fileConfig},
	"yarn.lock":         {lucide.Lock, fileConfig},
	"bun.lock":          {lucide.Lock, fileConfig},
	"cargo.toml":        {lucide.Package, fileRust},
	"cargo.lock":        {lucide.Lock, fileRust},
	"tsconfig.json":     {lucide.FileCog, fileTS},
	"dockerfile":        {lucide.Container, fileTS},
	"makefile":          {lucide.FileTerminal, fileConfig},
	".env":              {lucide.FileKey, fileConfig},
	".editorconfig":     {lucide.FileCog, fileConfig},
}

var fileExtensions = map[string]fileKind{
	".go":       {lucide.FileCode, fileGo},
	".jsx":      {lucide.Atom, fileReact},
	".tsx":      {lucide.Atom, fileReact},
	".ts":       {lucide.FileCode, fileTS},
	".mts":      {lucide.FileCode, fileTS},
	".cts":      {lucide.FileCode, fileTS},
	".js":       {lucide.FileCode, fileJS},
	".mjs":      {lucide.FileCode, fileJS},
	".cjs":      {lucide.FileCode, fileJS},
	".rs":       {lucide.FileCode, fileRust},
	".py":       {lucide.FileCode, filePython},
	".swift":    {lucide.FileCode, fileSwift},
	".md":       {lucide.FileText, fileDoc},
	".mdx":      {lucide.FileText, fileDoc},
	".markdown": {lucide.FileText, fileDoc},
	".txt":      {lucide.FileText, ""},
	".json":     {lucide.FileBraces, fileJS},
	".yaml":     {lucide.FileCog, fileConfig},
	".yml":      {lucide.FileCog, fileConfig},
	".toml":     {lucide.FileCog, fileConfig},
	".css":      {lucide.FileCode, fileStyle},
	".scss":     {lucide.FileCode, fileStyle},
	".html":     {lucide.FileCode, fileGit},
	".sh":       {lucide.FileTerminal, fileShell},
	".zsh":      {lucide.FileTerminal, fileShell},
	".bash":     {lucide.FileTerminal, fileShell},
	".ps1":      {lucide.FileTerminal, fileTS},
	".png":      {lucide.FileImage, fileImage},
	".jpg":      {lucide.FileImage, fileImage},
	".jpeg":     {lucide.FileImage, fileImage},
	".gif":      {lucide.FileImage, fileImage},
	".webp":     {lucide.FileImage, fileImage},
	".bmp":      {lucide.FileImage, fileImage},
	".ico":      {lucide.FileImage, fileImage},
	".icns":     {lucide.FileImage, fileImage},
	".svg":      {lucide.FileImage, fileImage},
	".lock":     {lucide.Lock, fileConfig},
	".zip":      {lucide.FileArchive, fileConfig},
	".gz":       {lucide.FileArchive, fileConfig},
	".tar":      {lucide.FileArchive, fileConfig},
}

// fileKindOf chooses a file's icon by its name, then its extension.
func fileKindOf(name string) fileKind {
	lower := strings.ToLower(path.Base(strings.ReplaceAll(name, `\`, "/")))
	if k, ok := fileNames[lower]; ok {
		return k
	}
	switch {
	case strings.HasPrefix(lower, ".env."):
		return fileNames[".env"]
	case strings.HasPrefix(lower, "license") || strings.HasPrefix(lower, "licence"):
		return fileNames["license"]
	case strings.HasPrefix(lower, "readme."):
		return fileNames["readme.md"]
	case strings.HasSuffix(lower, "_test.go"):
		return fileKind{lucide.FileCheck, fileGo}
	}
	if k, ok := fileExtensions[path.Ext(lower)]; ok {
		return k
	}
	return fileKind{lucide.File, ""}
}

// FileIcon creates the type icon of a file named name, as code editors' file
// trees show: a Lucide icon chosen by its whole name, such as go.mod or
// .gitignore, or else its extension, tinted in its ecosystem's color. name may
// be a path; only its last element counts. Unknown types show a plain file
// icon in the inherited text color.
//
// Like Icon, it inherits font size; use Size for explicit dimensions and
// TextColor to override the tint.
func FileIcon(c *ui.Context, name string) ui.Element {
	k := fileKindOf(name)
	e := Icon(c, k.icon)
	if k.color != "" {
		e.TextColor(ui.Hex(k.color))
	}
	return e
}
