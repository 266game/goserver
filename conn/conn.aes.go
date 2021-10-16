package conn

import "github.com/yangtizi/crypto/aes"

type TAes struct {
	binAesKey  []byte // aesKey
	binIV      []byte // 填充
	strPadding string // PKCS7
}

func NewAES(binAesKey []byte) *TAes {
	return &TAes{
		binAesKey: binAesKey,
		binIV: []byte{
			0, 0, 0, 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
			0, 0, 0, 0},
		strPadding: "PKCS7",
	}
}

func (m *TAes) SetAesKey(binAesKey []byte) {
	m.binAesKey = binAesKey
}

func (m *TAes) SetIV(binIV []byte) {
	m.binIV = binIV
}

// CoAES  aes进行加密
func (m *TAes) CoAES(buff []byte) ([]byte, error) {
	return aes.CoAES(buff, m.binAesKey, m.binIV, m.strPadding)
}

// UnAES aes进行解密
func (m *TAes) UnAES(buff []byte) ([]byte, error) {
	return aes.UnAES(buff, m.binAesKey, m.binIV, m.strPadding)
}
