// Code generated by golang.org/x/tools/cmd/bundle. DO NOT EDIT.
//
// This file was generated with the following command and then edited by hand to add
// a build constraint and modify gen_header:
//   bundle -o unicode_generate_gen.go -pkg main golang.org/x/text/internal/gen

// +build ignore

// Package gen contains common code for the various code generation tools in the
// text repository. Its usage ensures consistency between tools.
//
// This package defines command line flags that are common to most generation
// tools. The flags allow for specifying specific Unicode and CLDR versions
// in the public Unicode data repository (https://www.unicode.org/Public).
//
// A local Unicode data mirror can be set through the flag -local or the
// environment variable UNICODE_DIR. The former takes precedence. The local
// directory should follow the same structure as the public repository.
//
// IANA data can also optionally be mirrored by putting it in the iana directory
// rooted at the top of the local mirror. Beware, though, that IANA data is not
// versioned. So it is up to the developer to use the right version.
//

package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"go/build"
	"go/format"
	"hash"
	"hash/fnv"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/cldr"
)

// This file contains utilities for generating code.

// TODO: other write methods like:
// - slices, maps, types, etc.

// CodeWriter is a utility for writing structured code. It computes the content
// hash and size of written content. It ensures there are newlines between
// written code blocks.
type gen_CodeWriter struct {
	buf  bytes.Buffer
	Size int
	Hash hash.Hash32 // content hash
	gob  *gob.Encoder
	// For comments we skip the usual one-line separator if they are followed by
	// a code block.
	skipSep bool
}

func (w *gen_CodeWriter) Write(p []byte) (n int, err error) {
	return w.buf.Write(p)
}

// NewCodeWriter returns a new CodeWriter.
func gen_NewCodeWriter() *gen_CodeWriter {
	h := fnv.New32()
	return &gen_CodeWriter{Hash: h, gob: gob.NewEncoder(h)}
}

// WriteGoFile appends the buffer with the total size of all created structures
// and writes it as a Go file to the given file with the given package name.
func (w *gen_CodeWriter) WriteGoFile(filename, pkg string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Could not create file %s: %v", filename, err)
	}
	defer f.Close()
	if _, err = w.WriteGo(f, pkg, ""); err != nil {
		log.Fatalf("Error writing file %s: %v", filename, err)
	}
}

// WriteVersionedGoFile appends the buffer with the total size of all created
// structures and writes it as a Go file to the given file with the given
// package name and build tags for the current Unicode version,
func (w *gen_CodeWriter) WriteVersionedGoFile(filename, pkg string) {
	tags := gen_buildTags()
	if tags != "" {
		pattern := gen_fileToPattern(filename)
		gen_updateBuildTags(pattern)
		filename = fmt.Sprintf(pattern, gen_UnicodeVersion())
	}
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Could not create file %s: %v", filename, err)
	}
	defer f.Close()
	if _, err = w.WriteGo(f, pkg, tags); err != nil {
		log.Fatalf("Error writing file %s: %v", filename, err)
	}
}

// WriteGo appends the buffer with the total size of all created structures and
// writes it as a Go file to the given writer with the given package name.
func (w *gen_CodeWriter) WriteGo(out io.Writer, pkg, tags string) (n int, err error) {
	sz := w.Size
	if sz > 0 {
		w.WriteComment("Total table size %d bytes (%dKiB); checksum: %X\n", sz, sz/1024, w.Hash.Sum32())
	}
	defer w.buf.Reset()
	return gen_WriteGo(out, pkg, tags, w.buf.Bytes())
}

func (w *gen_CodeWriter) printf(f string, x ...interface{}) {
	fmt.Fprintf(w, f, x...)
}

func (w *gen_CodeWriter) insertSep() {
	if w.skipSep {
		w.skipSep = false
		return
	}
	// Use at least two newlines to ensure a blank space between the previous
	// block. WriteGoFile will remove extraneous newlines.
	w.printf("\n\n")
}

// WriteComment writes a comment block. All line starts are prefixed with "//".
// Initial empty lines are gobbled. The indentation for the first line is
// stripped from consecutive lines.
func (w *gen_CodeWriter) WriteComment(comment string, args ...interface{}) {
	s := fmt.Sprintf(comment, args...)
	s = strings.Trim(s, "\n")

	// Use at least two newlines to ensure a blank space between the previous
	// block. WriteGoFile will remove extraneous newlines.
	w.printf("\n\n// ")
	w.skipSep = true

	// strip first indent level.
	sep := "\n"
	for ; len(s) > 0 && (s[0] == '\t' || s[0] == ' '); s = s[1:] {
		sep += s[:1]
	}

	strings.NewReplacer(sep, "\n// ", "\n", "\n// ").WriteString(w, s)

	w.printf("\n")
}

