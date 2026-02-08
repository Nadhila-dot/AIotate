package store

import (
	"fmt"
	"time"
)

// CreateStyle creates a new style for a user
func CreateStyle(db *DB, username, name, prompt, description string, isDefault bool) (*Style, error) {
	store, err := db.GetStore("styles")
	if err != nil {
		return nil, err
	}

	var styles map[string]map[string]Style
	if err := store.GetData(&styles); err != nil {
		styles = make(map[string]map[string]Style)
	}

	if _, exists := styles[username]; !exists {
		styles[username] = make(map[string]Style)
	}

	if _, exists := styles[username][name]; exists {
		return nil, fmt.Errorf("style %s already exists", name)
	}

	now := time.Now()
	style := Style{
		Name:        name,
		Username:    username,
		Prompt:      prompt,
		Description: description,
		IsDefault:   isDefault,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if isDefault {
		for key, s := range styles[username] {
			s.IsDefault = false
			styles[username][key] = s
		}
	}

	styles[username][name] = style

	if err := store.SetData(styles); err != nil {
		return nil, err
	}

	return &style, nil
}

// UpdateStyle updates an existing style for a user
func UpdateStyle(db *DB, username, name, prompt, description string) (*Style, error) {
	store, err := db.GetStore("styles")
	if err != nil {
		return nil, err
	}

	var styles map[string]map[string]Style
	if err := store.GetData(&styles); err != nil {
		return nil, err
	}

	userStyles, exists := styles[username]
	if !exists {
		return nil, fmt.Errorf("no styles found for user %s", username)
	}

	style, exists := userStyles[name]
	if !exists {
		return nil, fmt.Errorf("style %s not found", name)
	}

	style.Prompt = prompt
	style.Description = description
	style.UpdatedAt = time.Now()

	userStyles[name] = style
	styles[username] = userStyles

	if err := store.SetData(styles); err != nil {
		return nil, err
	}

	return &style, nil
}

// GetStyle retrieves a style by name
func GetStyle(db *DB, username, name string) (*Style, error) {
	store, err := db.GetStore("styles")
	if err != nil {
		return nil, err
	}

	var styles map[string]map[string]Style
	if err := store.GetData(&styles); err != nil {
		return nil, err
	}

	userStyles, exists := styles[username]
	if !exists {
		return nil, fmt.Errorf("no styles found for user %s", username)
	}

	style, exists := userStyles[name]
	if !exists {
		return nil, fmt.Errorf("style %s not found", name)
	}

	return &style, nil
}

// GetAllStyles retrieves all styles for a user
func GetAllStyles(db *DB, username string) ([]Style, error) {
	store, err := db.GetStore("styles")
	if err != nil {
		return nil, err
	}

	var styles map[string]map[string]Style
	if err := store.GetData(&styles); err != nil {
		return []Style{}, nil
	}

	userStyles, exists := styles[username]
	if !exists {
		return []Style{}, nil
	}

	result := make([]Style, 0, len(userStyles))
	for _, style := range userStyles {
		result = append(result, style)
	}

	return result, nil
}

// DeleteStyle removes a style by name
func DeleteStyle(db *DB, username, name string) error {
	store, err := db.GetStore("styles")
	if err != nil {
		return err
	}

	var styles map[string]map[string]Style
	if err := store.GetData(&styles); err != nil {
		return err
	}

	userStyles, exists := styles[username]
	if !exists {
		return nil
	}

	if _, exists := userStyles[name]; !exists {
		return nil
	}

	delete(userStyles, name)
	styles[username] = userStyles

	return store.SetData(styles)
}

// SetDefaultStyle marks a style as default for a user
func SetDefaultStyle(db *DB, username, name string) (*Style, error) {
	store, err := db.GetStore("styles")
	if err != nil {
		return nil, err
	}

	var styles map[string]map[string]Style
	if err := store.GetData(&styles); err != nil {
		return nil, err
	}

	userStyles, exists := styles[username]
	if !exists {
		return nil, fmt.Errorf("no styles found for user %s", username)
	}

	style, exists := userStyles[name]
	if !exists {
		return nil, fmt.Errorf("style %s not found", name)
	}

	for key, s := range userStyles {
		s.IsDefault = false
		userStyles[key] = s
	}

	style.IsDefault = true
	style.UpdatedAt = time.Now()
	userStyles[name] = style
	styles[username] = userStyles

	if err := store.SetData(styles); err != nil {
		return nil, err
	}

	return &style, nil
}

// GetDefaultStyle retrieves the default style for a user
func GetDefaultStyle(db *DB, username string) (*Style, error) {
	store, err := db.GetStore("styles")
	if err != nil {
		return nil, err
	}

	var styles map[string]map[string]Style
	if err := store.GetData(&styles); err != nil {
		return nil, err
	}

	userStyles, exists := styles[username]
	if !exists {
		return nil, fmt.Errorf("no styles found for user %s", username)
	}

	for _, style := range userStyles {
		if style.IsDefault {
			return &style, nil
		}
	}

	return nil, fmt.Errorf("default style not set")
}
