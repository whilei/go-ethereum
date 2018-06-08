package main

import (
	"flag"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"testing"
)

func TestAppConfigToml(t *testing.T) {

	// PTAL God damn, this shit is the worst. And I wrote it.
	makeTmpDataDir(t)
	defer rmTmpDataDir(t)

	fs := &flag.FlagSet{}
	mustParseSet := func(set *flag.FlagSet, flags []string) *flag.FlagSet {
		if e := set.Parse(flags); e != nil {
			t.Fatal(e)
		}
		return set
	}
	appBase := makeCLIApp()
	appBase.Writer = ioutil.Discard
	contextBase := cli.NewContext(appBase, mustParseSet(fs, []string{}), nil)

	// establish default baselines
	falseDefaultsBase := []cli.BoolFlag{FastSyncFlag, RPCEnabledFlag}
	for _, d := range falseDefaultsBase {
		if contextBase.GlobalBool(d.Name) {
			t.Errorf("fast sync enabled by default")
		}
	}
	intDefaultsBbase := []cli.IntFlag{CacheFlag, VerbosityFlag}
	intDefaultsBbaseExpect := []int{CacheFlag.Value, VerbosityFlag.Value}
	for i, d := range intDefaultsBbase {
		if got := contextBase.GlobalInt(d.Name); got != intDefaultsBbaseExpect[i] {
			t.Errorf("got: %v, want: %v", got, intDefaultsBbaseExpect[i])
		}
	}

	// compareContexts := func(ctxBase *cli.Context, ctxRef *cli.Context, getValFn interface{}, name string) (diff []string) {
	// 	switch getValFn.(type) {
	// 	case func(string) bool:
	// 		base, ref := ctxBase.GlobalBool(name), ctxRef.GlobalBool(name)
	// 		if base != ref  {
	// 			t.Errorf("mismatch: base: %v, ref: %v", base, ref)
	// 		}
	// 	}
	// }
	//
	//
	//

}
