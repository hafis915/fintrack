package apperror_test

import (
	"errors"
	"testing"

	"github.com/hafis915/fintrack/pkg/apperror"
	"github.com/stretchr/testify/require"
)

func TestNotFoundIs(t *testing.T) {
	e := apperror.NotFound("user", "id=x")
	require.Equal(t, apperror.CodeNotFound, e.Code)
	require.True(t, errors.Is(e, apperror.ErrNotFound))
}

func TestValidationFields(t *testing.T) {
	e := apperror.Validation("bad", map[string]string{"amount": "must be > 0"})
	require.Equal(t, "bad", e.Message)
	require.Equal(t, "must be > 0", e.Fields["amount"])
}
