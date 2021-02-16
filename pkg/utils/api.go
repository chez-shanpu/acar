package utils

import "github.com/chez-shanpu/acar/api/types"

func NewResult(ok bool, errStr string) *types.Result {
	return &types.Result{
		Ok: ok,
	}
}
