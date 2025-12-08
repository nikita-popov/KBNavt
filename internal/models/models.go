package models

type SearchResult struct {
	Path    string `json:"path" example:"notes/todo.md"`
	Excerpt string `json:"excerpt,omitempty" example:"TODO: refactor API"`
}

type FileContent struct {
	Path    string `json:"path" example:"notes/todo.md"`
	Content string `json:"content" example:"# TODO\n- refactor API\n"`
	Type    string `json:"type" example:".md"`
}
