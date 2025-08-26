package domain

// IsValid validates the category data
func (c *Category) IsValid() error {
	if c.Name == "" {
		return ErrInvalidInput("category name is required")
	}
	if c.CreatorID == "" {
		return ErrInvalidInput("category creator is required")
	}
	return nil
}

// IsValid validates the tag data
func (t *Tag) IsValid() error {
	if t.Name == "" {
		return ErrInvalidInput("tag name is required")
	}
	if t.CreatorID == "" {
		return ErrInvalidInput("tag creator is required")
	}
	return nil
}
