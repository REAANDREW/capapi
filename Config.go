package main

const (
	//KeySize the size of the API Keys in bits
	KeySize = 512
)

//KeySizeBytes returns the size of the key in bytes
func KeySizeBytes() uint {
	return KeySize / 8
}
