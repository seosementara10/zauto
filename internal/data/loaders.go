package data

import (
	"fmt"
	"os"
	"strings"
)

type NameEntry struct {
	First  string
	Last   string
	Gender string
}

func LoadNames(path string) ([]NameEntry, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var names []NameEntry
	for i, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			return nil, fmt.Errorf("%s:%d invalid name line", path, i+1)
		}
		names = append(names, NameEntry{
			First:  strings.TrimSpace(parts[0]),
			Last:   strings.TrimSpace(parts[1]),
			Gender: normalizeGender(strings.TrimSpace(parts[2])),
		})
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("no names in %s", path)
	}
	return names, nil
}

func GetName(path string, index int) (NameEntry, error) {
	names, err := LoadNames(path)
	if err != nil {
		return NameEntry{}, err
	}
	if index < 0 || index >= len(names) {
		return NameEntry{}, fmt.Errorf("name_index %d out of range", index)
	}
	return names[index], nil
}

func LoadEmails(path string) ([]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var emails []string
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.Contains(line, "@") {
			continue
		}
		emails = append(emails, line)
	}
	if len(emails) == 0 {
		return nil, fmt.Errorf("no emails in %s", path)
	}
	return emails, nil
}

func GetEmail(path string, index int) (string, error) {
	emails, err := LoadEmails(path)
	if err != nil {
		return "", err
	}
	if index < 0 || index >= len(emails) {
		return "", fmt.Errorf("email_index %d out of range", index)
	}
	return emails[index], nil
}

func normalizeGender(g string) string {
	g = strings.ToLower(strings.TrimSpace(g))
	switch g {
	case "m", "male", "laki-laki", "pria":
		return "male"
	case "f", "female", "perempuan", "wanita":
		return "female"
	default:
		return g
	}
}
