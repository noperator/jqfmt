package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/noperator/jqfmt"
	log "github.com/sirupsen/logrus"
)

func assertErrorToNilf(message string, err error) {
	if err != nil {
		log.Fatalf(message, err)
	}
}

func main() {
	// funcsStr := flag.String("fn", "", "functions")
	opsStr := flag.String("op", "", "operators")
	obj := flag.Bool("ob", false, "objects")
	arr := flag.Bool("ar", false, "arrays")
	oneLn := flag.Bool("o", false, "one line")
	file := flag.String("f", "", "file")
	write := flag.Bool("w", false, "write result to source file instead of stdout")
	verbose := flag.Bool("v", false, "verbose")
	// noHang := flag.Bool("nh", false, "no hanging indent")
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "Usage: %s [options] [files...]\n\n", os.Args[0])
		fmt.Fprintln(out, "Formats jq files or stdin. With -w, writes changes back to files.")
		fmt.Fprintln(out, "If no files are specified, input is read from stdin and output is written to stdout.\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *verbose {
		log.SetLevel(log.DebugLevel)
	}

	args := flag.Args()
	if *file != "" && len(args) > 0 {
		log.Fatal("cannot use -f with positional files")
	}
	if *write && *file == "" && len(args) == 0 {
		log.Fatal("no files specified")
	}

	// var funcs []string
	// if *funcsStr == "" {
	// 	funcs = []string{}
	// } else {
	// 	funcs = strings.Split(*funcsStr, ",")
	// }

	var ops []string
	if *opsStr == "" {
		ops = []string{}
	} else {
		ops = strings.Split(*opsStr, ",")
	}

	cliCfg, err := jqfmt.ValidateConfig(jqfmt.JqFmtCfg{
		Arr: *arr,
		// Funcs: funcs,
		Obj:   *obj,
		OneLn: *oneLn,
		Ops:   ops,
		// Hang:  !(*noHang),
	})
	assertErrorToNilf("invalid config: %v", err)

	if *file == "" && len(args) == 0 {
		jqBytes, err := io.ReadAll(os.Stdin)
		assertErrorToNilf("could not read stdin: %v", err)
		jqStrFmt, err := jqfmt.DoThing(string(jqBytes), cliCfg)
		assertErrorToNilf("could not format jq: %v", err)
		fmt.Print(jqStrFmt)
		return
	}

	var files []string
	if *file != "" {
		files = []string{*file}
	} else {
		files = args
	}

	for _, path := range files {
		jqBytes, err := os.ReadFile(path)
		assertErrorToNilf("could not read file: %v", err)
		jqStrFmt, err := jqfmt.DoThing(string(jqBytes), cliCfg)
		assertErrorToNilf("could not format jq: %v", err)
		if *write {
			err = writeFileIfChanged(path, jqBytes, jqStrFmt)
			assertErrorToNilf("could not write file: %v", err)
		} else {
			fmt.Print(jqStrFmt)
		}
	}
}

func writeFileIfChanged(path string, original []byte, formatted string) error {
	if string(original) == formatted {
		return nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".jqfmt-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.WriteString(formatted); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Chmod(tmpName, info.Mode().Perm()); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}
