package service

import (
	"context"

	"github.com/kouleen/lets-encrypt/internal/modle"
)

func CreateAcmeAccount(ctx context.Context, req *modle.AcmeAccountRegister) (any, error) {
	return true, nil
}
