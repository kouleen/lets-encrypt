package service

import (
	"context"

	"github.com/kouleen/lets-encrypt/internal/modle"
)

func SendCode(ctx context.Context, email string) (any, error) {
	return true, nil
}

func CreateAcmeAccount(ctx context.Context, req *modle.AcmeAccountRegister) (any, error) {
	return true, nil
}
