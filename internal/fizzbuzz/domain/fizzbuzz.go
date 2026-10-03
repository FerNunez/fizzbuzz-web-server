package domain

import (
	"errors"
	"fmt"
	"strconv"

	"fizzbuzz-web-server/pkg/env"
)

var (
	MaxLimit     = env.GetInt("MAX_LIMIT", 10000)
	MaxStrLenght = env.GetInt("MAX_STR_LENGHT", 100)
)

var (
	ErrInvalidParams = errors.New("invalid fizzbuzz params")
	ErrEmptyRepo     = errors.New("empty repository")
)

type ParamStat struct {
	Params *FizzbuzzParams
	Hits   int
}

type FizzbuzzParams struct {
	Int1  int
	Int2  int
	Limit int
	Str1  string
	Str2  string
}

// Validate checks if the params are in the desired thesholds
func (p FizzbuzzParams) Validate() error {
	if p.Int1 <= 0 || p.Int2 <= 0 || p.Limit <= 0 {
		return fmt.Errorf("%w: int1 and int2 must be greater than 0", ErrInvalidParams)
	}
	if p.Limit <= 0 || p.Limit > MaxLimit {
		return fmt.Errorf("%w: integers must be in between 1..%d", ErrInvalidParams, MaxLimit)
	}
	if len(p.Str1) <= 0 || len(p.Str1) > MaxStrLenght || len(p.Str2) <= 0 || len(p.Str2) > MaxStrLenght {
		return fmt.Errorf("%w: strings cannot be null or lenghtier than %d chatacters", ErrInvalidParams, MaxStrLenght)
	}
	return nil
}

// GenerateFizzbuzz generates a fizzbuzz string from the FizzbuzzParams: Int1, Int2, Limit int & Str1, Str2 string
func GenerateFizzbuzz(params *FizzbuzzParams) []string {
	output := make([]string, params.Limit)
	for i := 0; i < params.Limit; i++ {
		number := i + 1
		if number%params.Int1 == 0 && number%params.Int2 == 0 {
			output[i] = params.Str1 + params.Str2
		} else if number%params.Int1 == 0 {
			output[i] = params.Str1
		} else if number%params.Int2 == 0 {
			output[i] = params.Str2
		} else {
			output[i] = strconv.Itoa(number)
		}
	}
	return output
}
