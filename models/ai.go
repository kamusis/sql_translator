package models

// TranslationResult contains the result of SQL translation
type TranslationResult struct {
	SourceSQL         string
	TargetSQL         string
	SourceDB          string
	TargetDB          string
	Tables            []string
	Procedures        []string
	HasDependencies   bool
	TableDetails      map[string]*TableMetadata
	ProcDetails       map[string]*ProcedureMetadata
	Comments          []string
	RulePrompt        string
	TranslationPrompt string
}

// AnalyzeResponse represents the structured response from AI translation
type AnalyzeResponse struct {
	Objects []*DBObject `json:"objects"`
}

// DBObject represents a database object in AI response
type DBObject struct {
	Type     string `json:"type"`     // "table" 或 "procedure"
	Original string `json:"original"` // 原始引用
	Owner    string `json:"owner"`    // schema或owner
	Name     string `json:"name"`     // 对象名称
}