func (w *gen_CodeWriter) writeSizeInfo(size int) {
	w.printf("// Size: %d bytes\n", size)
}

// WriteConst writes a constant of the given name and value.
func (w *gen_CodeWriter) WriteConst(name string, x interface{}) {
	w.insertSep()
	v := reflect.ValueOf(x)

	switch v.Type().Kind() {
	case reflect.String:
		w.printf("const %s %s = ", name, gen_typeName(x))
		w.WriteString(v.String())
		w.printf("\n")
	default:
		w.printf("const %s = %#v\n", name, x)
	}
}

// WriteVar writes a variable of the given name and value.
func (w *gen_CodeWriter) WriteVar(name string, x interface{}) {
	w.insertSep()
	v := reflect.ValueOf(x)
	oldSize := w.Size
	sz := int(v.Type().Size())
	w.Size += sz

	switch v.Type().Kind() {
	case reflect.String:
		w.printf("var %s %s = ", name, gen_typeName(x))
		w.WriteString(v.String())
	case reflect.Struct:
		w.gob.Encode(x)
		fallthrough
	case reflect.Slice, reflect.Array:
		w.printf("var %s = ", name)
		w.writeValue(v)
		w.writeSizeInfo(w.Size - oldSize)
	default:
		w.printf("var %s %s = ", name, gen_typeName(x))
		w.gob.Encode(x)
		w.writeValue(v)
		w.writeSizeInfo(w.Size - oldSize)
	}
	w.printf("\n")
}

func (w *gen_CodeWriter) writeValue(v reflect.Value) {
	x := v.Interface()
	switch v.Kind() {
	case reflect.String:
		w.WriteString(v.String())
	case reflect.Array:
		// Don't double count: callers of WriteArray count on the size being
		// added, so we need to discount it here.
		w.Size -= int(v.Type().Size())
		w.writeSlice(x, true)
	case reflect.Slice:
		w.writeSlice(x, false)
	case reflect.Struct:
		w.printf("%s{\n", gen_typeName(v.Interface()))
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			w.printf("%s: ", t.Field(i).Name)
			w.writeValue(v.Field(i))
			w.printf(",\n")
		}
		w.printf("}")
	default:
		w.printf("%#v", x)
	}
}

// WriteString writes a string literal.
func (w *gen_CodeWriter) WriteString(s string) {
	io.WriteString(w.Hash, s) // content hash
	w.Size += len(s)

	const maxInline = 40
	if len(s) <= maxInline {
		w.printf("%q", s)
		return
	}

	// We will render the string as a multi-line string.
	const maxWidth = 80 - 4 - len(`"`) - len(`" +`)

	// When starting on its own line, go fmt indents line 2+ an extra level.
	n, max := maxWidth, maxWidth-4

	// As per https://golang.org/issue/18078, the compiler has trouble
	// compiling the concatenation of many strings, s0 + s1 + s2 + ... + sN,
	// for large N. We insert redundant, explicit parentheses to work around
	// that, lowering the N at any given step: (s0 + s1 + ... + s63) + (s64 +
	// ... + s127) + etc + (etc + ... + sN).
	explicitParens, extraComment := len(s) > 128*1024, ""
	if explicitParens {
		w.printf(`(`)
		extraComment = "; the redundant, explicit parens are for https://golang.org/issue/18078"
	}

	// Print "" +\n, if a string does not start on its own line.
	b := w.buf.Bytes()
	if p := len(bytes.TrimRight(b, " \t")); p > 0 && b[p-1] != '\n' {
		w.printf("\"\" + // Size: %d bytes%s\n", len(s), extraComment)
		n, max = maxWidth, maxWidth
	}

	w.printf(`"`)

	for sz, p, nLines := 0, 0, 0; p < len(s); {
		var r rune
		r, sz = utf8.DecodeRuneInString(s[p:])
		out := s[p : p+sz]
		chars := 1
		if !unicode.IsPrint(r) || r == utf8.RuneError || r == '"' {
			switch sz {
			case 1:
				out = fmt.Sprintf("\\x%02x", s[p])
			case 2, 3:
				out = fmt.Sprintf("\\u%04x", r)
			case 4:
				out = fmt.Sprintf("\\U%08x", r)
			}
			chars = len(out)
		} else if r == '\\' {
			out = "\\" + string(r)
			chars = 2
		}
		if n -= chars; n < 0 {
			nLines++
			if explicitParens && nLines&63 == 63 {
				w.printf("\") + (\"")
			}
			w.printf("\" +\n\"")
			n = max - len(out)
		}
		w.printf("%s", out)
		p += sz
	}
	w.printf(`"`)
	if explicitParens {
		w.printf(`)`)
	}
}

