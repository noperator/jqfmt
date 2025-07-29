package jqfmt

// TODO: Clean this up, pull only what's required from each file, and copypaste
// as much as possible without prepending "gojq." to various things.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	// "regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/itchyny/gojq"
	log "github.com/sirupsen/logrus"
)

// Misc
// ----------------------------------------

func toNumber(v string) interface{} {
	return normalizeNumber(json.Number(v))
}

func funcOpNegate(v interface{}) interface{} {
	switch v := v.(type) {
	case int:
		return -v
	case float64:
		return -v
	case *big.Int:
		return new(big.Int).Neg(v)
	default:
		return &unaryTypeError{"negate", v}
	}
}

type unaryTypeError struct {
	name string
	v    interface{}
}

// Query
// ----------------------------------------

// Query represents the abstract syntax tree of a jq query.
type Query struct {
	Meta     *ConstObject
	Imports  []*Import
	FuncDefs []*FuncDef
	Term     *Term
	Left     *Query
	Op       gojq.Operator
	Right    *Query
	Func     string
}

func (e *Query) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func nodeIdt(nodeToIdt, reason string) {
	log.Debugf("indenting \"%s\" for \"%s\" because \"%s\"\n", nodeToIdt, node, reason)
	nodeIdts[nodeToIdt] = append(nodeIdts[nodeToIdt], reason)
}

func prtIdt(s *strings.Builder) {
	if node == ".Identity" {
		return
	}
	if strings.HasSuffix(node, ".Query.Left.Func") {
		return
	}
	// idtHist := map[string]int{}
	if indented[line] == 0 {
		cIdt := 0
		for n, reason := range nodeIdts {
			if node == n {
				continue
			}
			if strings.HasPrefix(node, n) {
				log.Debugf("indt: \"%s\" -- %s\n", n, strings.Join(reason[:], ", "))
				cIdt += 1
			}
		}
		idtStr := ""
		log.Debugf("indt: %d\tnode: \"%s\"\n", cIdt, node)
		for i := 0; i < cIdt; i++ {
			idtStr += "    "
		}
		s.WriteString(idtStr)
		indented[line] = 1
	} else {
	}
}

func brk(s *strings.Builder) {

	s.WriteByte('\n')
	// idtStr := ""
	// for i := 0; i < idt; i++ {
	// 	// idtStr += "    "
	// }
	// s.WriteString(idtStr)
	line += 1
}

func descendsFrom(node string, ancestor string, parents []string) (bool, string) {
	nodeParts := strings.Split(node, ".")
	descendsFrom := true
	var n int
	for n = len(nodeParts) - 1; n >= 0; n-- {
		if nodeParts[n] == ancestor {
			break
		}
		parentValid := false
		for _, parent := range parents {
			if nodeParts[n] == parent {
				parentValid = true
			}
		}
		if !parentValid || n == 0 {
			descendsFrom = false
		}
	}
	return descendsFrom, strings.Join(nodeParts[:n+1], ".")
}

