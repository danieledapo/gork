package gork

import "testing"

func TestSkipWhites(t *testing.T) {
	data := []string{
		"",
		"    ",
		"\t\n\r   ",
	}
	expected := []int{
		-1,
		3,
		5,
	}

	for i, d := range data {
		if SkipWhites(d, 0) != expected[i] {
			t.Fail()
		}
	}
}

func TestSplitSentence(t *testing.T) {
	const wordsep string = ",."

	data := []string{
		"fred go fishing",
		"fred, go fishing",
		",.fred,go.,fishing,.",
		"   fred   go fishing  c",
	}
	expected := [][]string{
		[]string{"fred", "go", "fishing"},
		[]string{"fred", ",", "go", "fishing"},
		[]string{",", ".", "fred", ",", "go", ".", ",", "fishing", ",", "."},
		[]string{"fred", "go", "fishing", "c"},
	}

	for i, d := range data {
		words := SplitSentence(d, wordsep)

		for j, w := range words {
			if w != expected[i][j] {
				t.Fail()
			}
		}

	}

}
