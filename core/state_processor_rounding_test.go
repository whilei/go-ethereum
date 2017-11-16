package core

import (
	"testing"
	"math/big"
	"github.com/ethereumproject/go-ethereum/core/types"
)

// Test block winner rewards
func TestBlockRewardsInFuture(t *testing.T) {
	for i := 1; i < 50; i++ {
		era := big.NewInt(int64(i))

		eraSemantic := new(big.Int).Add(era, big.NewInt(1)) // ECIP1017 PR uses 1-indexed era labeling

		blockWinnerReward := GetBlockWinnerRewardByEra(era)

		eraUncleReward := getEraUncleBlockReward(era)
		era2UnclesReward := new(big.Int).Add(eraUncleReward, eraUncleReward)

		oneUncle := []*types.Header{{}}
		twoUncles := []*types.Header{{}, {}}
		if era1UncleWant := GetBlockWinnerRewardForUnclesByEra(era, oneUncle); eraUncleReward.Cmp(era1UncleWant) != 0 {
			t.Errorf("eraUncleReward got=%v want=%v", eraUncleReward, era1UncleWant)
		}
		if era2UnclesWant := GetBlockWinnerRewardForUnclesByEra(era, twoUncles); era2UnclesReward.Cmp(era2UnclesWant) != 0 {
			t.Errorf("era2UncleReward got=%v want=%v", eraUncleReward, era2UnclesWant)
		}

		blockWinnerRewardOneUncle := new(big.Int).Add(blockWinnerReward, eraUncleReward)
		blockWinnerRewardTwoUncles := new(big.Int).Add(blockWinnerReward, era2UnclesReward)
		if r := GetBlockWinnerRewardForUnclesByEra(era, oneUncle); blockWinnerRewardOneUncle.Cmp(new(big.Int).Add(blockWinnerReward, r)) != 0 {
			t.Errorf("eraWinner1UncleReward got=%v want=%v", blockWinnerRewardOneUncle, new(big.Int).Add(blockWinnerReward, r))
		}
		if r := GetBlockWinnerRewardForUnclesByEra(era, twoUncles); blockWinnerRewardTwoUncles.Cmp(new(big.Int).Add(blockWinnerReward, r)) != 0 {
			t.Errorf("eraWinner2UnclesReward got=%v want=%v", blockWinnerRewardTwoUncles, new(big.Int).Add(blockWinnerReward, r))
		}

		t.Logf("| era=%d | blockWinnerR=%d | 1uncleR=%d | 2unclesR=%d | blockWinner1UncleR=%d | blockWinner2UnclesR=%d |",
			eraSemantic, blockWinnerReward, eraUncleReward, era2UnclesReward, blockWinnerRewardOneUncle, blockWinnerRewardTwoUncles)

	}
}

func TestBlockWinnerDiffDemo(t *testing.T) {
	for i := 1; i < 50; i++ {
		era := big.NewInt(int64(i))
		eraSemantic := new(big.Int).Add(era, big.NewInt(1)) // again, 1-indexed in official spec

		// Get reward SHOULD (as-is)
		wrOK := GetBlockWinnerRewardByEra(era)

		// Get reward in the way it SHOULD NOT be
		maxR := MaximumBlockReward
		wrNotOK := maxR
		for j := 1; j <= i; j++ {
			wrNotOK = new(big.Int).Mul(wrNotOK, DisinflationRateQuotient)
			wrNotOK = new(big.Int).Div(wrNotOK, DisinflationRateDivisor)
		}

		t.Logf("| era=%d | winnerOK=%d | winnerNotOK=%d |", eraSemantic, wrOK, wrNotOK)
	}
}

// Test block winner uncle inclusion rewards
func TestBlockWinner2UnclesDiffDemo(t *testing.T) {
	for i := 1; i < 200; i++ {
		era := big.NewInt(int64(i))
		eraSemantic := new(big.Int).Add(era, big.NewInt(1)) // again, 1-indexed in official spec

		br := GetBlockWinnerRewardByEra(era)
		wruOK := GetBlockWinnerRewardForUnclesByEra(era, []*types.Header{{}, {}})

		wruNotOK := new(big.Int).Div(br, big.NewInt(16)) // divide by 16 instead of /32*2

		//wruNotOK := GetBlockWinnerRewardForUnclesByEra(era, []*types.Header{{}})
		//wruNotOK = new(big.Int).Mul(wruNotOK, big.NewInt(2))
		//
		brOK := new(big.Int).Add(br, wruOK)
		brNotOK := new(big.Int).Add(br, wruNotOK)

		t.Logf("| era=%d | blockReward=%d | brOK=%d | brNotOK=%d | wruOK=%d | wruNotOK=%d |", eraSemantic, br, brOK, brNotOK, wruOK, wruNotOK)
	}
}

// Test block uncle miner inclusion rewards

