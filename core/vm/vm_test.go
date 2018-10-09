package vm

import (
	"math/big"
	"testing"
)

func TestEIP1283SStoreGas(t *testing.T) {
	cases := []struct {
		wantGas    *big.Int
		wantRefund *big.Int
		originalV  *big.Int
		iters      [3]int64 // potential for 3 possible iterations, -1 will be used when only 2 iterations wanted
	}{
		{
			wantGas:    big.NewInt(412),
			wantRefund: big.NewInt(0),
			originalV:  big.NewInt(0),
			iters:      [3]int64{0, 0, -1},
		},
		{
			wantGas:    big.NewInt(20212),
			wantRefund: big.NewInt(0),
			originalV:  big.NewInt(0),
			iters:      [3]int64{0, 1, -1},
		},
		{
			wantGas:    big.NewInt(20212),
			wantRefund: big.NewInt(19800),
			originalV:  big.NewInt(0),
			iters:      [3]int64{1, 0, -1},
		},
	}

	for i, test := range cases {
		gotGas := big.NewInt(0)
		gotRefund := big.NewInt(0)
		var currentV *big.Int
		for _, iter := range test.iters {
			if iter == -1 {
				break
			}
			if currentV == nil {
				currentV = test.originalV
			}
			newV := new(big.Int).SetUint64(uint64(iter))
			gas, ref := eip1283sstoreGas(test.originalV, currentV, newV)

			gotGas.Add(gotGas, gas)
			gotRefund.Add(gotRefund, ref)

			currentV.Set(newV)
		}
		gotGas.Add(gotGas, big.NewInt(12)) // TODO: why is this again?
		if gotGas.Cmp(test.wantGas) != 0 {
			t.Errorf("test: %v; got: %v, want: %v", i+1, gotGas, test.wantGas)
		}
		if gotRefund.Cmp(test.wantRefund) != 0 {
			t.Errorf("test: %v; got: %v, want: %v", i+1, gotRefund, test.wantRefund)
		}
	}
}