func (e *Query) writeTo(s *strings.Builder) {
	prevNode := node
	// if e.Term != nil && e.Term.String() == "." {
	if node == "" {
		log.Debugln("----------------------------------------")
	}
	log.Debugln("---")
	// log.Debugf("node: %q\n", node)
	// log.Debugln("nodeIdts:", nodeIdts)

	// PrintJSON(e)

	// Where are we in the syntax tree?
	arrElem, _ := descendsFrom(node, "Array", []string{"", "Left", "Right"})
	firstQueryTerm, firstQueryAncestor := descendsFrom(node, "Query", []string{"Left"}) // needs Left/Right?
	// topQueryTerm, topQueryAncestor := descendsFrom(node, "", []string{"", "Left", "Right"})
	topQueryTerm, topQueryAncestor := descendsFrom(node, "", []string{"", "Left", "Query"})
	// log.Debugln("top:", topQueryTerm, "\tfirst:", firstQueryTerm, "\tnode:", node)
	log.Debugf("top: %t\tfirst: %t \tnode: %q\n", topQueryTerm, firstQueryTerm, node)
	// log.Debugln("first:", firstQueryTerm)

	if e.Meta != nil {
		s.WriteString("module ")
		node += ".Meta"
		e.Meta.writeTo(s)
		node = prevNode
		s.WriteString(";\n")
	}
	for _, im := range e.Imports {
		node += ".Imports"
		im.writeTo(s)
		node = prevNode
	}
	for i, fd := range e.FuncDefs {
		// if _, ok := funcDefs[fd.Name]; !ok {
		// 	funcDefs[fd.Name] = fd.String()
		// }
		if i > 0 {
			s.WriteByte(' ')
		}
		node += ".FuncDefs"
		fd.writeTo(s)
		node = prevNode
	}
	if len(e.FuncDefs) > 0 {
		s.WriteByte(' ')
	}
	if e.Func != "" {
		s.WriteString(e.Func)
	} else if e.Term != nil {
		// if e.Term.Func != nil {
		// found := false
		// for _, fn := range funcs {
		// 	if fn == e.Term.Func.Name {
		// 		found = true
		// 	}
		// }
		// if !found {
		// 	funcs = append(funcs, e.Term.Func.Name)
		// }
		// }
		node += fmt.Sprintf(".%s", strings.Replace(e.Term.Type.GoString(), "gojq.TermType", "", 1))
		prtIdt(s)
		log.Debugf("term: %q\n", e.Term)
		e.Term.writeTo(s)
		node = prevNode
	} else if e.Right != nil {

		node += ".Left"
		e.Left.writeTo(s)
		node = prevNode

		if true {

			if e.Op == gojq.OpComma {
				s.WriteString(", ")
			} else {
				s.WriteByte(' ')
				s.WriteString(e.Op.String())
				s.WriteByte(' ')
			}

			// Break on comma-separated array elements.
			if cfg.Arr && arrElem && e.Op == gojq.OpComma {
				brk(s)
			}

			// if e.Op == gojq.OpComma {
			// 	log.Debugln("COMMA!")
			// }

			opStr := e.Op.GoString()

			for _, op := range cfg.Ops {

				if opStr == fmt.Sprintf("gojq.Op%s", strings.Title(op)) {

					// log.Debugln("HERE!")

					// Does this order matter?
					ancestor := ""
					if topQueryTerm {
						ancestor = topQueryAncestor
					}
					if firstQueryTerm {
						ancestor = firstQueryAncestor
					}

					// log.Debugf("ancestor: \"%s\"\n", ancestor)
					// if (firstQueryTerm || topQueryTerm) && queries[ancestor] == 0 {
					// if (firstQueryTerm || topQueryTerm) && queries[ancestor] == 0 {
					if firstQueryTerm || topQueryTerm {
						// Don't indent twice for a query at the beginning of
						// the command string.
						// match, err := regexp.MatchString("(.Left)+.Query(.Left)+", node)
						// match, err := regexp.MatchString("Left", node)
						// if err != nil {
						// 	panic(err)
						// }
						// if match {
						// 	// log.Debugln("HERE2!!!")
						// 	// break
						// }
						// queries[ancestor] = 1
						// log.Debugln("here")
						// nodeIdts[ancestor+".Right"] = "first query term"
						// nodeIdts[ancestor+".Right"] = append(nodeIdts[ancestor+".Right"], "first query term")
						nodeIdt(ancestor+".Right", "first query term")

					} // else {

					// nodeIdts[ancestor+".Left"] = fmt.Sprintf("%s operator", op)
					// nodeIdts[ancestor+".Left"] = append(nodeIdts[ancestor+".Left"], fmt.Sprintf("%s operator", op))
					if e.Op != gojq.OpPipe { // Put this check here because arrays were getting indented twice. This seems to fix it.
						nodeIdt(ancestor+".Left", fmt.Sprintf("%s operator", op))
					}
					// }
					// if e.Op == gojq.OpComma {
					// 	if !arrElem {
					// 		// nodeIdts[ancestor+".Left.Right"] = 1
					// 		nodeIdts[ancestor+".Left"] = 1
					// 	} else {
					// 		continue
					// 	}
					// }
					brk(s)
				}
			}
		}

		node += ".Right"
		e.Right.writeTo(s)
		node = prevNode
	}
}

func (e *Query) minify() {
	for _, e := range e.FuncDefs {
		e.Minify()
	}
	if e.Term != nil {
		if name := e.Term.toFunc(); name != "" {
			e.Term = nil
			e.Func = name
		} else {
			e.Term.minify()
		}
	} else if e.Right != nil {
		e.Left.minify()
		e.Right.minify()
	}
}

func (e *Query) toIndexKey() interface{} {
	if e.Term == nil {
		return nil
	}
	return e.Term.toIndexKey()
}

func (e *Query) toIndices(xs []interface{}) []interface{} {
	if e.Term == nil {
		return nil
	}
	return e.Term.toIndices(xs)
}

// Import ...
type Import struct {
	ImportPath  string
	ImportAlias string
	IncludePath string
	Meta        *ConstObject
}

func (e *Import) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Import) writeTo(s *strings.Builder) {
	if e.ImportPath != "" {
		s.WriteString("import ")
		jsonEncodeString(s, e.ImportPath)
		s.WriteString(" as ")
		s.WriteString(e.ImportAlias)
	} else {
		s.WriteString("include ")
		jsonEncodeString(s, e.IncludePath)
	}
	if e.Meta != nil {
		s.WriteByte(' ')
		e.Meta.writeTo(s)
	}
	s.WriteString(";\n")
}

