package jqfmt

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

func TestFmt(t *testing.T) {
	fmt.Print("")
}

// func TestComma(t *testing.T) {
// 	cases := []struct {
// 		inFile  string
// 		outFile string
// 	}{
// 		{"testdata/comma-in.jq", "testdata/comma-out.jq"},
// 	}
//
// 	// cfg = JqFmtCfg{
// 	// 	Arr: true,
// 	// }
// 	cfg = JqFmtCfg{
// 		Ops: []string{"comma"},
// 	}
//
// 	for _, c := range cases {
// 		inBytes, err := ioutil.ReadFile(c.inFile)
// 		if err != nil {
// 			t.Fatalf("failed to open input file: %s", err)
// 		}
// 		in := string(inBytes)
//
// 		wantBytes, err := ioutil.ReadFile(c.outFile)
// 		if err != nil {
// 			t.Fatalf("failed to open want file: %s", err)
// 		}
// 		want := string(wantBytes)
//
// 		out, err := DoThing(in, cfg)
// 		if err != nil {
// 			t.Fatalf("could not do thing: %s", err)
// 		}
//
// 		if !reflect.DeepEqual(want, out) {
// 			t.Logf("want: %s", want)
// 			t.Logf("have: %s", out)
// 			t.Errorf("%s does not match %s", c.inFile, c.outFile)
// 		}
// 	}
// }

func TestArray(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/array-in.jq", "testdata/array-out.jq"},
	}

	cfg = JqFmtCfg{
		Arr: true,
	}

	for _, c := range cases {
		inBytes, err := ioutil.ReadFile(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}
		in := string(inBytes)

		wantBytes, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}
		want := string(wantBytes)

		out, err := DoThing(in, cfg)
		if err != nil {
			t.Fatalf("could not do thing: %s", err)
		}

		if !reflect.DeepEqual(want, out) {
			t.Logf("want: %s", want)
			t.Logf("have: %s", out)
			t.Errorf("%s does not match %s", c.inFile, c.outFile)
		}
	}
}

func TestObject(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/object-in.jq", "testdata/object-out.jq"},
	}

	cfg = JqFmtCfg{
		Obj: true,
	}

	for _, c := range cases {
		inBytes, err := ioutil.ReadFile(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}
		in := string(inBytes)

		wantBytes, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}
		want := string(wantBytes)

		out, err := DoThing(in, cfg)
		if err != nil {
			t.Fatalf("could not do thing: %s", err)
		}

		if !reflect.DeepEqual(want, out) {
			t.Logf("want: %s", want)
			t.Logf("have: %s", out)
			t.Errorf("%s does not match %s", c.inFile, c.outFile)
		}
	}
}

func TestOperator(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/operator-pipe-in.jq", "testdata/operator-pipe-out.jq"},
		{"testdata/operator-comma-in.jq", "testdata/operator-comma-out.jq"},
		{"testdata/operator-add-in.jq", "testdata/operator-add-out.jq"},
		{"testdata/operator-sub-in.jq", "testdata/operator-sub-out.jq"},
		{"testdata/operator-mul-in.jq", "testdata/operator-mul-out.jq"},
		{"testdata/operator-div-in.jq", "testdata/operator-div-out.jq"},
		{"testdata/operator-mod-in.jq", "testdata/operator-mod-out.jq"},
		{"testdata/operator-eq-in.jq", "testdata/operator-eq-out.jq"},
		{"testdata/operator-ne-in.jq", "testdata/operator-ne-out.jq"},
		{"testdata/operator-gt-in.jq", "testdata/operator-gt-out.jq"},
		{"testdata/operator-lt-in.jq", "testdata/operator-lt-out.jq"},
		{"testdata/operator-ge-in.jq", "testdata/operator-ge-out.jq"},
		{"testdata/operator-le-in.jq", "testdata/operator-le-out.jq"},
		{"testdata/operator-and-in.jq", "testdata/operator-and-out.jq"},
		{"testdata/operator-or-in.jq", "testdata/operator-or-out.jq"},
		{"testdata/operator-alt-in.jq", "testdata/operator-alt-out.jq"},
		{"testdata/operator-assign-in.jq", "testdata/operator-assign-out.jq"},
		{"testdata/operator-modify-in.jq", "testdata/operator-modify-out.jq"},
		{"testdata/operator-updateAdd-in.jq", "testdata/operator-updateAdd-out.jq"},
		{"testdata/operator-updateSub-in.jq", "testdata/operator-updateSub-out.jq"},
		{"testdata/operator-updateMul-in.jq", "testdata/operator-updateMul-out.jq"},
		{"testdata/operator-updateDiv-in.jq", "testdata/operator-updateDiv-out.jq"},
		{"testdata/operator-updateMod-in.jq", "testdata/operator-updateMod-out.jq"},
		{"testdata/operator-updateAlt-in.jq", "testdata/operator-updateAlt-out.jq"},
	}

	for _, c := range cases {

		op := strings.Split(c.inFile, "-")[1]

		cfg = JqFmtCfg{
			Ops: []string{op},
		}

		inBytes, err := ioutil.ReadFile(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}
		in := string(inBytes)

		wantBytes, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}
		want := string(wantBytes)

		out, err := DoThing(in, cfg)
		if err != nil {
			t.Fatalf("could not do thing: %s", err)
		}

		if !reflect.DeepEqual(want, out) {
			t.Logf("want: %s", want)
			t.Logf("have: %s", out)
			t.Errorf("%s does not match %s", c.inFile, c.outFile)
		}
	}
}

