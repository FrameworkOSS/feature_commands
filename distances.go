package commands

import (
	"math"
	"strings"
)

const (
	phraseWeight = 0.15
	wordsWeight  = 0.20
	lengthWeight = 0.04
	startWeight  = 0.50

	minWeight = 0.1
	maxWeight = 1.0
)

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func LevenshteinDistance(input string, test ...string) (match string, distance float64) {
	min := func(a, b float64) float64 {
		if a < b {
			return a
		}
		return b
	}

	max := func(a, b float64) float64 {
		if a > b {
			return a
		}
		return b
	}

	lev := func(a, b string) float64 {
		r1 := []rune(strings.ToLower(a))
		r2 := []rune(strings.ToLower(b))

		l1, l2 := len(r1), len(r2)
		d := make([][]int, l1+1)
		for i := range d {
			d[i] = make([]int, l2+1)
		}

		for i := 0; i <= l1; i++ {
			d[i][0] = i
		}
		for j := 0; j <= l2; j++ {
			d[0][j] = j
		}

		for j := 1; j <= l2; j++ {
			for i := 1; i <= l1; i++ {
				cost := 0
				if r1[i-1] != r2[j-1] {
					cost = 1
				}
				d[i][j] = minInt(
					d[i-1][j]+1,
					minInt(d[i][j-1]+1, d[i-1][j-1]+cost),
				)
			}
		}
		return float64(d[l1][l2])
	}

	split := func(s string) []string {
		return strings.FieldsFunc(s, func(r rune) bool {
			return r == ' ' || r == '_' || r == '-'
		})
	}

	wordValue := func(a, b string) float64 {
		w1 := split(a)
		w2 := split(b)

		total := 0.0
		for _, x := range w1 {
			best := float64(len(b))
			for _, y := range w2 {
				d := lev(x, y)
				if d < best {
					best = d
				}
				if d == 0 {
					break
				}
			}
			total += best
		}
		return total
	}

	startValue := func(a, b string) float64 {
		total := 0.0
		lenA := float64(len(a))
		lenB := float64(len(b))

		for i := 0; float64(i) < min(lenA, lenB); i++ {
			if a[i] != b[i] {
				break
			}
			total += 1
		}
		return total
	}

	score := func(a, b string) float64 {
		phraseValue := lev(a, b)
		wordsValue := wordValue(a, b)
		lengthValue := math.Abs(float64(len(a) - len(b)))
		startValue := startValue(a, b)

		pw := phraseWeight * phraseValue
		ww := wordsWeight * wordsValue
		sw := startWeight * startValue

		score := min(pw, ww)*minWeight +
			max(pw, ww)*maxWeight +
			lengthWeight*lengthValue -
			sw

		return score
	}

	best := math.MaxFloat64
	bestMatch := ""

	for _, t := range test {
		d := score(input, t)
		if d < best {
			best = d
			bestMatch = t
		}
	}

	return bestMatch, best
}
