package UserStructs

import (
	"fmt"
	"log"
	"os"

	"strings"
)

//reasonable minimum and maximum password lengths
const passwordMinLength int = 10
const passwordMaxLength int = 256

//a password must contain at least one character from each character set
//
//character sets are compared by turning the candidate letter to lower case
const alphas string = "abcdefghijklmnopqrstuvwxyz"
const numerics string = "1234567890"
const additionals string = "!@#$^&*()-_=+[{]}|;:'\",<.>/?"

var characterSets = [...]string{alphas, numerics, additionals}

//returns if the password meets the length and complexity requirements
func passwordMeetsRequirements(password string) bool {

	complexity := true
	length := true

	if len(password) >= passwordMinLength &&
		len(password) < passwordMaxLength {
		length = true
	} else {
		//prevent potentially costly attempts to match
		//the password complexity requirements
		return false
	}

	for _, charSet := range characterSets {
		if !strings.ContainsAny(charSet, password) {
			complexity = false
		}
	}

	return complexity && length

}

//a quick wrapper around how we derive our passwords
func passwordDerivation(password string) (nonce, passwordHash []byte, err error) {

	nonce, err = getArrayOfRandBytes(32)
	if err != nil {
		return nil, nil, err
	}

	passwordHash, err = derivePasswordRaw([]byte(password), nonce)
	if err != nil {
		return nil, nil, err
	}

	return

}

func getLogger(fName, name string) (aLogger *log.Logger) {
	file, err := os.OpenFile(fName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Starting logger failed, I have no mouth but must scream!")
		fmt.Println(err)
		os.Exit(0)
	}

	aLogger = log.New(file, name+" ", log.Ldate|log.Ltime|log.Lshortfile)

	return
}
