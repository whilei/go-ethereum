package distip

import (
	"fmt"
	"net"
	"testing"
)

func parseIP(s string) net.IP {
	ip := net.ParseIP(s)
	if ip == nil {
		panic("invalid " + s)
	}
	return ip
}

func checkContains(t *testing.T, fn func(net.IP) bool, inc, exc []string) {
	for _, s := range inc {
		if !fn(parseIP(s)) {
			t.Error("returned false for included address", s)
		}
	}
	for _, s := range exc {
		if fn(parseIP(s)) {
			t.Error("returned true for excluded address", s)
		}
	}
}

func TestDistinctNetSet(t *testing.T) {
	ops := []struct {
		add, remove string
		fails       bool
	}{
		{add: "127.0.0.1"},
		{add: "127.0.0.2"},
		{add: "127.0.0.3", fails: true},
		{add: "127.32.0.1"},
		{add: "127.32.0.2"},
		{add: "127.32.0.3", fails: true},
		{add: "127.33.0.1", fails: true},
		{add: "127.34.0.1"},
		{add: "127.34.0.2"},
		{add: "127.34.0.3", fails: true},
		// Make room for an address, then add again.
		{remove: "127.0.0.1"},
		{add: "127.0.0.3"},
		{add: "127.0.0.3", fails: true},
	}

	set := DistinctNetSet{Subnet: 15, Limit: 2}
	for _, op := range ops {
		var desc string
		if op.add != "" {
			desc = fmt.Sprintf("Add(%s)", op.add)
			if ok := set.Add(parseIP(op.add)); ok != !op.fails {
				t.Errorf("%s == %t, want %t", desc, ok, !op.fails)
			}
		} else {
			desc = fmt.Sprintf("Remove(%s)", op.remove)
			set.Remove(parseIP(op.remove))
		}
		t.Logf("%s: %v", desc, set)
	}
}

func TestIsLAN(t *testing.T) {
	checkContains(t, IsLAN,
		[]string{ // included
			"0.0.0.0",
			"0.2.0.8",
			"127.0.0.1",
			"10.0.1.1",
			"10.22.0.3",
			"172.31.252.251",
			"192.168.1.4",
			"fe80::f4a1:8eff:fec5:9d9d",
			"febf::ab32:2233",
			"fc00::4",
		},
		[]string{ // excluded
			"192.0.2.1",
			"1.0.0.0",
			"172.32.0.1",
			"fec0::2233",
		},
	)
}
