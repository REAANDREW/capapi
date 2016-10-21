package main

import (
	"crypto/rand"
	"encoding/base64"

	log "github.com/Sirupsen/logrus"
)

//CreateKey creates a new 512 bit key
func CreateKey() (string, error) {
	keyBytes := make([]byte, 64)
	_, err := rand.Read(keyBytes)
	if err != nil {
		return "", err
	}

	key := base64.StdEncoding.EncodeToString(keyBytes)
	log.WithFields(log.Fields{
		"key": key,
	}).Info("Key Generated")

	return key, nil
}