// FuncDef ...
type FuncDef struct {
	Name string
	Args []string
	Body *Query
}

func (e *FuncDef) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *FuncDef) writeTo(s *strings.Builder) {
	s.WriteString("def ")
	s.WriteString(e.Name)
	if len(e.Args) > 0 {
		s.WriteByte('(')
		for i, e := range e.Args {
			if i > 0 {
				s.WriteString("; ")
			}
			s.WriteString(e)
		}
		s.WriteByte(')')
	}
	s.WriteString(": ")
	e.Body.writeTo(s)
	s.WriteByte(';')
}

// Minify ...
func (e *FuncDef) Minify() {
	e.Body.minify()
}

// Term ...
type Term struct {
	Type       gojq.TermType
	Index      *Index
	Func       *Func
	Object     *Object
	Array      *Array
	Number     string
	Unary      *Unary
	Format     string
	Str        *String
	If         *If
	Try        *Try
	Reduce     *Reduce
	Foreach    *Foreach
	Label      *Label
	Break      string
	Query      *Query
	SuffixList []*Suffix
}

func (e *Term) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Term) writeTo(s *strings.Builder) {
	switch e.Type {
	case gojq.TermTypeIdentity:
		s.WriteByte('.')
	case gojq.TermTypeRecurse:
		s.WriteString("..")
	case gojq.TermTypeNull:
		s.WriteString("null")
	case gojq.TermTypeTrue:
		s.WriteString("true")
	case gojq.TermTypeFalse:
		s.WriteString("false")
	case gojq.TermTypeIndex:
		e.Index.writeTo(s)
	case gojq.TermTypeFunc:
		e.Func.writeTo(s)
	case gojq.TermTypeObject:
		e.Object.writeTo(s)
	case gojq.TermTypeArray:
		e.Array.writeTo(s)
	case gojq.TermTypeNumber:
		s.WriteString(e.Number)
	case gojq.TermTypeUnary:
		e.Unary.writeTo(s)
	case gojq.TermTypeFormat:
		s.WriteString(e.Format)
		if e.Str != nil {
			s.WriteByte(' ')
			e.Str.writeTo(s)
		}
	case gojq.TermTypeString:
		e.Str.writeTo(s)
	case gojq.TermTypeIf:
		e.If.writeTo(s)
	case gojq.TermTypeTry:
		e.Try.writeTo(s)
	case gojq.TermTypeReduce:
		e.Reduce.writeTo(s)
	case gojq.TermTypeForeach:
		e.Foreach.writeTo(s)
	case gojq.TermTypeLabel:
		e.Label.writeTo(s)
	case gojq.TermTypeBreak:
		s.WriteString("break ")
		s.WriteString(e.Break)
	case gojq.TermTypeQuery:
		s.WriteByte('(')
		e.Query.writeTo(s)
		s.WriteByte(')')
	}
	for _, e := range e.SuffixList {
		e.writeTo(s)
	}
}

func (e *Term) minify() {
	switch e.Type {
	case gojq.TermTypeIndex:
		e.Index.minify()
	case gojq.TermTypeFunc:
		e.Func.minify()
	case gojq.TermTypeObject:
		e.Object.minify()
	case gojq.TermTypeArray:
		e.Array.minify()
	case gojq.TermTypeUnary:
		e.Unary.minify()
	case gojq.TermTypeFormat:
		if e.Str != nil {
			e.Str.minify()
		}
	case gojq.TermTypeString:
		e.Str.minify()
	case gojq.TermTypeIf:
		e.If.minify()
	case gojq.TermTypeTry:
		e.Try.minify()
	case gojq.TermTypeReduce:
		e.Reduce.minify()
	case gojq.TermTypeForeach:
		e.Foreach.minify()
	case gojq.TermTypeLabel:
		e.Label.minify()
	case gojq.TermTypeQuery:
		e.Query.minify()
	}
	for _, e := range e.SuffixList {
		e.minify()
	}
}

func (e *Term) toFunc() string {
	if len(e.SuffixList) != 0 {
		return ""
	}
	// ref: compiler#compileQuery
	switch e.Type {
	case gojq.TermTypeIdentity:
		return "."
	case gojq.TermTypeRecurse:
		return ".."
	case gojq.TermTypeNull:
		return "null"
	case gojq.TermTypeTrue:
		return "true"
	case gojq.TermTypeFalse:
		return "false"
	case gojq.TermTypeFunc:
		return e.Func.toFunc()
	default:
		return ""
	}
}

func (e *Term) toIndexKey() interface{} {
	switch e.Type {
	case gojq.TermTypeNumber:
		return toNumber(e.Number)
	case gojq.TermTypeUnary:
		return e.Unary.toNumber()
	case gojq.TermTypeString:
		if e.Str.Queries == nil {
			return e.Str.Str
		}
		return nil
	default:
		return nil
	}
}

