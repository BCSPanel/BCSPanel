package myrand

// 生成长度为n的随机字符串，从letters里抽取字符
func RandStr(n int, letters []byte) string {
	b := make([]byte, n)
	max := uint8(len(letters) - 1)
	for i := range b {
		b[i] = letters[RandUint8(max)]
	}
	return string(b)
}

func RandStr95(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 32 + RandUint8(126-32)
	}
	return string(b)
}

func RandStr94(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 33 + RandUint8(126-33)
	}
	return string(b)
}

var Str16Lower = []byte("0123456789abcdef")

func RandStr16Lower(n int) string {
	return RandStr(n, Str16Lower)
}

var Str16Upper = []byte("0123456789ABCDEF")

func RandStr16Upper(n int) string {
	return RandStr(n, Str16Upper)
}

var Str36Lower = []byte("0123456789abcdefghijklmnopqrstuvwxyz")

func RandStr36Lower(n int) string {
	return RandStr(n, Str36Lower)
}

var Str36Upper = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStr36Upper(n int) string {
	return RandStr(n, Str36Upper)
}

var Str62 = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStr62(n int) string {
	return RandStr(n, Str62)
}

var Str64 = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-_")

func RandStr64(n int) string {
	return RandStr(n, Str64)
}
