package auth

import "testing"

func TestHashPassword(t *testing.T) {
	usecases := []struct {
		name     string
		password string
	}{
		{
			name:     "tough password",
			password: "pa$$word",
		},
		{
			name:     "easy password",
			password: "name",
		},
	}

	for i, val := range usecases {
		t.Run(val.name, func(t *testing.T) {
			hash, err := HashPassword(val.password)
			if err != nil {
				t.Errorf("Hash Failed")
				return
			}

			match, err := CheckPasswordHash(val.password, hash)
			if err != nil {
				t.Errorf("Check hash pasword Failed")
				return
			}

			if !match {
				t.Errorf("TEST: %v, Failed to match password: GOT: %v, RESULT: %v", i, hash, match)
			}
		})
	}
}
