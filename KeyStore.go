package main

type KeyStore interface {
	Set(key string, scope []byte)
	Get(key string) ([]byte, error)
}
