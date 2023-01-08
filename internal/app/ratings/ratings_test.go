package ratings

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestComputeNewElos(t *testing.T) {
	anandRating := 2600
	borisRating := 2300

	newAnandRating, newBorisRating := ComputeNewElos(anandRating, borisRating, true, K, K)

	assert.Equal(t, 2604, newAnandRating)
	assert.Equal(t, 2295, newBorisRating)

	newAnandRating2, newBorisRating2 := ComputeNewElos(anandRating, borisRating, false, K, K)
	assert.Equal(t, 2572, newAnandRating2)
	assert.Equal(t, 2327, newBorisRating2)
}
