package translator

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tmc/langchaingo/llms/openai"
	"regexp"
	"sql_translator/models"
	"strings"

	"github.com/tmc/langchaingo/prompts"

	"sql_translator/metadata"
	"sql_translator/rules"

	"github.com/tmc/langchaingo/llms"
)

// SQLTranslator handles SQL translation between different database dialects
type SQLTranslator struct {
	llm                 llms.Model
	metadataProvider    metadata.Provider
	rules               rules.RuleProvider
	maxTokens           int
	maxExamplesPerChunk int // Maximum number of rule examples per chunk
}

// NewSQLTranslator creates a new SQLTranslator instance
func NewSQLTranslator(apiKey string, metadataProvider metadata.Provider, rules rules.RuleProvider) (*SQLTranslator, error) {
	// TODO 传递 llm 接口 支持更多ai模型
	llm, err := openai.New(openai.WithToken(apiKey))
	// llm, err := ollama.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM: %w", err)
	}
	return &SQLTranslator{
		llm:                 llm,
		metadataProvider:    metadataProvider,
		rules:               rules,
		maxTokens:           2000,
		maxExamplesPerChunk: 5, // Default to 5 examples per chunk to avoid token limits
	}, nil
}

const analyzeTemplate = `Analyze the following SQL and identify all table and procedure references.
Please return the result in JSON format with the following structure:
{
    "objects": [
        {
            "type": "table|procedure",
            "owner": "schema_or_owner_name",
            "name": "object_name"
        }
    ]
}

SQL:
{{.SQL}}`

const ruleTemplate = `Understand these {{.SourceType}} to {{.TargetType}} translation rules:
{{range .Rules}}
[{{.Category}} - Priority:{{.Priority}}] {{.Description}}
{{range .Examples}}• {{.Source}} → {{.Target}} ({{.Explanation}})
{{end}}{{end}}
Respond "Understood" if processed.`

const translateTemplate = `Using the previously provided translation rules from {{.SourceType}} to {{.TargetType}},
please translate the following SQL. Follow these steps STRICTLY in order:

1. First and most importantly, scan the ENTIRE SQL for ALL variable names (including those in declarations, parameters, assignments, and conditions)
2. For each variable, check if it conflicts with any PostgreSQL keywords (like 'current_time', 'timestamp', 'date', etc.)
3. If a variable conflicts:
   - First try adding '_var' suffix
   - If that name exists, add numeric suffix (_var1, _var2, etc.)
   - Ensure to replace ALL occurrences of the variable consistently
4. Only after resolving ALL variable name conflicts, proceed with other translation rules
5. Do not use select var := value syntax
6. Keep the original comments in the code and do not add additional comments
7. Return only the translated SQL statement

Source SQL:
{{.SQL}}

Available Metadata:
{{.Metadata}}`

// AnalyzeSQL analyzes a SQL statement for dependencies
func (t *SQLTranslator) AnalyzeSQL(ctx context.Context, sql string) (*models.AnalyzeResponse, error) {
	promptTemp := prompts.NewPromptTemplate(analyzeTemplate, []string{"SQL"})
	s, err := promptTemp.Format(map[string]any{
		"SQL": sql,
	})
	if err != nil {
		return nil, err
	}
	response, err := t.llm.Call(ctx, s, llms.WithMaxTokens(t.maxTokens))
	if err != nil {
		return nil, fmt.Errorf("failed to analyze SQL: %w", err)
	}
	var result models.AnalyzeResponse

	list := ParseResponseJson(response)
	if len(list) == 0 {
		return &result, nil
	}
	for i := range list {
		var temp models.AnalyzeResponse
		if err := json.Unmarshal([]byte(list[i]), &temp); err != nil {
			continue
		}
		result.Objects = append(result.Objects, temp.Objects...)
	}
	return &result, nil
}

func ParseResponseJson(s string) []string {
	if s == "" {
		return nil
	}
	return regexp.MustCompile("(?is)```json\\s+(.*?)```").FindAllString(s, -1)
}

// TranslateSQL translates SQL from one dialect to another
// TODO 考虑SQL的类型. 来加载不同的rule规则.
//
//	如SQL语句就不需要加载存储相关的rule.
func (t *SQLTranslator) TranslateSQL(ctx context.Context, sql, sourceDB, targetDB string) (*models.TranslationResult, error) {
	// Analyze SQL for dependencies
	// TODO

	var metadataContext = &models.MetadataContext{}
	// TODO
	//  Cache 针对SQL语句做Cache.防止二次分析.
	//	但是可能AI第一次没分析对也会导致结构不对
	//	metadata提供里需要考虑Cache. 不需要底层数据库做
	//		在metadata层做
	if t.metadataProvider != nil {
		analysis, err := t.AnalyzeSQL(ctx, sql)
		if err != nil {
			return nil, fmt.Errorf("analysis failed: %w", err)
		}
		for _, obj := range analysis.Objects {
			switch obj.Type {
			case "table":
				metadata, err := t.metadataProvider.GetTableMetadata(ctx, obj)
				if err != nil {
					continue // Skip problematic metadata
				}
				metadataContext.Tables[obj.Name] = metadata
			case "procedure":
				metadata, err := t.metadataProvider.GetProcedureMetadata(ctx, obj)
				if err != nil {
					continue // Skip problematic metadata
				}
				metadataContext.Procedures[obj.Name] = metadata
			}
		}
	}

	var ruleList []*rules.TranslationRule
	if t.rules != nil {
		// Get translation rules
		ruleList = t.rules.GetRules(sourceDB, targetDB)
	}

	promptTemp := prompts.NewPromptTemplate(ruleTemplate, []string{"SourceType", "TargetType", "Rules"})
	rulePrompt, err := promptTemp.Format(map[string]any{
		"SourceType": sourceDB,
		"TargetType": targetDB,
		"Rules":      ruleList,
	})
	// TODO 错误编码
	if err != nil {
		return nil, err
	}
	tranTemp := prompts.NewPromptTemplate(translateTemplate, []string{"SourceType", "TargetType", "SQL", "Metadata"})
	tranPrompt, err := tranTemp.Format(map[string]any{
		"SourceType": sourceDB,
		"TargetType": targetDB,
		"SQL":        sql,
		"Metadata":   metadataContext.String(),
	})
	// TODO 错误编码
	if err != nil {
		return nil, err
	}
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: rulePrompt},
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: tranPrompt},
			},
		},
	}
	response, err := t.llm.GenerateContent(ctx, messages, llms.WithMaxTokens(t.maxTokens),
		llms.WithTemperature(0.1))
	if err != nil {
		return nil, fmt.Errorf("translation failed: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no translation generated")
	}

	sqlText := response.Choices[0].Content
	sqlText = strings.TrimPrefix(sqlText, "```sql")
	sqlText = strings.TrimSuffix(sqlText, "```")
	return &models.TranslationResult{
		SourceSQL:         sql,
		TargetSQL:         strings.TrimSpace(sqlText),
		SourceDB:          sourceDB,
		TargetDB:          targetDB,
		RulePrompt:        rulePrompt,
		TranslationPrompt: tranPrompt,
	}, nil
}
