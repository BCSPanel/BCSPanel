package myrand

import (
	"crypto/rand"
	"math/big"
)

// 生成 0 到 max 的随机数，范围不包含 max 。
func RandBigInt(max *big.Int) *big.Int {
	rn, err := rand.Int(rand.Reader, max)
	if err != nil {
		panic(err)
	}
	return rn
}

// 基于 RandBigInt
func RandInt64(max int64) int64 {
	return RandBigInt(big.NewInt(max)).Int64()
}

// 基于 RandBigInt
func RandInt(max int) int {
	return int(RandInt64(int64(max)))
}

// 生成 0 到 max 的随机数，范围包含 max 。
// 如果 max 为 0 则抛出错误。
// 参考了 rand.Int
func RandUint8(max uint8) uint8 {
	b := max
	b |= b >> 1
	b |= b >> 2
	b |= b >> 4
	bytes := make([]byte, 1)
	for {
		_, err := rand.Read(bytes)
		if err != nil {
			panic(err)
		}
		// 仅保留必要的bits
		bytes[0] &= b
		// 如果选择的数在范围内就返回，否则重新生成
		if bytes[0] <= max {
			return bytes[0]
		}
	}
}
