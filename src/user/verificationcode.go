package user

import (
	"crypto/hmac"
	"sync"
	"time"

	"github.com/bddjr/BCSPanel/src/myrand"
)

const VerificationCodeTimeout = 10 * time.Minute

type VerifyCodeType struct {
	Code           string
	ExpirationTime time.Time
	lock           sync.Mutex
}

var RegisterVerifyCode = &VerifyCodeType{
	Code:           "",
	ExpirationTime: time.UnixMicro(0),
}

func (vc *VerifyCodeType) clear() {
	vc.Code = ""
	vc.ExpirationTime = time.UnixMicro(0)
}

func (vc *VerifyCodeType) IsValid() bool {
	return vc.Code != "" && time.Now().Before(RegisterVerifyCode.ExpirationTime)
}

func (vc *VerifyCodeType) IsValidWithAutoClear() bool {
	vc.lock.Lock()
	defer vc.lock.Unlock()
	if vc.Code == "" {
		return false
	}
	if time.Now().Before(RegisterVerifyCode.ExpirationTime) {
		vc.clear()
		return false
	}
	return true
}

func (vc *VerifyCodeType) Fill() *VerifyCodeType {
	vc.lock.Lock()
	defer vc.lock.Unlock()
	vc.Code = myrand.RandStr64(16)
	vc.ExpirationTime = time.Now().Add(VerificationCodeTimeout)
	return vc
}

func (vc *VerifyCodeType) CodeEqual(inCode string) bool {
	return vc.IsValid() && hmac.Equal([]byte(inCode), []byte(vc.Code))
}

func (vc *VerifyCodeType) CodeEqualWithAutoClear(inCode string) bool {
	if vc.CodeEqual(inCode) {
		vc.clear()
		return true
	}
	return false
}
