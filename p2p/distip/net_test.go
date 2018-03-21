package distip

import (
	"net"
	"strconv"
	"testing"
)

const distinctNetSetLimit = 42
const testIpPrefix = "24.207.212."

func makeTestDistinctNetSet() *DistinctNetSet {
	return &DistinctNetSet{
		Subnet: 24,
		Limit:  distinctNetSetLimit,
	}
}

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
	set := makeTestDistinctNetSet()

	// 0 <= i <= 252
	makeTestIp := func(i int) net.IP {
		return net.ParseIP(testIpPrefix + strconv.Itoa(i))
	}

	var setIps []net.IP

	i := 1
	for i <= distinctNetSetLimit {
		testip := makeTestIp(i)
		key := set.key(testip)
		t.Logf("ip: %v, key: %v", testip, key)

		if ok := set.Add(testip); !ok {
			t.Errorf("got: %v, want: %v", ok, true)
		}
		if contains := set.Contains(testip); !contains {
			t.Errorf("got: %v, want: %v", contains, true)
		}
		if len := set.Len(); len != uint(i) {
			t.Errorf("got: %v, want: %v", len, i)
		}
		setIps = append(setIps, testip)
		i++
	}

	// Show that addition fails once above set limit, eg. i=43
	if i != distinctNetSetLimit+1 {
		t.Fatalf("i<=distinctNetSetLimit -> got: %v, want: %v", i, distinctNetSetLimit+1)
	}
	testip := makeTestIp(i) // i == distinctNetSetLimit + 1
	if ok := set.Add(testip); ok {
		t.Errorf("got: %v, want: %v", ok, false)
	}
	if contains := set.Contains(testip); contains {
		t.Errorf("got: %v, want: %v, testip: %v", contains, false, testip)
	}
	if len := set.Len(); len == uint(i) {
		t.Errorf("got: %v, want: %v", len, i-1)
	}
	i-- // set back to setlimit, eg. i=42

	for _, ip := range setIps {
		set.Remove(ip)
		i--
		if contains := set.Contains(ip); contains {
			t.Errorf("got: %v, want: %v, ip: %v", contains, false, ip)
		}
		if len := set.Len(); len != uint(i) {
			t.Errorf("got: %v, want: %v", len, i)
		}
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
