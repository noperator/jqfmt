package jqfmt

import (
	"encoding/json"
	"fmt"
)

// var AssertErrorToNilf = assertErrorToNilf
// var GetCmdStdDir = getCmdStdDir
// var Indent = indent
// var ModProg = modProg
// var ParseProg = parseProg
// var TreeToStr = treeToStr
// var FmtProg = fmtProg
// var FmtJq = fmtJq
// var FmtSh = fmtSh
// var ImplodeSh = implodeSh
// var ExplodeSh = explodeSh

var Cfg = cfg

// var Cmds = cmds

// TODO: Move this to util or something.
func PrintJSON(obj interface{}) {
	bytes, _ := json.MarshalIndent(obj, "\t", "\t")
	fmt.Println(string(bytes))
}