// WriteSlice writes a slice value.
func (w *gen_CodeWriter) WriteSlice(x interface{}) {
	w.writeSlice(x, false)
}

// WriteArray writes an array value.
func (w *gen_CodeWriter) WriteArray(x interface{}) {
	w.writeSlice(x, true)
}

func (w *gen_CodeWriter) writeSlice(x interface{}, isArray bool) {
	v := reflect.ValueOf(x)
	w.gob.Encode(v.Len())
	w.Size += v.Len() * int(v.Type().Elem().Size())
	name := gen_typeName(x)
	if isArray {
		name = fmt.Sprintf("[%d]%s", v.Len(), name[strings.Index(name, "]")+1:])
	}
	if isArray {
		w.printf("%s{\n", name)
	} else {
		w.printf("%s{ // %d elements\n", name, v.Len())
	}

	switch kind := v.Type().Elem().Kind(); kind {
	case reflect.String:
		for _, s := range x.([]string) {
			w.WriteString(s)
			w.printf(",\n")
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// nLine and nBlock are the number of elements per line and block.
		nLine, nBlock, format := 8, 64, "%d,"
		switch kind {
		case reflect.Uint8:
			format = "%#02x,"
		case reflect.Uint16:
			format = "%#04x,"
		case reflect.Uint32:
			nLine, nBlock, format = 4, 32, "%#08x,"
		case reflect.Uint, reflect.Uint64:
			nLine, nBlock, format = 4, 32, "%#016x,"
		case reflect.Int8:
			nLine = 16
		}
		n := nLine
		for i := 0; i < v.Len(); i++ {
			if i%nBlock == 0 && v.Len() > nBlock {
				w.printf("// Entry %X - %X\n", i, i+nBlock-1)
			}
			x := v.Index(i).Interface()
			w.gob.Encode(x)
			w.printf(format, x)
			if n--; n == 0 {
				n = nLine
				w.printf("\n")
			}
		}
		w.printf("\n")
	case reflect.Struct:
		zero := reflect.Zero(v.Type().Elem()).Interface()
		for i := 0; i < v.Len(); i++ {
			x := v.Index(i).Interface()
			w.gob.EncodeValue(v)
			if !reflect.DeepEqual(zero, x) {
				line := fmt.Sprintf("%#v,\n", x)
				line = line[strings.IndexByte(line, '{'):]
				w.printf("%d: ", i)
				w.printf(line)
			}
		}
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			w.printf("%d: %#v,\n", i, v.Index(i).Interface())
		}
	default:
		panic("gen: slice elem type not supported")
	}
	w.printf("}")
}

// WriteType writes a definition of the type of the given value and returns the
// type name.
func (w *gen_CodeWriter) WriteType(x interface{}) string {
	t := reflect.TypeOf(x)
	w.printf("type %s struct {\n", t.Name())
	for i := 0; i < t.NumField(); i++ {
		w.printf("\t%s %s\n", t.Field(i).Name, t.Field(i).Type)
	}
	w.printf("}\n")
	return t.Name()
}

// typeName returns the name of the go type of x.
func gen_typeName(x interface{}) string {
	t := reflect.ValueOf(x).Type()
	return strings.Replace(fmt.Sprint(t), "main.", "", 1)
}

var (
	gen_url = flag.String("url",
		"https://www.unicode.org/Public",
		"URL of Unicode database directory")
	gen_iana = flag.String("iana",
		"http://www.iana.org",
		"URL of the IANA repository")
	gen_unicodeVersion = flag.String("unicode",
		gen_getEnv("UNICODE_VERSION", unicode.Version),
		"unicode version to use")
	gen_cldrVersion = flag.String("cldr",
		gen_getEnv("CLDR_VERSION", cldr.Version),
		"cldr version to use")
)

func gen_getEnv(name, def string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return def
}

// Init performs common initialization for a gen command. It parses the flags
// and sets up the standard logging parameters.
func gen_Init() {
	log.SetPrefix("")
	log.SetFlags(log.Lshortfile)
	flag.Parse()
}

const gen_header = `// Code generated by running "go generate". DO NOT EDIT.

`

// UnicodeVersion reports the requested Unicode version.
func gen_UnicodeVersion() string {
	return *gen_unicodeVersion
}

// CLDRVersion reports the requested CLDR version.
func gen_CLDRVersion() string {
	return *gen_cldrVersion
}

