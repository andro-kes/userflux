package agent

import "crypto/rand"

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
func generateBody(fields []string) (map[string]string, error) {
	m := make(map[string]string)
	for i := 0; i < len(fields); i++ {
		s, err := generateRandString()
		if err != nil {
			return nil, err
		}
		m[fields[i]] = s
	}

	return m, nil
}

func generateRandString() (string, error) {
	b := make([]byte, 10)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = letters[int(b[i])%len(letters)]
	}
	return string(b), nil
}