func (e *Term) toIndices(xs []interface{}) []interface{} {
	switch e.Type {
	case gojq.TermTypeIndex:
		if xs = e.Index.toIndices(xs); xs == nil {
			return nil
		}
	case gojq.TermTypeQuery:
		if xs = e.Query.toIndices(xs); xs == nil {
			return nil
		}
	default:
		return nil
	}
	for _, s := range e.SuffixList {
		if xs = s.toIndices(xs); xs == nil {
			return nil
		}
	}
	return xs
}

func (e *Term) toNumber() interface{} {
	if e.Type == gojq.TermTypeNumber {
		return toNumber(e.Number)
	}
	return nil
}

// Unary ...
type Unary struct {
	Op   gojq.Operator
	Term *Term
}

func (e *Unary) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Unary) writeTo(s *strings.Builder) {
	s.WriteString(e.Op.String())
	e.Term.writeTo(s)
}

func (e *Unary) minify() {
	e.Term.minify()
}

func (e *Unary) toNumber() interface{} {
	v := e.Term.toNumber()
	if v != nil && e.Op == gojq.OpSub {
		v = funcOpNegate(v)
	}
	return v
}

// Pattern ...
type Pattern struct {
	Name   string
	Array  []*Pattern
	Object []*PatternObject
}

func (e *Pattern) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Pattern) writeTo(s *strings.Builder) {
	if e.Name != "" {
		s.WriteString(e.Name)
	} else if len(e.Array) > 0 {
		s.WriteByte('[')
		for i, e := range e.Array {
			if i > 0 {
				s.WriteString(", ")
			}
			e.writeTo(s)
		}
		s.WriteByte(']')
	} else if len(e.Object) > 0 {
		s.WriteByte('{')
		for i, e := range e.Object {
			if i > 0 {
				s.WriteString(", ")
			}
			e.writeTo(s)
		}
		s.WriteByte('}')
	}
}

// PatternObject ...
type PatternObject struct {
	Key       string
	KeyString *String
	KeyQuery  *Query
	Val       *Pattern
}

func (e *PatternObject) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *PatternObject) writeTo(s *strings.Builder) {
	if e.Key != "" {
		s.WriteString(e.Key)
	} else if e.KeyString != nil {
		e.KeyString.writeTo(s)
	} else if e.KeyQuery != nil {
		s.WriteByte('(')
		e.KeyQuery.writeTo(s)
		s.WriteByte(')')
	}
	if e.Val != nil {
		s.WriteString(": ")
		e.Val.writeTo(s)
	}
}

// Index ...
type Index struct {
	Name    string
	Str     *String
	Start   *Query
	End     *Query
	IsSlice bool
}

func (e *Index) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Index) writeTo(s *strings.Builder) {
	if l := s.Len(); l > 0 {
		// ". .x" != "..x" and "0 .x" != "0.x"
		if c := s.String()[l-1]; c == '.' || '0' <= c && c <= '9' {
			s.WriteByte(' ')
		}
	}
	s.WriteByte('.')
	e.writeSuffixTo(s)
}

func (e *Index) writeSuffixTo(s *strings.Builder) {
	if e.Name != "" {
		s.WriteString(e.Name)
	} else if e.Str != nil {
		e.Str.writeTo(s)
	} else {
		s.WriteByte('[')
		if e.IsSlice {
			if e.Start != nil {
				e.Start.writeTo(s)
			}
			s.WriteByte(':')
			if e.End != nil {
				e.End.writeTo(s)
			}
		} else {
			e.Start.writeTo(s)
		}
		s.WriteByte(']')
	}
}

func (e *Index) minify() {
	if e.Str != nil {
		e.Str.minify()
	}
	if e.Start != nil {
		e.Start.minify()
	}
	if e.End != nil {
		e.End.minify()
	}
}

func (e *Index) toIndexKey() interface{} {
	if e.Name != "" {
		return e.Name
	} else if e.Str != nil {
		if e.Str.Queries == nil {
			return e.Str.Str
		}
	} else if !e.IsSlice {
		return e.Start.toIndexKey()
	} else {
		var start, end interface{}
		ok := true
		if e.Start != nil {
			start = e.Start.toIndexKey()
			ok = start != nil
		}
		if e.End != nil && ok {
			end = e.End.toIndexKey()
			ok = end != nil
		}
		if ok {
			return map[string]interface{}{"start": start, "end": end}
		}
	}
	return nil
}

func (e *Index) toIndices(xs []interface{}) []interface{} {
	if k := e.toIndexKey(); k != nil {
		return append(xs, k)
	}
	return nil
}

// Func ...
type Func struct {
	Name string
	Args []*Query
}

