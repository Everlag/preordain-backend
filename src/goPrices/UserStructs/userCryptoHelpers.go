package UserStructs

import (
	"fmt"

	//password derivation and storage specific
	"code.google.com/p/go.crypto/scrypt"
	"crypto/rand"
	"io"
)

const alphanum = "!@#0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// gets an array of random bytes from the crypto generator
func getArrayOfRandBytes(arrayLength int) ([]byte, error) {
	workingArray := make([]byte, arrayLength)

	_, err := io.ReadFull(rand.Reader, workingArray)
	if err != nil {
		return nil, fmt.Errorf("Failed to acquire entropy")
	}

	return workingArray, nil
}

//derives a password using parameters specified in function.
//
//Uses the wonderful scrypt. Arguably better than bcrypt
func derivePasswordRaw(plaintext, nonce []byte) ([]byte, error) {
	//ok, now for scrypt parameters
	//we input the password and the nonce as usual
	//
	//the difficulty is 2**15 which is 2**5 times weaker than what I've been
	//using for key derivation, that's fine
	//
	//the memory usage is 2 which allows for memory hardness
	//parallelism is low
	//output is a 256bit password hash.
	passwordHash, err := scrypt.Key([]byte(plaintext), nonce, 32768, 2, 1, 32)
	if err != nil {
		return nil, fmt.Errorf("Failed to derive password, try again")
	}

	return passwordHash, nil

}

//to populate reset tokens
func randString(n int) string {
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}
