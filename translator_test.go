// 2024/12/2 Bin Liu <bin.liu@enmotech.com>

package translator

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/toolkits/file"
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
	"sql_translator/metadata"
	"sql_translator/models"
	"sql_translator/rules"
	"testing"
	textTemplate "text/template"
)

func Test_PrintRule(t *testing.T) {
	r := &rules.SybaseToPostgresRules{}

	ruleList := r.GetRules(models.DBTypeSybase, models.DBTypePostgreSQL)

	b, err := yaml.Marshal(ruleList)

	if err != nil {
		t.Fatal(err)
	}

	_, err = file.WriteBytes("./docs/sybase_to_postgresql.yaml", b)

	if err != nil {
		t.Fatal(err)
	}
}

func TestNewSQLTranslator(t *testing.T) {
	type args struct {
		apiKey           string
		metadataProvider metadata.Provider
		rules            rules.RuleProvider
	}
	tests := []struct {
		name    string
		args    args
		want    *SQLTranslator
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSQLTranslator(tt.args.apiKey, tt.args.metadataProvider, tt.args.rules)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSQLTranslator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSQLTranslator() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLTranslator_AnalyzeSQL(t1 *testing.T) {
	type fields struct {
		llm              llms.Model
		metadataProvider metadata.Provider
		rules            rules.RuleProvider
		maxTokens        int
	}
	type args struct {
		ctx context.Context
		sql string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *models.AnalyzeResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &SQLTranslator{
				llm:              tt.fields.llm,
				metadataProvider: tt.fields.metadataProvider,
				rules:            tt.fields.rules,
				maxTokens:        tt.fields.maxTokens,
			}
			got, err := t.AnalyzeSQL(tt.args.ctx, tt.args.sql)
			if (err != nil) != tt.wantErr {
				t1.Errorf("AnalyzeSQL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("AnalyzeSQL() got = %v, want %v", got, tt.want)
			}
		})
	}
}

const (
	StatusSucceed = "succeed"
	StatusFailed  = "failed"
)

type Result struct {
	Result []*ResultDetail `json:"result"`
	SQL    string
}

type ResultDetail struct {
	SQL               string `json:"sql"`
	CreateStatus      string `json:"status"`
	CreateError       string `json:"error"`
	ExecStatus        string
	ExecError         string `json:"error"`
	RulePrompt        string
	TranslationPrompt string
}

func TestSQLTranslator_SybaseToPostgreSQL_TranslateSQL(t *testing.T) {
	llm, err := openai.New(openai.WithModel("gpt-4o-mini"))
	// llm, err := ollama.New(ollama.WithModel("llama3.1"))
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("postgres", "postgres://postgres:root@127.0.0.1:5432/mtk?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}
	db.Exec("set search_path TO a1,public")
	// ruleProvider := rules.NewRuleRegistry()
	// ruleProvider.Register(models.DBTypeSybase, models.DBTypePostgreSQL, rules.NewSybaseToPostgresRules())
	//
	translator := &SQLTranslator{
		llm:              llm,
		metadataProvider: nil,
		rules:            &rules.SybaseToPostgresRules{},
	}

	tests := []struct {
		name    string
		sql     string
		want    string
		wantErr bool
		count   int
	}{
		{
			name: "variable name conflicts with keyword",
			sql: `CREATE PROCEDURE test_proc
    @current_time int,
    @timestamp int
AS
BEGIN
    SELECT @current_time, @timestamp;
END`,
			want: `CREATE OR REPLACE PROCEDURE test_proc(
    current_time_var int,
    timestamp_val int
) AS $$
BEGIN
    SELECT current_time_var, timestamp_val;
END; $$ LANGUAGE plpgsql;`,
			wantErr: false,
		},
		{
			name: "multiple variable name conflicts",
			sql: `DECLARE
    @current_time int;
    @current_time_var int;
    @timestamp int;
    @timestamp_val int;`,
			want: `DECLARE
    current_time_var1 int;
    current_time_var int;
    timestamp_val1 int;
    timestamp_val int;`,
			wantErr: false,
		},
		{
			/*
				DO
				$$
				DECLARE
				    validation_status char(30);
				    user_status char(30);
				   user_id char(30) := '1';
				BEGIN
				    call matrix_authProfileAccess(validation_status,user_status,user_id);
				    RAISE NOTICE '% : %', validation_status, user_status;
				end;
				$$;
			*/
			name: "matrix_authProfileAccess",
			sql: `create proc matrix_authProfileAccess(
	@validation_status char(30) output, 
	@user_status  char(30) output,
	@user_id char(30)) as 
begin 
	declare @user_profile char(10)
	select @user_profile = profile_id from dc_web_profile where profile_id 
	  in (select group_id from dc_web_profile_member where member= @user_id)

	if exists (select '1' from dc_web_profile_access where profile_id= @user_profile) 
	begin
		/* do time check */
		declare @access_begintime datetime
		declare @access_endtime datetime
		declare @current_time datetime
		declare @var_current_time datetime
		select @access_begintime=access_st_time, @access_endtime=access_end_time 
		  from dc_web_profile_access where profile_id= @user_profile

		select @current_time=getdate()
		
		select @var_current_time=getdate()
		
		if (datepart(hh, @current_time) >= datepart(hh, @access_begintime)
		  and datepart(hh, @current_time) <= datepart(hh, @access_endtime))
		begin
			if (datepart(hh, @current_time) = datepart(hh, @access_begintime))
			begin
				if (not(datepart(mi, @current_time) > datepart(mi, @access_begintime)))
				begin
					select @validation_status="INVALIDPROFILEACCESSTIME", 
					  @user_status="UNAUTHENTICATED"
					return									
				end
			end
			if (datepart(hh, @current_time) = datepart(hh, @access_endtime))
			begin
				if (not(datepart(mi, @current_time) < datepart(mi, @access_endtime)))
				begin
					select @validation_status="INVALIDPROFILEACCESSTIME", 
					  @user_status="UNAUTHENTICATED"
					return									
				end
			end
		end
		else 
		begin
			select @validation_status="INVALIDPROFILEACCESSTIME", 
			  @user_status="UNAUTHENTICATED"
			return
		end
	end
end`,
			want:    "",
			wantErr: false,
			count:   10,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			if tt.count == 0 {
				tt.count = 1
			}
			var r = &Result{
				SQL:    tt.sql,
				Result: make([]*ResultDetail, 0),
			}
			for i := 0; i < tt.count; i++ {
				result, err := translator.TranslateSQL(ctx, tt.sql, models.DBTypeSybase, models.DBTypePostgreSQL)
				if (err != nil) != tt.wantErr {
					t1.Errorf("TranslateSQL() error = %v, wantErr %v", err, tt.wantErr)
					continue
				}
				if result == nil {
					continue
				}
				if result.TargetSQL == "" {
					continue
				}
				res := &ResultDetail{
					SQL: result.TargetSQL,
				}
				r.Result = append(r.Result, res)
				res.CreateStatus = StatusSucceed
				res.TranslationPrompt = result.TranslationPrompt
				res.RulePrompt = result.RulePrompt
				_, err = db.Exec(result.TargetSQL)

				if (err != nil) != tt.wantErr {
					res.CreateError = err.Error()
					res.CreateStatus = StatusFailed
					continue
				}
				_, err = db.Exec(`DO
				$$
				DECLARE
				    validation_status char(30);
				    user_status char(30);
				   	user_id char(30) := '1';
				BEGIN
				    call matrix_authProfileAccess(validation_status,user_status,user_id);
				    RAISE NOTICE '% : %', validation_status, user_status;
				end;
				$$;`)
				res.ExecStatus = StatusSucceed
				if (err != nil) != tt.wantErr {
					res.ExecStatus = StatusFailed
					res.ExecError = err.Error()
					continue
				}
			}
			err = printResult(r, "./docs/"+tt.name+".md")
			if err != nil {
				t.Error(err)
			}
		})
	}
}

var (
	tempText = `# Report

Source SQL    :` + "\n\n```sql\n{{.SQL}}\n```\n" + `

## Summary

| No |     Create     |      Exec       |
|----|----------------|-----------------|
{{- range $i, $o := .Result }}
{{printf "|[%v](#%v)|%-19s|%-19s|" $i $i $o.CreateStatus $o.ExecStatus}}
{{- end }}

{{range $i, $o := .Result }}

## {{$i}}

Create Error  :` + "\n\n```sql\n{{$o.CreateError}}\n```\n" + `
Exec   Error  :` + "\n\n```sql\n{{$o.ExecError}}\n```\n" + `
Target SQL    :` + "\n\n```sql\n{{$o.SQL}}\n```\n" + `
### Rule Prompt
	
` + "```text\n{{$o.RulePrompt}}\n```\n" +
		`
### Tran Prompt

` + "```text\n{{$o.TranslationPrompt}}\n```\n" +
		`
{{- end }}
`
)

func printResult(result *Result, fileName string) error {
	f, err := os.Create(fileName) /* #nosec */
	if err != nil {
		return err
	}
	defer f.Close() /* #nosec */
	tmpl, err := textTemplate.New("textReport").Parse(tempText)
	if err != nil {
		return err
	}
	err = tmpl.Execute(f, result)
	if err != nil {
		return err
	}
	return nil
}
