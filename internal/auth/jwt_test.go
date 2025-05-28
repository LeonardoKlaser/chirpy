package auth

import (
	"testing"
	"time"
	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T){
	tests := []struct{
		UserID string
		TokenSecret string
		ExpiresIn int
		ExpectedError bool
	}{
		{
			UserID: "123e4567-e89b-12d3-a456-426614174000",
			TokenSecret: "mysecret",
			ExpiresIn: 3600,
			ExpectedError: false,
		},
		{
			UserID: "invalid-uuid",
			TokenSecret : "mysecret",
			ExpiresIn: 3600,
			ExpectedError: true,
		},
	}

	for _, tt := range tests{
		userID, err := uuid.Parse(tt.UserID)
		if err != nil {
			if tt.ExpectedError {
				continue // Skip this test case if the UUID is invalid
			}
			t.Errorf("Failed to parse UserID %s: %v", tt.UserID, err)
			continue
		}
		 token, err := MakeJWT(userID, tt.TokenSecret, time.Duration(tt.ExpiresIn) * time.Second)
		 if err != nil {
			t.Errorf("MakeJWT failed for UserID %s: %v", tt.UserID, err)
			continue
		 }

		 if token == "" {
			t.Errorf("MakeJWT returned an empty token for UserID %s", tt.UserID)
			continue
		}
		
	}
}