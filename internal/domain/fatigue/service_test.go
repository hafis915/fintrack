package fatigue_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hafis915/fintrack/internal/domain/fatigue"
)

func TestStatus_Thresholds(t *testing.T) {
	require.Equal(t, "fresh", fatigue.ComputeStatus(0, 100))
	require.Equal(t, "fresh", fatigue.ComputeStatus(59, 100))
	require.Equal(t, "warning", fatigue.ComputeStatus(60, 100))
	require.Equal(t, "warning", fatigue.ComputeStatus(84, 100))
	require.Equal(t, "fatigued", fatigue.ComputeStatus(85, 100))
}

func TestStatus_ZeroAllocated(t *testing.T) {
	require.Equal(t, "fresh", fatigue.ComputeStatus(100, 0))
}