// func TestFunction(t *testing.T) {
// 	cases := []struct {
// 		inFile  string
// 		outFile string
// 	}{
// 		{"testdata/function-map-in.jq", "testdata/function-map-out.jq"},
// 	}
//
// 	for _, c := range cases {
//
// 		fn := strings.Split(c.inFile, "-")[1]
//
// 		cfg = JqFmtCfg{
// 			Funcs: []string{fn},
// 		}
//
// 		inBytes, err := ioutil.ReadFile(c.inFile)
// 		if err != nil {
// 			t.Fatalf("failed to open input file: %s", err)
// 		}
// 		in := string(inBytes)
//
// 		wantBytes, err := ioutil.ReadFile(c.outFile)
// 		if err != nil {
// 			t.Fatalf("failed to open want file: %s", err)
// 		}
// 		want := string(wantBytes)
//
// 		out, err := DoThing(in, cfg)
// 		if err != nil {
// 			t.Fatalf("could not do thing: %s", err)
// 		}
//
// 		if !reflect.DeepEqual(want, out) {
// 			t.Logf("want: %s", want)
// 			t.Logf("have: %s", out)
// 			t.Errorf("%s does not match %s", c.inFile, c.outFile)
// 		}
// 	}
// }

func TestMulti(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/multi-1-in.jq", "testdata/multi-1-out.jq"},
	}

	for _, c := range cases {

		// n := strings.Split(c.inFile, "-")[1]

		cfg = JqFmtCfg{
			Ops: []string{"pipe"},
			Arr: true,
		}

		inBytes, err := ioutil.ReadFile(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}
		in := string(inBytes)

		wantBytes, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}
		want := string(wantBytes)

		out, err := DoThing(in, cfg)
		if err != nil {
			t.Fatalf("could not do thing: %s", err)
		}
		// fmt.Printf("want: %x\n", []byte(want))
		// fmt.Printf("out:  %x\n", []byte(out))

		if !reflect.DeepEqual(want, out) {
			t.Logf("want: %d %s", len(want), want)
			t.Logf("have: %d %s", len(out), out)
			t.Errorf("%s does not match %s", c.inFile, c.outFile)
		}
	}
}

func TestTrailingWhitespace(t *testing.T) {
	cfg = JqFmtCfg{
		Arr: true,
		Obj: true,
		Ops: []string{
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
		},
	}

	inBytes, err := ioutil.ReadFile("testdata/trailing-space-in.jq")
	if err != nil {
		t.Fatalf("failed to open input file: %s", err)
	}
	in := string(inBytes)

	wantBytes, err := ioutil.ReadFile("testdata/trailing-space-out.jq")
	if err != nil {
		t.Fatalf("failed to open want file: %s", err)
	}
	want := string(wantBytes)

	out, err := DoThing(in, cfg)
	if err != nil {
		t.Fatalf("could not do thing: %s", err)
	}

	if !reflect.DeepEqual(want, out) {
		t.Logf("want: %s", want)
		t.Logf("have: %s", out)
		t.Errorf("testdata/trailing-space-in.jq does not match testdata/trailing-space-out.jq")
	}
}

func TestFuncDef(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/funcdef-in.jq", "testdata/funcdef-out.jq"},
	}

	cfg = JqFmtCfg{}

	for _, c := range cases {
		inBytes, err := ioutil.ReadFile(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}
		in := string(inBytes)

		wantBytes, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}
		want := string(wantBytes)

		out, err := DoThing(in, cfg)
		if err != nil {
			t.Fatalf("could not do thing: %s", err)
		}

		if !reflect.DeepEqual(want, out) {
			t.Logf("want: %s", want)
			t.Logf("have: %s", out)
			t.Errorf("%s does not match %s", c.inFile, c.outFile)
		}
	}
}

/*
func TestFunction(t *testing.T) {
	cases := []struct {
		inFile  string
		outFile string
	}{
		{"testdata/function-map-in.jq", "testdata/function-map-out.jq"},
	}

	for _, c := range cases {

		fn := strings.Split(c.inFile, "-")[1]

		cfg = JqFmtCfg{
			Funcs: []string{fn},
		}

		inBytes, err := ioutil.ReadFile(c.inFile)
		if err != nil {
			t.Fatalf("failed to open input file: %s", err)
		}
		in := string(inBytes)

		wantBytes, err := ioutil.ReadFile(c.outFile)
		if err != nil {
			t.Fatalf("failed to open want file: %s", err)
		}
		want := string(wantBytes)

		out, err := DoThing(in, cfg)
		if err != nil {
			t.Fatalf("could not do thing: %s", err)
		}

		if !reflect.DeepEqual(want, out) {
			t.Logf("want: %s", want)
			t.Logf("have: %s", out)
			t.Errorf("%s does not match %s", c.inFile, c.outFile)
		}
	}
}
*/
