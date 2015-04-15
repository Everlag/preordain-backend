package userDB

import(

	"fmt"
	"io"

	"code.google.com/p/go.crypto/scrypt"
	"crypto/rand"

)

// CSPRNG secure bytes are delicious
func getArrayOfRandBytes(arrayLength int) ([]byte, error) {
	workingArray := make([]byte, arrayLength)

	_, err := io.ReadFull(rand.Reader, workingArray)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire entropy")
	}

	return workingArray, nil
}

// Derives a password using scrypt. Requires plaintext and nonce
func derivePasswordWithNonce(plaintext, nonce []byte) ([]byte, error) {

	// Output a 32 byte hash using
	passwordHash, err := scrypt.Key([]byte(plaintext), nonce, 32768, 2, 1, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to derive password, try again")
	}

	return passwordHash, nil

}

// Derives a password using scrypt. Requires plaintext, nonce
// is returned alongside the hash
func derivePassword(plaintext []byte) (passwordHash, nonce []byte, err error) {

	nonce, err = getArrayOfRandBytes(32)
	if err!=nil {
		return
	}

	// Output a 32 byte hash using
	passwordHash, err= scrypt.Key([]byte(plaintext), nonce, 32768, 2, 1, 32)
	if err != nil {
		return
	}

	return

}

const alphanum = "!@#0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
// Returns a completely random string of length n
func randString(n int) string {
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

// Returns a random single byte
func randByte() byte {
	var bytes = make([]byte, 1)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return bytes[0]
}