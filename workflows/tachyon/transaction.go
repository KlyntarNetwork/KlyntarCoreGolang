package tachyon

type Transaction struct {
	V                                uint
	Fee, Creator, Sig, Type, SigType string
	Nonce                            int
	Payload                          map[string]interface{}
}
