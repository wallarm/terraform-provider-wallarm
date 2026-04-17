package wallarm

import (
	crand "crypto/rand"
	"fmt"
	"math/big"
	"unicode"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/wallarm/wallarm-go"
)

type ruleNotFoundError struct {
	clientID int
	ruleID   int
}

func (e *ruleNotFoundError) Error() string {
	return fmt.Sprintf("rule %d for client %d not found", e.ruleID, e.clientID)
}

func expandInterfaceToStringList(list interface{}) []string {
	ifaceList := list.([]interface{})
	vs := []string{}
	for _, v := range ifaceList {
		vs = append(vs, v.(string))
	}
	return vs
}

func interfaceToString(i interface{}) string {
	r, _ := i.(string)
	return r
}

func interfaceToInt(i interface{}) int {
	r, _ := i.(int)
	return r
}

// retrieveClientID extracts client_id from a resource or falls back to the provider default.
func retrieveClientID(d *schema.ResourceData, m interface{}) (int, error) {
	meta := m.(*ProviderMeta)
	return meta.RetrieveClientID(d)
}

// apiClient extracts the wallarm.API client from the provider meta.
func apiClient(m interface{}) wallarm.API {
	return m.(*ProviderMeta).Client
}

func passwordGenerate(length int) (string, error) {
	digits := "0123456789"
	specials := "~=+%^*()_[]{}!@#$?"
	all := "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		digits + specials
	buf := make([]byte, length)
	var err error
	if buf[0], err = cryptoRandByte(digits); err != nil {
		return "", err
	}
	if buf[1], err = cryptoRandByte(specials); err != nil {
		return "", err
	}
	for i := 2; i < length; i++ {
		if buf[i], err = cryptoRandByte(all); err != nil {
			return "", err
		}
	}
	// Fisher-Yates shuffle using crypto/rand
	for i := len(buf) - 1; i > 0; i-- {
		j, err := cryptoRandIntn(i + 1)
		if err != nil {
			return "", err
		}
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf), nil
}

func cryptoRandByte(charset string) (byte, error) {
	idx, err := cryptoRandIntn(len(charset))
	if err != nil {
		return 0, err
	}
	return charset[idx], nil
}

func cryptoRandIntn(n int) (int, error) {
	maxN := big.NewInt(int64(n))
	v, err := crand.Int(crand.Reader, maxN)
	if err != nil {
		return 0, fmt.Errorf("crypto/rand failed: %w", err)
	}
	return int(v.Int64()), nil
}

func isPasswordValid(s string) bool {
	var (
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
	if len(s) >= 7 {
		hasMinLen = true
	}
	for _, char := range s {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	return hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial
}

func containsInt(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

func containsStr(slice []string, val string) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}
