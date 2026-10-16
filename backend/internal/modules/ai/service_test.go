package ai

import (
	"math"
	"testing"
)

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a        []float32
		b        []float32
		expected float64
		epsilon  float64
	}{
		{
			name:     "Identical vectors",
			a:        []float32{1.0, 0.0, 0.0},
			b:        []float32{1.0, 0.0, 0.0},
			expected: 1.0,
			epsilon:  0.0001,
		},
		{
			name:     "Orthogonal vectors",
			a:        []float32{1.0, 0.0, 0.0},
			b:        []float32{0.0, 1.0, 0.0},
			expected: 0.0,
			epsilon:  0.0001,
		},
		{
			name:     "Opposite vectors",
			a:        []float32{1.0, 0.0},
			b:        []float32{-1.0, 0.0},
			expected: -1.0,
			epsilon:  0.0001,
		},
		{
			name:     "Different length vectors return zero",
			a:        []float32{1.0, 0.0},
			b:        []float32{1.0, 0.0, 1.0},
			expected: 0.0,
			epsilon:  0.0001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sim := cosineSimilarity(tt.a, tt.b)
			if math.Abs(sim-tt.expected) > tt.epsilon {
				t.Errorf("expected similarity %f, got %f", tt.expected, sim)
			}
		})
	}
}

func TestKeywordScore(t *testing.T) {
	question := "What are the quarterly financial results?"
	contentMatching := "This report covers our quarterly financial results for Q3, showing 20% growth."
	contentNoMatch := "Human resources update regarding holiday schedules and annual leave."

	scoreMatch := keywordScore(question, contentMatching)
	scoreNoMatch := keywordScore(question, contentNoMatch)

	if scoreMatch <= 0 {
		t.Errorf("expected positive keyword match score, got %f", scoreMatch)
	}

	if scoreNoMatch != 0 {
		t.Errorf("expected 0 match score for irrelevant content, got %f", scoreNoMatch)
	}

	if scoreMatch <= scoreNoMatch {
		t.Errorf("expected matching content score to be higher than non-matching: %f vs %f", scoreMatch, scoreNoMatch)
	}
}
