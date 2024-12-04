package rules

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// RuleCategory represents the category of a translation rule
type RuleCategory int

// RuleSubCategory represents the sub-category of a translation rule
type RuleSubCategory int

const (
	// TypeRule 数据类型和变量类型转换规则 (优先级: 100)
	TypeRule RuleCategory = iota + 1
	// SQLRule SQL语句相关转换规则 (优先级: 95)
	SQLRule
	// FunctionRule 函数转换规则 (优先级: 90)
	FunctionRule
	// ProcedureRule 存储过程规则 (优先级: 85)
	ProcedureRule
)

var (
	ruleCategoryToString = map[RuleCategory]string{
		TypeRule:      "Type",
		SQLRule:       "SQL",
		FunctionRule:  "Function",
		ProcedureRule: "Procedure",
	}
	stringToRuleCategory = map[string]RuleCategory{
		"Type":      TypeRule,
		"SQL":       SQLRule,
		"Function":  FunctionRule,
		"Procedure": ProcedureRule,
	}
)

// MarshalJSON implements json.Marshaler interface
func (r RuleCategory) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

// UnmarshalJSON implements json.Unmarshaler interface
func (r *RuleCategory) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if val, ok := stringToRuleCategory[s]; ok {
		*r = val
		return nil
	}
	return fmt.Errorf("invalid RuleCategory: %s", s)
}

// MarshalYAML implements yaml.Marshaler interface
func (r RuleCategory) MarshalYAML() (interface{}, error) {
	return r.String(), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface
func (r *RuleCategory) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	if val, ok := stringToRuleCategory[s]; ok {
		*r = val
		return nil
	}
	return fmt.Errorf("invalid RuleCategory: %s", s)
}

// String returns the string representation of RuleCategory
func (r RuleCategory) String() string {
	if s, ok := ruleCategoryToString[r]; ok {
		return s
	}
	return "Unknown"
}

// GetDefaultPriority returns the default priority for a rule category
func (r RuleCategory) GetDefaultPriority() int {
	switch r {
	case TypeRule:
		return 100
	case SQLRule:
		return 95
	case FunctionRule:
		return 90
	case ProcedureRule:
		return 85
	default:
		return 50
	}
}

// ValidateSubCategory checks if the subcategory is valid for the category
func (r RuleCategory) ValidateSubCategory(sub RuleSubCategory) bool {
	switch r {
	case TypeRule:
		return sub == DataTypeRule || sub == VariableRule
	case SQLRule:
		return sub == BasicSQLRule || sub == RecursiveRule || sub == JoinRule
	case FunctionRule:
		return sub == BuiltinFuncRule || sub == DateFuncRule
	case ProcedureRule:
		return sub == ProcSyntaxRule || sub == ProcFlowRule || sub == ProcErrorRule
	default:
		return false
	}
}

const (
	// DataTypeRule 类型映射子类别
	DataTypeRule RuleSubCategory = iota + 1
	// VariableRule 变量类型映射
	VariableRule

	// BasicSQLRule SQL语句子类别
	BasicSQLRule
	// RecursiveRule 递归查询转换
	RecursiveRule
	// JoinRule 连接语法转换
	JoinRule

	// BuiltinFuncRule 函数转换子类别
	BuiltinFuncRule
	// DateFuncRule 日期函数转换
	DateFuncRule

	// ProcSyntaxRule 存储过程子类别
	ProcSyntaxRule
	// ProcFlowRule 存储过程流程控制
	ProcFlowRule
	// ProcErrorRule 存储过程错误处理
	ProcErrorRule
)

var (
	ruleSubCategoryToString = map[RuleSubCategory]string{
		DataTypeRule:    "DataType",
		VariableRule:    "Variable",
		BasicSQLRule:    "BasicSQL",
		RecursiveRule:   "Recursive",
		JoinRule:        "Join",
		BuiltinFuncRule: "BuiltinFunc",
		DateFuncRule:    "DateFunc",
		ProcSyntaxRule:  "ProcSyntax",
		ProcFlowRule:    "ProcFlow",
		ProcErrorRule:   "ProcError",
	}

	stringToRuleSubCategory = map[string]RuleSubCategory{
		"DataType":    DataTypeRule,
		"Variable":    VariableRule,
		"BasicSQL":    BasicSQLRule,
		"Recursive":   RecursiveRule,
		"Join":        JoinRule,
		"BuiltinFunc": BuiltinFuncRule,
		"DateFunc":    DateFuncRule,
		"ProcSyntax":  ProcSyntaxRule,
		"ProcFlow":    ProcFlowRule,
		"ProcError":   ProcErrorRule,
	}
)

// MarshalJSON implements json.Marshaler interface
func (r *RuleSubCategory) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

// UnmarshalJSON implements json.Unmarshaler interface
func (r RuleSubCategory) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if val, ok := stringToRuleSubCategory[s]; ok {
		r = val
		return nil
	}
	return fmt.Errorf("invalid RuleSubCategory: %s", s)
}

// MarshalYAML implements yaml.Marshaler interface
func (r RuleSubCategory) MarshalYAML() (interface{}, error) {
	return r.String(), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface
func (r RuleSubCategory) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	if val, ok := stringToRuleSubCategory[s]; ok {
		r = val
		return nil
	}
	return fmt.Errorf("invalid RuleSubCategory: %s", s)
}

