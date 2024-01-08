package jqfmt

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/itchyny/gojq"
)

type JqFmtCfg struct {
	// Funcs []string
	Ops []string
	Obj bool
	Arr bool

	Hang  bool
	OneLn bool
}

var cfg JqFmtCfg

var line int
var node string
var ancestor string
var idt int
var nodeIdts map[string][]string
var queries map[string]int
var indented map[int]int

// var funcs []string
// var funcDefs map[string]string
// var modFuncs []string
// var modFuncDefs map[string]string
// var progFuncs []string
// var progFuncDefs map[string]string
var lastIdt int

func ValidateConfig(cfg JqFmtCfg) (JqFmtCfg, error) {
	validOps := []string{
		"pipe",
		"comma",
		"add",
		"sub",
		"mul",
		"div",
		"mod",
		"eq",
		"ne",
		"gt",
		"lt",
		"ge",
		"le",
		"and",
		"or",
		"alt",
		"assign",
		"modify",
		"updateAdd",
		"updateSub",
		"updateMul",
		"updateDiv",
		"updateMod",
		"updateAlt",
	}

	ops := cfg.Ops
	for o, op := range ops {
		valid := false
		for _, vop := range validOps {
			if strings.ToLower(op) == strings.ToLower(vop) {
				ops[o] = vop
				valid = true
			}
		}
		if !valid {
			return cfg, fmt.Errorf("invalid operator \"%s\"; valid operators: %s\n", op, strings.Join(validOps[:], ", "))
		}
	}
	cfg.Ops = ops

	return cfg, nil

}

func strToQuery(jqStr string) (Query, error) {

	jqAstQ := Query{}

	// Parse into AST.
	jqAst, err := gojq.Parse(jqStr)
	// TODO: print gojq pretty errors
	if err != nil {
		return jqAstQ, fmt.Errorf("could not parse jq: %w", err)
	}

	// Initially format jq to give us something consistent to start with.
	jqAstPty, err := gojq.Parse(jqAst.String())
	if err != nil {
		return jqAstQ, fmt.Errorf("could not parse jq: %w", err)
	}

	// Convert from gojq.Query to Query.
	jqAstJson, err := json.Marshal(jqAstPty)
	if err != nil {
		return jqAstQ, fmt.Errorf("could not convert query: %w", err)
	}
	json.Unmarshal([]byte(jqAstJson), &jqAstQ)

	return jqAstQ, nil
}

func DoThing(jqStr string, cfg_ JqFmtCfg) (string, error) {

	cfg = cfg_

	// if !cfg.Hang {
	// 	idt = 0
	// } else {
	// 	idt = 1
	// }
	idt = 0
	lastIdt = 0
	line = 1
	node = ""
	ancestor = ""
	nodeIdts = map[string][]string{}
	// funcs = []string{}
	// funcDefs = map[string]string{}
	queries = map[string]int{}
	indented = map[int]int{}

	// Read in ~/.jq.
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("could not get user: %w", err)
	}
	dir := usr.HomeDir
	file := filepath.Join(dir, ".jq")
	modJqBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("could not read file: %w", err)
	}
	modJqStr := string(modJqBytes)
	modQ, err := strToQuery(modJqStr)
	if err != nil {
		return "", fmt.Errorf("could not convert jq to query: %w", err)
	}
	_ = modQ.String()
	// modFuncs = funcs
	// modFuncDefs = funcDefs
	// funcs = []string{}
	// funcDefs = map[string]string{}

	initQ, err := strToQuery(jqStr)
	if err != nil {
		return "", fmt.Errorf("could not convert jq to query: %w", err)
	}

	// This'll populate funcs and funcDefs.
	temp := initQ.String()
	// progFuncs = funcs
	// progFuncDefs = funcDefs

	/*
		which funcs are used in the program?
		are any of those defined only in ~/.jq (i.e., not in prog)?
		prepend those
		which funcDefs did we prepend?
		which funcs do those use?
		are any not already included?
		if so, prepend those
		this is some kind of loop...
	*/

	// for _, pf := range progFuncs {
	// 	inp := false
	// 	inm := false
	// 	for fn := range progFuncDefs {
	// 		if fn == pf {
	// 			inp = true
	// 			break
	// 		}
	// 	}
	// 	for fn := range modFuncDefs {
	// 		if fn == pf {
	// 			inm = true
	// 			break
	// 		}
	// 	}
	// 	if !inp && inm {
	// 		temp = fmt.Sprintf("%s\n%s", modFuncDefs[pf], temp)
	// 	}
	// }

	fnlQ, err := strToQuery(temp)
	if err != nil {
		return "", fmt.Errorf("could not convert jq to query: %w", err)
	}

	fnl := fnlQ.String()

	return fnl, nil
	// fnlStr, err := indent(fnl, jqStr)
	// if err != nil {
	// 	return jqStr, fmt.Errorf("could not convert query: %w", err)
	// }
	// return fnlStr, nil
}

func indent(fnl string, jqStr string) (string, error) {

	// If the smallest indent is greater than 4 spaces (the intended minimum
	// indent), then bring it down to 4 by subtracting the difference.
	min := -1
	for _, jqLn := range strings.Split(fnl, "\n") {
		re, err := regexp.Compile("(^ +).*")
		if err != nil {
			return jqStr, fmt.Errorf("could not compile regex: %w", err)
		}
		n := re.FindStringSubmatch(jqLn)
		if len(n) > 1 {
			if min == -1 || len(n[1]) < min {
				min = len(n[1])
			}
		}
	}
	trunc := 0
	if min > 4 {
		trunc = min - 4
	}

	first := true
	out := ""
	for _, jqLn := range strings.Split(fnl, "\n") {

		// Separate out indent and line.
		re, err := regexp.Compile("^( *)(.*)")
		if err != nil {
			return jqStr, fmt.Errorf("could not compile regex: %w", err)
		}
		parts := re.FindStringSubmatch(jqLn)
		idtPart := parts[1]
		lnPart := parts[2]

		// Drop blank lines if they made their way in somehow.
		if lnPart == "" {
			continue
		}

		// if cfg.Hang && first {
		if first {

			// Leave the first line with no added indentation.
			out += fmt.Sprintf("%s\n", lnPart)
			first = false
		} else {

			// Indent all other lines.
			if len(idtPart) >= trunc {
				idtPart = idtPart[trunc:]
			}
			out += fmt.Sprintf("%s%s\n", idtPart, lnPart)
		}
	}

	return out[:len(out)-1], nil

}

// https://stedolan.github.io/jq/manual/#Modules
func loadModules() (map[string]*gojq.FuncDef, error) {

	funcs := map[string]*gojq.FuncDef{}

	usr, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("could not get user: %w", err)
	}
	dir := usr.HomeDir
	file := filepath.Join(dir, ".jq")
	jqBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}
	jqStr := string(jqBytes)

	jqAst, err := gojq.Parse(jqStr)
	if err != nil {
		return nil, fmt.Errorf("could not parse jq: %w", err)
	}

	for _, fd := range jqAst.FuncDefs {
		funcs[fd.Name] = fd
	}

	return funcs, nil
}