func (e *Func) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Func) writeTo(s *strings.Builder) {
	// loadMod(e.Name)
	// fmt.Println(mods)
	// for _, f := range cfg.Funcs {
	// 	if e.Name == f {
	// 		brk(s)
	// 	}
	// }
	s.WriteString(e.Name)
	if len(e.Args) > 0 {
		s.WriteByte('(')
		for i, e := range e.Args {
			if i > 0 {
				s.WriteString("; ")
			}
			e.writeTo(s)
		}
		s.WriteByte(')')
	}
}

func (e *Func) minify() {
	for _, x := range e.Args {
		x.minify()
	}
}

func (e *Func) toFunc() string {
	if len(e.Args) != 0 {
		return ""
	}
	return e.Name
}

// String ...
type String struct {
	Str     string
	Queries []*Query
}

func (e *String) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *String) writeTo(s *strings.Builder) {
	if e.Queries == nil {
		jsonEncodeString(s, e.Str)
		return
	}
	s.WriteByte('"')
	for _, e := range e.Queries {
		if e.Term.Str == nil {
			s.WriteString(`\`)
			e.writeTo(s)
		} else {
			es := e.String()
			s.WriteString(es[1 : len(es)-1])
		}
	}
	s.WriteByte('"')
}

func (e *String) minify() {
	for _, e := range e.Queries {
		e.minify()
	}
}

// Object ...
type Object struct {
	KeyVals []*ObjectKeyVal
}

func (e *Object) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Object) writeTo(s *strings.Builder) {
	if len(e.KeyVals) == 0 {
		s.WriteString("{}")
		return
	}
	s.WriteString("{ ")
	if cfg.Obj {
		// nodeIdts[node] = "object"
		// nodeIdts[node] = append(nodeIdts[node], "object")
		nodeIdt(node, "object")
	}
	for i, kv := range e.KeyVals {
		if i > 0 {
			s.WriteString(", ")
		}
		if cfg.Obj {
			prevNode := node
			node += ".KeyVals"
			brk(s)
			prtIdt(s)
			node = prevNode
		}
		kv.writeTo(s)
	}
	if cfg.Obj {
		brk(s)
		prtIdt(s)
		s.WriteString("}")
	} else {
		s.WriteString(" }")
	}
}

func (e *Object) minify() {
	for _, e := range e.KeyVals {
		e.minify()
	}
}

// ObjectKeyVal ...
type ObjectKeyVal struct {
	Key       string
	KeyString *String
	KeyQuery  *Query
	Val       *ObjectVal
}

func (e *ObjectKeyVal) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *ObjectKeyVal) writeTo(s *strings.Builder) {
	if e.Key != "" {
		s.WriteString(e.Key)
	} else if e.KeyString != nil {
		e.KeyString.writeTo(s)
	} else if e.KeyQuery != nil {
		s.WriteByte('(')
		e.KeyQuery.writeTo(s)
		s.WriteByte(')')
	}
	if cfg.Obj {
	}
	if e.Val != nil {
		s.WriteString(": ")
		e.Val.writeTo(s)
	}
	if cfg.Obj {
	}
}

func (e *ObjectKeyVal) minify() {
	if e.KeyString != nil {
		e.KeyString.minify()
	} else if e.KeyQuery != nil {
		e.KeyQuery.minify()
	}
	if e.Val != nil {
		e.Val.minify()
	}
}

// ObjectVal ...
type ObjectVal struct {
	Queries []*Query
}

func (e *ObjectVal) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *ObjectVal) writeTo(s *strings.Builder) {
	for i, e := range e.Queries {
		if i > 0 {
			s.WriteString(" | ")
		}
		e.writeTo(s)
	}
}

func (e *ObjectVal) minify() {
	for _, e := range e.Queries {
		e.minify()
	}
}

// Array ...
type Array struct {
	Query *Query
}

func (e *Array) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Array) writeTo(s *strings.Builder) {

	prtIdt(s)
	s.WriteByte('[')
	if cfg.Arr {
		brk(s)
		// nodeIdts[node] = "array"
		// nodeIdts[node] = append(nodeIdts[node], "array")
		nodeIdt(node, "array")
	}
	if e.Query != nil {
		arrQ := e.Query
		arrQ.writeTo(s)
	}
	if cfg.Arr {
		brk(s)
	}
	prtIdt(s)
	s.WriteByte(']')
}

func (e *Array) minify() {
	if e.Query != nil {
		e.Query.minify()
	}
}

// Suffix ...
type Suffix struct {
	Index    *Index
	Iter     bool
	Optional bool
	Bind     *Bind
}

func (e *Suffix) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Suffix) writeTo(s *strings.Builder) {
	if e.Index != nil {
		if e.Index.Name != "" || e.Index.Str != nil {
			e.Index.writeTo(s)
		} else {
			e.Index.writeSuffixTo(s)
		}
	} else if e.Iter {
		s.WriteString("[]")
	} else if e.Optional {
		s.WriteByte('?')
	} else if e.Bind != nil {
		e.Bind.writeTo(s)
	}
}

func (e *Suffix) minify() {
	if e.Index != nil {
		e.Index.minify()
	} else if e.Bind != nil {
		e.Bind.minify()
	}
}

func (e *Suffix) toTerm() *Term {
	if e.Index != nil {
		return &Term{Type: gojq.TermTypeIndex, Index: e.Index}
	} else if e.Iter {
		return &Term{Type: gojq.TermTypeIdentity, SuffixList: []*Suffix{{Iter: true}}}
	} else {
		return nil
	}
}

func (e *Suffix) toIndices(xs []interface{}) []interface{} {
	if e.Index == nil {
		return nil
	}
	return e.Index.toIndices(xs)
}

// Bind ...
type Bind struct {
	Patterns []*Pattern
	Body     *Query
}

func (e *Bind) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Bind) writeTo(s *strings.Builder) {
	for i, p := range e.Patterns {
		if i == 0 {
			s.WriteString(" as ")
			p.writeTo(s)
			s.WriteByte(' ')
		} else {
			s.WriteString("?// ")
			p.writeTo(s)
			s.WriteByte(' ')
		}
	}
	s.WriteString("| ")
	e.Body.writeTo(s)
}

func (e *Bind) minify() {
	e.Body.minify()
}

// If ...
type If struct {
	Cond *Query
	Then *Query
	Elif []*IfElif
	Else *Query
}

func (e *If) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *If) writeTo(s *strings.Builder) {
	s.WriteString("if ")
	e.Cond.writeTo(s)
	s.WriteString(" then ")
	e.Then.writeTo(s)
	for _, e := range e.Elif {
		s.WriteByte(' ')
		e.writeTo(s)
	}
	if e.Else != nil {
		s.WriteString(" else ")
		e.Else.writeTo(s)
	}
	s.WriteString(" end")
}

func (e *If) minify() {
	e.Cond.minify()
	e.Then.minify()
	for _, x := range e.Elif {
		x.minify()
	}
	if e.Else != nil {
		e.Else.minify()
	}
}

// IfElif ...
type IfElif struct {
	Cond *Query
	Then *Query
}

func (e *IfElif) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *IfElif) writeTo(s *strings.Builder) {
	s.WriteString("elif ")
	e.Cond.writeTo(s)
	s.WriteString(" then ")
	e.Then.writeTo(s)
}

func (e *IfElif) minify() {
	e.Cond.minify()
	e.Then.minify()
}

// Try ...
type Try struct {
	Body  *Query
	Catch *Query
}

func (e *Try) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Try) writeTo(s *strings.Builder) {
	s.WriteString("try ")
	e.Body.writeTo(s)
	if e.Catch != nil {
		s.WriteString(" catch ")
		e.Catch.writeTo(s)
	}
}

func (e *Try) minify() {
	e.Body.minify()
	if e.Catch != nil {
		e.Catch.minify()
	}
}

// Reduce ...
type Reduce struct {
	Term    *Term
	Pattern *Pattern
	Start   *Query
	Update  *Query
}

func (e *Reduce) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Reduce) writeTo(s *strings.Builder) {
	s.WriteString("reduce ")
	e.Term.writeTo(s)
	s.WriteString(" as ")
	e.Pattern.writeTo(s)
	s.WriteString(" (")
	e.Start.writeTo(s)
	s.WriteString("; ")
	e.Update.writeTo(s)
	s.WriteByte(')')
}

func (e *Reduce) minify() {
	e.Term.minify()
	e.Start.minify()
	e.Update.minify()
}

// Foreach ...
type Foreach struct {
	Term    *Term
	Pattern *Pattern
	Start   *Query
	Update  *Query
	Extract *Query
}

func (e *Foreach) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Foreach) writeTo(s *strings.Builder) {
	s.WriteString("foreach ")
	e.Term.writeTo(s)
	s.WriteString(" as ")
	e.Pattern.writeTo(s)
	s.WriteString(" (")
	e.Start.writeTo(s)
	s.WriteString("; ")
	e.Update.writeTo(s)
	if e.Extract != nil {
		s.WriteString("; ")
		e.Extract.writeTo(s)
	}
	s.WriteByte(')')
}

func (e *Foreach) minify() {
	e.Term.minify()
	e.Start.minify()
	e.Update.minify()
	if e.Extract != nil {
		e.Extract.minify()
	}
}

// Label ...
type Label struct {
	Ident string
	Body  *Query
}

func (e *Label) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *Label) writeTo(s *strings.Builder) {
	s.WriteString("label ")
	s.WriteString(e.Ident)
	s.WriteString(" | ")
	e.Body.writeTo(s)
}

func (e *Label) minify() {
	e.Body.minify()
}

// ConstTerm ...
type ConstTerm struct {
	Object *ConstObject
	Array  *ConstArray
	Number string
	Str    string
	Null   bool
	True   bool
	False  bool
}

func (e *ConstTerm) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *ConstTerm) writeTo(s *strings.Builder) {
	if e.Object != nil {
		e.Object.writeTo(s)
	} else if e.Array != nil {
		e.Array.writeTo(s)
	} else if e.Number != "" {
		s.WriteString(e.Number)
	} else if e.Null {
		s.WriteString("null")
	} else if e.True {
		s.WriteString("true")
	} else if e.False {
		s.WriteString("false")
	} else {
		jsonEncodeString(s, e.Str)
	}
}

func (e *ConstTerm) toValue() interface{} {
	if e.Object != nil {
		return e.Object.ToValue()
	} else if e.Array != nil {
		return e.Array.toValue()
	} else if e.Number != "" {
		return toNumber(e.Number)
	} else if e.Null {
		return nil
	} else if e.True {
		return true
	} else if e.False {
		return false
	} else {
		return e.Str
	}
}

// ConstObject ...
type ConstObject struct {
	KeyVals []*ConstObjectKeyVal
}

func (e *ConstObject) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *ConstObject) writeTo(s *strings.Builder) {
	if len(e.KeyVals) == 0 {
		s.WriteString("{}")
		return
	}
	s.WriteString("{ ")
	for i, kv := range e.KeyVals {
		if i > 0 {
			s.WriteString(", ")
		}
		kv.writeTo(s)
	}
	s.WriteString(" }")
}

// ToValue converts the object to map[string]interface{}.
func (e *ConstObject) ToValue() map[string]interface{} {
	if e == nil {
		return nil
	}
	v := make(map[string]interface{}, len(e.KeyVals))
	for _, e := range e.KeyVals {
		key := e.Key
		if key == "" {
			key = e.KeyString
		}
		v[key] = e.Val.toValue()
	}
	return v
}

// ConstObjectKeyVal ...
type ConstObjectKeyVal struct {
	Key       string
	KeyString string
	Val       *ConstTerm
}

func (e *ConstObjectKeyVal) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *ConstObjectKeyVal) writeTo(s *strings.Builder) {
	if e.Key != "" {
		s.WriteString(e.Key)
	} else {
		s.WriteString(e.KeyString)
	}
	s.WriteString(": ")
	e.Val.writeTo(s)
}

// ConstArray ...
type ConstArray struct {
	Elems []*ConstTerm
}

func (e *ConstArray) String() string {
	var s strings.Builder
	e.writeTo(&s)
	return s.String()
}

func (e *ConstArray) writeTo(s *strings.Builder) {
	s.WriteByte('[')
	for i, e := range e.Elems {
		if i > 0 {
			s.WriteString(", ")
		}
		e.writeTo(s)
	}
	s.WriteByte(']')
}

func (e *ConstArray) toValue() []interface{} {
	v := make([]interface{}, len(e.Elems))
	for i, e := range e.Elems {
		v[i] = e.toValue()
	}
	return v
}

// Encoder
// ----------------------------------------

// Marshal returns the jq-flavored JSON encoding of v.
//
// This method accepts only limited types (nil, bool, int, float64, *big.Int,
// string, []interface{}, and map[string]interface{}) because these are the
// possible types a gojq iterator can emit. This method marshals NaN to null,
// truncates infinities to (+|-) math.MaxFloat64, uses \b and \f in strings,
// and does not escape '<', '>', '&', '\u2028', and '\u2029'. These behaviors
// are based on the marshaler of jq command, and different from json.Marshal in
// the Go standard library. Note that the result is not safe to embed in HTML.
func Marshal(v interface{}) ([]byte, error) {
	var b bytes.Buffer
	(&encoder{w: &b}).encode(v)
	return b.Bytes(), nil
}

func jsonMarshal(v interface{}) string {
	var sb strings.Builder
	(&encoder{w: &sb}).encode(v)
	return sb.String()
}

func jsonEncodeString(sb *strings.Builder, v string) {
	(&encoder{w: sb}).encodeString(v)
}

type encoder struct {
	w interface {
		io.Writer
		io.ByteWriter
		io.StringWriter
	}
	buf [64]byte
}

func (e *encoder) encode(v interface{}) {
	switch v := v.(type) {
	case nil:
		e.w.WriteString("null")
	case bool:
		if v {
			e.w.WriteString("true")
		} else {
			e.w.WriteString("false")
		}
	case int:
		e.w.Write(strconv.AppendInt(e.buf[:0], int64(v), 10))
	case float64:
		e.encodeFloat64(v)
	case *big.Int:
		e.w.Write(v.Append(e.buf[:0], 10))
	case string:
		e.encodeString(v)
	case []interface{}:
		e.encodeArray(v)
	case map[string]interface{}:
		e.encodeMap(v)
	default:
		panic(fmt.Sprintf("invalid type: %[1]T (%[1]v)", v))
	}
}

// ref: floatEncoder in encoding/json
func (e *encoder) encodeFloat64(f float64) {
	if math.IsNaN(f) {
		e.w.WriteString("null")
		return
	}
	if f >= math.MaxFloat64 {
		f = math.MaxFloat64
	} else if f <= -math.MaxFloat64 {
		f = -math.MaxFloat64
	}
	fmt := byte('f')
	if x := math.Abs(f); x != 0 && x < 1e-6 || x >= 1e21 {
		fmt = 'e'
	}
	buf := strconv.AppendFloat(e.buf[:0], f, fmt, -1, 64)
	if fmt == 'e' {
		// clean up e-09 to e-9
		if n := len(buf); n >= 4 && buf[n-4] == 'e' && buf[n-3] == '-' && buf[n-2] == '0' {
			buf[n-2] = buf[n-1]
			buf = buf[:n-1]
		}
	}
	e.w.Write(buf)
}

// ref: encodeState#string in encoding/json
func (e *encoder) encodeString(s string) {
	e.w.WriteByte('"')
	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if ' ' <= b && b <= '~' && b != '"' && b != '\\' {
				i++
				continue
			}
			if start < i {
				e.w.WriteString(s[start:i])
			}
			switch b {
			case '"':
				e.w.WriteString(`\"`)
			case '\\':
				e.w.WriteString(`\\`)
			case '\b':
				e.w.WriteString(`\b`)
			case '\f':
				e.w.WriteString(`\f`)
			case '\n':
				e.w.WriteString(`\n`)
			case '\r':
				e.w.WriteString(`\r`)
			case '\t':
				e.w.WriteString(`\t`)
			default:
				const hex = "0123456789abcdef"
				e.w.WriteString(`\u00`)
				e.w.WriteByte(hex[b>>4])
				e.w.WriteByte(hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				e.w.WriteString(s[start:i])
			}
			e.w.WriteString(`\ufffd`)
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		e.w.WriteString(s[start:])
	}
	e.w.WriteByte('"')
}

func (e *encoder) encodeArray(vs []interface{}) {
	e.w.WriteByte('[')
	for i, v := range vs {
		if i > 0 {
			e.w.WriteByte(',')
		}
		e.encode(v)
	}
	e.w.WriteByte(']')
}

func (e *encoder) encodeMap(vs map[string]interface{}) {
	e.w.WriteByte('{')
	type keyVal struct {
		key string
		val interface{}
	}
	kvs := make([]keyVal, len(vs))
	var i int
	for k, v := range vs {
		kvs[i] = keyVal{k, v}
		i++
	}
	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].key < kvs[j].key
	})
	for i, kv := range kvs {
		if i > 0 {
			e.w.WriteByte(',')
		}
		e.encodeString(kv.key)
		e.w.WriteByte(':')
		e.encode(kv.val)
	}
	e.w.WriteByte('}')
}

