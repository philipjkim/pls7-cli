package poker

import "github.com/sirupsen/logrus"

// combinations returns all unique combinations of `n` cards from the given `pool`.
// This is a recursive helper function for hand evaluation where specific numbers
// of cards must be drawn from different sets (e.g., hole cards vs. community cards).
func combinations(pool []Card, n int) [][]Card {
	logrus.Debugf("combinations called with pool size %d, n=%d", len(pool), n)

	// If n is 0, return a slice containing one empty slice, representing one combination of zero cards.
	if n == 0 {
		logrus.Debug("n is 0, returning one empty combination")
		return [][]Card{{}}
	}
	// If the pool is too small to form a combination of size n, return nil.
	if len(pool) < n {
		logrus.Debugf("pool size %d is less than n %d, returning nil", len(pool), n)
		return nil
	}

	// If we need to choose exactly as many cards as are in the pool, return the pool itself as the only combination.
	if len(pool) == n {
		logrus.Debugf("pool size equals n (%d), returning the pool as the only combination", n)
		// Make a copy to avoid modifying the original slice
		newPool := make([]Card, len(pool))
		copy(newPool, pool)
		return [][]Card{newPool}
	}

	// Recursive step:
	// 1. Get combinations from the rest of the pool (pool[1:]) that are of size n-1.
	//    These are the combinations that include the first element (pool[0]).
	logrus.Debugf("recursing for combinations WITH first element: combinations(pool_size=%d, n=%d)", len(pool)-1, n-1)
	subCombinationsWithFirst := combinations(pool[1:], n-1)
	for i := range subCombinationsWithFirst {
		// Prepend the first element to each of these sub-combinations.
		subCombinationsWithFirst[i] = append([]Card{pool[0]}, subCombinationsWithFirst[i]...)
	}
	logrus.Debugf("found %d combinations WITH %s", len(subCombinationsWithFirst), pool[0])

	// 2. Get combinations from the rest of the pool (pool[1:]) that are of size n.
	//    These are the combinations that do not include the first element.
	logrus.Debugf("recursing for combinations WITHOUT first element: combinations(pool_size=%d, n=%d)", len(pool)-1, n)
	subCombinationsWithoutFirst := combinations(pool[1:], n)
	logrus.Debugf("found %d combinations WITHOUT %s", len(subCombinationsWithoutFirst), pool[0])

	// Combine the two sets of combinations.
	result := append(subCombinationsWithFirst, subCombinationsWithoutFirst...)
	logrus.Debugf("returning %d total combinations", len(result))
	return result
}
