package repository

import (
	"math/big"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
)

// pgNumeric serializes a float64 percentage into the pgtype.Numeric shape
// the generated code expects. We use 2 d.p. — matching domain/budget.pct.
func pgNumeric(f float64) pgtype.Numeric {
	n := pgtype.Numeric{}
	// Set via the string path so we don't fight floating-point representation.
	_ = n.Scan(strconv.FormatFloat(f, 'f', 2, 64))
	return n
}

// fromPgNumeric reverses the above. Treats null / NaN as 0 — the
// percentage column is NOT NULL once a plan is written, so this is just
// defensive.
func fromPgNumeric(n pgtype.Numeric) float64 {
	if !n.Valid || n.NaN {
		return 0
	}
	if n.Int == nil {
		return 0
	}
	// pgtype.Numeric is base-10: value = Int * 10^Exp
	intPart := new(big.Float).SetInt(n.Int)
	pow := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(absInt32(n.Exp))), nil))
	var result *big.Float
	if n.Exp >= 0 {
		result = new(big.Float).Mul(intPart, pow)
	} else {
		result = new(big.Float).Quo(intPart, pow)
	}
	out, _ := result.Float64()
	return out
}

func absInt32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}