var gen_tags = []struct{ version, buildTags string }{
	{"9.0.0", "!go1.10"},
	{"10.0.0", "go1.10,!go1.13"},
	{"11.0.0", "go1.13"},
}

// buildTags reports the build tags used for the current Unicode version.
func gen_buildTags() string {
	v := gen_UnicodeVersion()
	for _, e := range gen_tags {
		if e.version == v {
			return e.buildTags
		}
	}
	log.Fatalf("Unknown build tags for Unicode version %q.", v)
	return ""
}

// IsLocal reports whether data files are available locally.
func gen_IsLocal() bool {
	dir, err := gen_localReadmeFile()
	if err != nil {
		return false
	}
	if _, err = os.Stat(dir); err != nil {
		return false
	}
	return true
}

// OpenUCDFile opens the requested UCD file. The file is specified relative to
// the public Unicode root directory. It will call log.Fatal if there are any
// errors.
func gen_OpenUCDFile(file string) io.ReadCloser {
	return gen_openUnicode(path.Join(*gen_unicodeVersion, "ucd", file))
}

// OpenCLDRCoreZip opens the CLDR core zip file. It will call log.Fatal if there
// are any errors.
func gen_OpenCLDRCoreZip() io.ReadCloser {
	return gen_OpenUnicodeFile("cldr", *gen_cldrVersion, "core.zip")
}

// OpenUnicodeFile opens the requested file of the requested category from the
// root of the Unicode data archive. The file is specified relative to the
// public Unicode root directory. If version is "", it will use the default
// Unicode version. It will call log.Fatal if there are any errors.
func gen_OpenUnicodeFile(category, version, file string) io.ReadCloser {
	if version == "" {
		version = gen_UnicodeVersion()
	}
	return gen_openUnicode(path.Join(category, version, file))
}

// OpenIANAFile opens the requested IANA file. The file is specified relative
// to the IANA root, which is typically either http://www.iana.org or the
// iana directory in the local mirror. It will call log.Fatal if there are any
// errors.
func gen_OpenIANAFile(path string) io.ReadCloser {
	return gen_Open(*gen_iana, "iana", path)
}

var (
	gen_dirMutex sync.Mutex
	gen_localDir string
)

const gen_permissions = 0755

func gen_localReadmeFile() (string, error) {
	p, err := build.Import("golang.org/x/text", "", build.FindOnly)
	if err != nil {
		return "", fmt.Errorf("Could not locate package: %v", err)
	}
	return filepath.Join(p.Dir, "DATA", "README"), nil
}

func gen_getLocalDir() string {
	gen_dirMutex.Lock()
	defer gen_dirMutex.Unlock()

	readme, err := gen_localReadmeFile()
	if err != nil {
		log.Fatal(err)
	}
	dir := filepath.Dir(readme)
	if _, err := os.Stat(readme); err != nil {
		if err := os.MkdirAll(dir, gen_permissions); err != nil {
			log.Fatalf("Could not create directory: %v", err)
		}
		ioutil.WriteFile(readme, []byte(gen_readmeTxt), gen_permissions)
	}
	return dir
}

const gen_readmeTxt = `Generated by golang.org/x/text/internal/gen. DO NOT EDIT.

This directory contains downloaded files used to generate the various tables
in the golang.org/x/text subrepo.

Note that the language subtag repo (iana/assignments/language-subtag-registry)
and all other times in the iana subdirectory are not versioned and will need
to be periodically manually updated. The easiest way to do this is to remove
the entire iana directory. This is mostly of concern when updating the language
package.
`

// Open opens subdir/path if a local directory is specified and the file exists,
// where subdir is a directory relative to the local root, or fetches it from
// urlRoot/path otherwise. It will call log.Fatal if there are any errors.
func gen_Open(urlRoot, subdir, path string) io.ReadCloser {
	file := filepath.Join(gen_getLocalDir(), subdir, filepath.FromSlash(path))
	return gen_open(file, urlRoot, path)
}

func gen_openUnicode(path string) io.ReadCloser {
	file := filepath.Join(gen_getLocalDir(), filepath.FromSlash(path))
	return gen_open(file, *gen_url, path)
}

// TODO: automatically periodically update non-versioned files.

func gen_open(file, urlRoot, path string) io.ReadCloser {
	if f, err := os.Open(file); err == nil {
		return f
	}
	r := gen_get(urlRoot, path)
	defer r.Close()
	b, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatalf("Could not download file: %v", err)
	}
	os.MkdirAll(filepath.Dir(file), gen_permissions)
	if err := ioutil.WriteFile(file, b, gen_permissions); err != nil {
		log.Fatalf("Could not create file: %v", err)
	}
	return ioutil.NopCloser(bytes.NewReader(b))
}

