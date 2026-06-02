package dto

type CategoryResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Icon      string `json:"icon,omitempty"`
	Type      string `json:"type"`
	IsDefault bool   `json:"is_default"`
	IsActive  bool   `json:"is_active"`
	SortOrder int    `json:"sort_order"`
}

type CreateCategoryRequest struct {
	Name string  `json:"name" validate:"required,min=1,max=100"`
	Icon *string `json:"icon"`
	Type string  `json:"type" validate:"required,oneof=fixed variable debt want"`
}
