package distip

import (
	"net"
	"strconv"
	"testing"
)

const distinctNetSetLimit = 42

func makeTestDistinctNetSet() *DistinctNetSet {
	return &DistinctNetSet{
		Subnet: 24,
		Limit:  distinctNetSetLimit,
	}
}

func TestDistinctNetSet(t *testing.T) {
	set := makeTestDistinctNetSet()

	// 0 <= i <= 252
	makeTestIp := func(i int) net.IP {
		return net.ParseIP("24.207.212." + strconv.Itoa(i))
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

	// Show that addition fails once above set limit
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
	i-- // set back to setlimit

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