func gen_get(root, path string) io.ReadCloser {
	url := root + "/" + path
	fmt.Printf("Fetching %s...", url)
	defer fmt.Println(" done.")
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("HTTP GET: %v", err)
	}
	if resp.StatusCode != 200 {
		log.Fatalf("Bad GET status for %q: %q", url, resp.Status)
	}
	return resp.Body
}

// TODO: use Write*Version in all applicable packages.

// WriteUnicodeVersion writes a constant for the Unicode version from which the
// tables are generated.
func gen_WriteUnicodeVersion(w io.Writer) {
	fmt.Fprintf(w, "// UnicodeVersion is the Unicode version from which the tables in this package are derived.\n")
	fmt.Fprintf(w, "const UnicodeVersion = %q\n\n", gen_UnicodeVersion())
}

// WriteCLDRVersion writes a constant for the CLDR version from which the
// tables are generated.
func gen_WriteCLDRVersion(w io.Writer) {
	fmt.Fprintf(w, "// CLDRVersion is the CLDR version from which the tables in this package are derived.\n")
	fmt.Fprintf(w, "const CLDRVersion = %q\n\n", gen_CLDRVersion())
}

// WriteGoFile prepends a standard file comment and package statement to the
// given bytes, applies gofmt, and writes them to a file with the given name.
// It will call log.Fatal if there are any errors.
func gen_WriteGoFile(filename, pkg string, b []byte) {
	w, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Could not create file %s: %v", filename, err)
	}
	defer w.Close()
	if _, err = gen_WriteGo(w, pkg, "", b); err != nil {
		log.Fatalf("Error writing file %s: %v", filename, err)
	}
}

func gen_fileToPattern(filename string) string {
	suffix := ".go"
	if strings.HasSuffix(filename, "_test.go") {
		suffix = "_test.go"
	}
	prefix := filename[:len(filename)-len(suffix)]
	return fmt.Sprint(prefix, "%s", suffix)
}

func gen_updateBuildTags(pattern string) {
	for _, t := range gen_tags {
		oldFile := fmt.Sprintf(pattern, t.version)
		b, err := ioutil.ReadFile(oldFile)
		if err != nil {
			continue
		}
		build := fmt.Sprintf("// +build %s", t.buildTags)
		b = regexp.MustCompile(`// \+build .*`).ReplaceAll(b, []byte(build))
		err = ioutil.WriteFile(oldFile, b, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// WriteVersionedGoFile prepends a standard file comment, adds build tags to
// version the file for the current Unicode version, and package statement to
// the given bytes, applies gofmt, and writes them to a file with the given
// name. It will call log.Fatal if there are any errors.
func gen_WriteVersionedGoFile(filename, pkg string, b []byte) {
	pattern := gen_fileToPattern(filename)
	gen_updateBuildTags(pattern)
	filename = fmt.Sprintf(pattern, gen_UnicodeVersion())

	w, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Could not create file %s: %v", filename, err)
	}
	defer w.Close()
	if _, err = gen_WriteGo(w, pkg, gen_buildTags(), b); err != nil {
		log.Fatalf("Error writing file %s: %v", filename, err)
	}
}

// WriteGo prepends a standard file comment and package statement to the given
// bytes, applies gofmt, and writes them to w.
func gen_WriteGo(w io.Writer, pkg, tags string, b []byte) (n int, err error) {
	src := []byte(gen_header)
	if tags != "" {
		src = append(src, fmt.Sprintf("// +build %s\n\n", tags)...)
	}
	src = append(src, fmt.Sprintf("package %s\n\n", pkg)...)
	src = append(src, b...)
	formatted, err := format.Source(src)
	if err != nil {
		// Print the generated code even in case of an error so that the
		// returned error can be meaningfully interpreted.
		n, _ = w.Write(src)
		return n, err
	}
	return w.Write(formatted)
}

// Repackage rewrites a Go file from belonging to package main to belonging to
// the given package.
func gen_Repackage(inFile, outFile, pkg string) {
	src, err := ioutil.ReadFile(inFile)
	if err != nil {
		log.Fatalf("reading %s: %v", inFile, err)
	}
	const toDelete = "package main\n\n"
	i := bytes.Index(src, []byte(toDelete))
	if i < 0 {
		log.Fatalf("Could not find %q in %s.", toDelete, inFile)
	}
	w := &bytes.Buffer{}
	w.Write(src[i+len(toDelete):])
	gen_WriteGoFile(outFile, pkg, w.Bytes())
}
