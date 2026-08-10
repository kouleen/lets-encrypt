package main

import (
	"fmt"
	"testing"

	"github.com/kouleen/lets-encrypt/pkg/util"
)

func TestCheckRemoteCertExpire(T *testing.T) {
	certInfo, err := util.GetRemoteCertExpire("www.kouleen.cn", 3)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("证书过期时间: %s, 剩余天数: %d, 是否需要续期（剩余≤3 天）: %t", certInfo.ExpireTime.Format("2006-01-02 15:04:05"), certInfo.RemainDay, certInfo.NeedRenew)
}

func TestCheckLocalCertExpire(t *testing.T) {
	certInfo, err := util.GetLocalCertExpire("doc.kouleen.cn.pem", 3)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("证书过期时间: %s, 剩余天数: %d, 是否需要续期（剩余≤3 天）: %t", certInfo.ExpireTime.Format("2006-01-02 15:04:05"), certInfo.RemainDay, certInfo.NeedRenew)
}
