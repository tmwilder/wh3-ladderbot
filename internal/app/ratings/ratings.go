package ratings

import "math"

const K = 32.0

func ComputeNewElos(p1Rating int, p2Rating int, p1Won bool) (newP1Rating int, newP2Rating int) {
	expectedScore1 := 1.0 / (1.0 + math.Pow(10, (float64(p2Rating)-float64(p1Rating))/400))
	expectedScore2 := 1.0 / (1.0 + math.Pow(10, (float64(p1Rating)-float64(p2Rating))/400))

	p1Score := 0.0
	p2Score := 0.0

	if p1Won {
		p1Score = 1.0
	} else {
		p2Score = 1.0
	}
	newP1Rating = int(float64(p1Rating) + K*(p1Score-expectedScore1))
	newP2Rating = int(float64(p2Rating) + K*(p2Score-expectedScore2))
	return newP1Rating, newP2Rating
}
