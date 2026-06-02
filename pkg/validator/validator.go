package validator

import (
	"strings"

	gpv "github.com/go-playground/validator/v10"

	"github.com/hafis915/fintrack/pkg/apperror"
)

type V struct{ v *gpv.Validate }

func New() *V { return &V{v: gpv.New()} }

func (x *V) Validate(i any) error { return x.v.Struct(i) }

func ToAppError(err error) *apperror.Error {
	if err == nil {
		return nil
	}
	verrs, ok := err.(gpv.ValidationErrors)
	if !ok {
		return apperror.Validation(err.Error(), nil)
	}
	fields := make(map[string]string, len(verrs))
	for _, fe := range verrs {
		fields[strings.ToLower(fe.Field())] = fe.Tag()
	}
	return apperror.Validation("validation failed", fields)
}