// Normalize
// ----------------------------------------

func normalizeNumber(v json.Number) interface{} {
	if i, err := v.Int64(); err == nil && math.MinInt <= i && i <= math.MaxInt {
		return int(i)
	}
	if strings.ContainsAny(v.String(), ".eE") {
		if f, err := v.Float64(); err == nil {
			return f
		}
	}
	if bi, ok := new(big.Int).SetString(v.String(), 10); ok {
		return bi
	}
	if strings.HasPrefix(v.String(), "-") {
		return math.Inf(-1)
	}
	return math.Inf(1)
}

func normalizeNumbers(v interface{}) interface{} {
	switch v := v.(type) {
	case json.Number:
		return normalizeNumber(v)
	case *big.Int:
		if v.IsInt64() {
			if i := v.Int64(); math.MinInt <= i && i <= math.MaxInt {
				return int(i)
			}
		}
		return v
	case int64:
		if math.MinInt <= v && v <= math.MaxInt {
			return int(v)
		}
		return big.NewInt(v)
	case int32:
		return int(v)
	case int16:
		return int(v)
	case int8:
		return int(v)
	case uint:
		if v <= math.MaxInt {
			return int(v)
		}
		return new(big.Int).SetUint64(uint64(v))
	case uint64:
		if v <= math.MaxInt {
			return int(v)
		}
		return new(big.Int).SetUint64(v)
	case uint32:
		if uint64(v) <= math.MaxInt {
			return int(v)
		}
		return new(big.Int).SetUint64(uint64(v))
	case uint16:
		return int(v)
	case uint8:
		return int(v)
	case float32:
		return float64(v)
	case []interface{}:
		for i, x := range v {
			v[i] = normalizeNumbers(x)
		}
		return v
	case map[string]interface{}:
		for k, x := range v {
			v[k] = normalizeNumbers(x)
		}
		return v
	default:
		return v
	}
}
