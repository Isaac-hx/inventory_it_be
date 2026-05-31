// this file contain e2e test integration to handler
package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

const baseURL = "http://localhost:2511"

type LoginResponse struct {
	Token string `json:"token"`
}

func loginAndGetToken(t *testing.T) string {
	t.Helper()

	loginData := userRequest{
		Username: "isaachx",
		Password: "saydimas78",
	}

	jsonData, err := json.Marshal(loginData)
	if err != nil {
		t.Fatalf("Failed to marshal json: %v", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		baseURL+"/login",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected login status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if loginResp.Token == "" {
		t.Fatal("Login token is empty")
	}

	return loginResp.Token
}

func TestLogin(t *testing.T) {
	t.Run("positive case - login success", func(t *testing.T) {
		token := loginAndGetToken(t)

		if token == "" {
			t.Fatal("Expected token, got empty token")
		}
	})

	t.Run("negative case - wrong password", func(t *testing.T) {
		loginData := userRequest{
			Username: "isaachx",
			Password: "passwordsalah",
		}

		jsonData, err := json.Marshal(loginData)
		if err != nil {
			t.Fatalf("Failed to marshal json: %v", err)
		}

		req, err := http.NewRequest(
			http.MethodPost,
			baseURL+"/login",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("Expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
		}
	})

	t.Run("negative case - username not found", func(t *testing.T) {
		loginData := userRequest{
			Username: "user_tidak_ada",
			Password: "saydimas78",
		}

		jsonData, err := json.Marshal(loginData)
		if err != nil {
			t.Fatalf("Failed to marshal json: %v", err)
		}

		req, err := http.NewRequest(
			http.MethodPost,
			baseURL+"/login",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("Expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
		}
	})

	t.Run("negative case - empty request body", func(t *testing.T) {
		req, err := http.NewRequest(
			http.MethodPost,
			baseURL+"/login",
			bytes.NewBuffer([]byte{}),
		)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})
}

func TestRegister(t *testing.T) {
	t.Run("positive case - register success", func(t *testing.T) {
		token := loginAndGetToken(t)

		registerData := userRequest{
			Username:      "testuser",
			Password:      "testuser123",
			Email:         "testuser@gmail.com",
			Role:          "admin_it",
			Department_id: "DEPT-IT",
		}

		jsonData, err := json.Marshal(registerData)
		if err != nil {
			t.Fatalf("Failed to marshal json: %v", err)
		}

		req, err := http.NewRequest(
			http.MethodPost,
			baseURL+"/register",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d", http.StatusCreated, resp.StatusCode)
		}
	})

	t.Run("negative case - register without token", func(t *testing.T) {
		registerData := userRequest{
			Username:      "testuser2",
			Password:      "testuser123",
			Email:         "testuser2@gmail.com",
			Role:          "admin_it",
			Department_id: "DEPT-IT",
		}

		jsonData, err := json.Marshal(registerData)
		if err != nil {
			t.Fatalf("Failed to marshal json: %v", err)
		}

		req, err := http.NewRequest(
			http.MethodPost,
			baseURL+"/register",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("Expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
		}
	})
}
