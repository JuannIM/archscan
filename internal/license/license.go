// Package license provides offline license key validation for archscan Pro.
//
// Key format: ARCHSCAN-XXXXXXXX-XXXXXXXX-XXXXXXXX-XXXXXXXX
// Keys are HMAC-SHA256 signatures of "email|plan|expiry" using a compiled-in seed.
// No server required for validation.
package license

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Plan represents the license tier.
type Plan string

const (
	PlanFree Plan = "free"
	PlanPro  Plan = "pro"
)

// License holds decoded license information.
type License struct {
	Email   string    `json:"email"`
	Plan    Plan      `json:"plan"`
	Expiry  time.Time `json:"expiry"`
	Key     string    `json:"key"`
}

// IsPro returns true if the license is a valid, non-expired Pro license.
func (l *License) IsPro() bool {
	return l != nil && l.Plan == PlanPro && time.Now().Before(l.Expiry)
}

// DaysRemaining returns how many days until the license expires.
func (l *License) DaysRemaining() int {
	if l == nil {
		return 0
	}
	d := time.Until(l.Expiry)
	if d < 0 {
		return 0
	}
	return int(d.Hours() / 24)
}

// xorSeed is the HMAC secret, obfuscated with XOR to avoid trivial string search.
// Key: 0x5A repeated.
var xorSeed = []byte{
	0x1b ^ 0x5a, 0x38 ^ 0x5a, 0x2c ^ 0x5a, 0x4f ^ 0x5a,
	0x0d ^ 0x5a, 0x3e ^ 0x5a, 0x19 ^ 0x5a, 0x2b ^ 0x5a,
	0x4c ^ 0x5a, 0x0e ^ 0x5a, 0x3d ^ 0x5a, 0x18 ^ 0x5a,
	0x2a ^ 0x5a, 0x4b ^ 0x5a, 0x0f ^ 0x5a, 0x3c ^ 0x5a,
	0x17 ^ 0x5a, 0x29 ^ 0x5a, 0x4a ^ 0x5a, 0x0c ^ 0x5a,
	0x3b ^ 0x5a, 0x16 ^ 0x5a, 0x28 ^ 0x5a, 0x49 ^ 0x5a,
	0x0b ^ 0x5a, 0x3a ^ 0x5a, 0x15 ^ 0x5a, 0x27 ^ 0x5a,
	0x48 ^ 0x5a, 0x0a ^ 0x5a, 0x39 ^ 0x5a, 0x14 ^ 0x5a,
}

func decodeSeed() []byte {
	out := make([]byte, len(xorSeed))
	for i, b := range xorSeed {
		out[i] = b ^ 0x5a
	}
	return out
}

// GenerateKey creates a new Pro license key for the given email and duration.
// This is used by the seller (you) to generate keys for customers.
func GenerateKey(email string, plan Plan, validDays int) (string, error) {
	expiry := time.Now().AddDate(0, 0, validDays).Format("2006-01-02")
	payload := fmt.Sprintf("%s|%s|%s", strings.ToLower(email), plan, expiry)

	mac := hmac.New(sha256.New, decodeSeed())
	mac.Write([]byte(payload))
	sig := mac.Sum(nil)

	// Encode first 20 bytes as base32, split into groups of 8
	encoded := base32.StdEncoding.EncodeToString(sig[:20])
	key := fmt.Sprintf("ARCHSCAN-%s-%s-%s-%s",
		encoded[0:8], encoded[8:16], encoded[16:24], encoded[24:32])

	return key, nil
}

// ValidateKey checks if the key is valid for the given email and plan.
func ValidateKey(key, email string, plan Plan) (*License, error) {
	key = strings.TrimSpace(strings.ToUpper(key))

	// Try all possible expiry dates within a 3-year window
	// (brute-force over dates is fine since HMAC validation is O(1) per date)
	now := time.Now()
	for days := -1; days <= 365*3; days++ {
		candidate := now.AddDate(0, 0, days).Format("2006-01-02")
		payload := fmt.Sprintf("%s|%s|%s", strings.ToLower(email), plan, candidate)

		mac := hmac.New(sha256.New, decodeSeed())
		mac.Write([]byte(payload))
		sig := mac.Sum(nil)

		encoded := base32.StdEncoding.EncodeToString(sig[:20])
		expected := fmt.Sprintf("ARCHSCAN-%s-%s-%s-%s",
			encoded[0:8], encoded[8:16], encoded[16:24], encoded[24:32])

		if hmac.Equal([]byte(key), []byte(expected)) {
			expiry, _ := time.Parse("2006-01-02", candidate)
			return &License{
				Email:  email,
				Plan:   plan,
				Expiry: expiry.Add(24*time.Hour - time.Second), // end of day
				Key:    key,
			}, nil
		}
	}

	return nil, fmt.Errorf("invalid license key")
}

// storePath returns the path to the stored license file.
func storePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "archscan")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "license.json"), nil
}

// Save persists the license to disk.
func (l *License) Save() error {
	path, err := storePath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// Load reads the license from disk. Returns a free license if none found.
func Load() *License {
	path, err := storePath()
	if err != nil {
		return freeLicense()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return freeLicense()
	}
	var l License
	if err := json.Unmarshal(data, &l); err != nil {
		return freeLicense()
	}
	// Re-validate the key to prevent tampering
	validated, err := ValidateKey(l.Key, l.Email, l.Plan)
	if err != nil || !validated.IsPro() {
		return freeLicense()
	}
	return validated
}

// Remove deletes the stored license.
func Remove() error {
	path, err := storePath()
	if err != nil {
		return err
	}
	return os.Remove(path)
}

func freeLicense() *License {
	return &License{
		Plan: PlanFree,
	}
}