// String returns the string representation of RuleSubCategory
func (s RuleSubCategory) String() string {
	if str, ok := ruleSubCategoryToString[s]; ok {
		return str
	}
	return "Unknown"
}

// TranslationRule defines a specific translation rule
type TranslationRule struct {
	Category    RuleCategory
	SubCategory RuleSubCategory
	Priority    int
	Name        string
	Description string
	Examples    []*TranslationExample
}

// WithExample adds an example to the translation rule
func (r *TranslationRule) WithExample(source, target, explanation string) *TranslationRule {
	r.Examples = append(r.Examples, &TranslationExample{
		Source:      source,
		Target:      target,
		Explanation: explanation,
	})
	return r
}

// WithDescription adds a description to the translation rule
func (r *TranslationRule) WithDescription(description string) *TranslationRule {
	r.Description = description
	return r
}

// WithCategory sets the category of the translation rule
func (r *TranslationRule) WithCategory(category RuleCategory) *TranslationRule {
	r.Category = category
	return r
}

// WithSubCategory sets the sub-category of the translation rule
func (r *TranslationRule) WithSubCategory(subCategory RuleSubCategory) *TranslationRule {
	r.SubCategory = subCategory
	return r
}

// WithPriority sets the priority of the translation rule
func (r *TranslationRule) WithPriority(priority int) *TranslationRule {
	r.Priority = priority
	return r
}

// NewTranslationRule creates a new translation rule
func NewTranslationRule(name string) *TranslationRule {
	return &TranslationRule{
		Name:     name,
		Examples: make([]*TranslationExample, 0),
	}
}

// NewTranslationRuleFromTemplate creates a new translation rule from a template
func NewTranslationRuleFromTemplate(template RuleTemplate, name string) *TranslationRule {
	return &TranslationRule{
		Category:    template.Category,
		SubCategory: template.SubCategory,
		Priority:    template.Priority,
		Name:        name,
		Examples:    make([]*TranslationExample, 0),
	}
}

// TranslationExample represents an example of rule application
type TranslationExample struct {
	Source      string
	Target      string
	Explanation string
}

// RuleTemplate provides a template for creating translation rules
type RuleTemplate struct {
	Category    RuleCategory
	SubCategory RuleSubCategory
	Priority    int
}

// Common rule templates
var (
	// Type rules (Priority: 100)
	DataTypeRuleTemplate = RuleTemplate{TypeRule, DataTypeRule, 100}
	VariableRuleTemplate = RuleTemplate{TypeRule, VariableRule, 100}

	// SQL rules (Priority: 95)
	BasicSQLRuleTemplate     = RuleTemplate{SQLRule, BasicSQLRule, 95}
	RecursiveSQLRuleTemplate = RuleTemplate{SQLRule, RecursiveRule, 95}
	JoinSQLRuleTemplate      = RuleTemplate{SQLRule, JoinRule, 95}

	// Function rules (Priority: 90)
	BuiltinFuncRuleTemplate = RuleTemplate{FunctionRule, BuiltinFuncRule, 90}
	DateFuncRuleTemplate    = RuleTemplate{FunctionRule, DateFuncRule, 90}

	// Procedure rules (Priority: 85)
	ProcSyntaxRuleTemplate = RuleTemplate{ProcedureRule, ProcSyntaxRule, 85}
	ProcFlowRuleTemplate   = RuleTemplate{ProcedureRule, ProcFlowRule, 85}
	ProcErrorRuleTemplate  = RuleTemplate{ProcedureRule, ProcErrorRule, 85}
)

// RuleProvider defines the interface for providing translation rules
type RuleProvider interface {
	// GetRules returns all translation rules for a specific source and target database
	GetRules(source, target string) []*TranslationRule
	// GetRulesByCategory returns translation rules filtered by category
	GetRulesByCategory(source, target string, category RuleCategory) []*TranslationRule
}

// RuleRegistry manages the registration of rule providers
type RuleRegistry struct {
	providers map[string]RuleProvider // key: "source:target", e.g., "oracle:postgresql"
}

// NewRuleRegistry creates a new rule registry
func NewRuleRegistry() *RuleRegistry {
	return &RuleRegistry{
		providers: make(map[string]RuleProvider),
	}
}

// Register registers a rule provider for a specific source and target database
func (r *RuleRegistry) Register(source, target string, provider RuleProvider) {
	key := source + ":" + target
	r.providers[key] = provider
}

// GetProvider returns the rule provider for a specific source and target database
func (r *RuleRegistry) GetProvider(source, target string) (RuleProvider, bool) {
	key := source + ":" + target
	provider, exists := r.providers[key]
	return provider, exists
}

// ChunkRules splits rules into chunks to avoid token limits
func ChunkRules(rules []TranslationRule, maxExamplesPerChunk int) [][]TranslationRule {
	if len(rules) == 0 {
		return nil
	}

	var chunks [][]TranslationRule
	currentChunk := make([]TranslationRule, 0)
	currentExamples := 0

	for _, rule := range rules {
		exampleCount := len(rule.Examples)
		if currentExamples+exampleCount > maxExamplesPerChunk && len(currentChunk) > 0 {
			// Current chunk is full, start a new one
			chunks = append(chunks, currentChunk)
			currentChunk = make([]TranslationRule, 0)
			currentExamples = 0
		}
		currentChunk = append(currentChunk, rule)
		currentExamples += exampleCount
	}

	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}
