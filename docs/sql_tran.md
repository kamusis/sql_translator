# SQL Tran



```shell
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
```


## 提取相关对象名

```text
Analyze the following SQL and identify all table and procedure references.
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
{{.SQL}}
```

## 查询相关对象信息

- 表结构信息
  - 表名
  - 列名
  - 列类型
  - 列默认值
  - 列非空
  - 约束？
  - 索引？
- 存储过程函数
  - 参数名
  - 参数类型
  - 默认值
  - 出参和入参
  - 重载函数怎么处理?

## 构造Rule 系统prompt

```text
Understand these {{.SourceType}} to {{.TargetType}} translation rules:
{{range .Rules}}
[{{.Category}} - Priority:{{.Priority}}] {{.Description}}
{{range .Examples}}• {{.Source}} → {{.Target}} ({{.Explanation}})
{{end}}{{end}}
Respond "Understood" if processed.
```

完整的rule prompt

```shell
Understand these sybase to postgresql translation rules:

[Type - Priority:100] 
• datetime → timestamp (Convert datetime to timestamp)
• text → text (Text type remains same)
• money → decimal(19,4) (Convert money to decimal)

[Function - Priority:90] 
• GETDATE() → CURRENT_TIMESTAMP (Convert GETDATE to CURRENT_TIMESTAMP)

[Procedure - Priority:85] Handles the following Sybase to PostgreSQL stored procedure conversions:
1. Procedure Structure:
	- Change the parameter declaration output to inout
	- Add language specification (LANGUAGE plpgsql)

2. Variable Declarations:
	- Extract variable declarations to the DECLARE section

3. Variable Assignments:
	- Use SELECT INTO only when assigning table column values to variables (SELECT @var = table.col -> SELECT table.col INTO var)
	- Use := operator for all other assignment scenarios:
		- Direct value assignments (SELECT @var = 100 -> var := 100)
		- Expression assignments (SELECT @var = col1 + col2 -> var := col1 + col2)
		- Multiple variable assignments (SELECT @var = 1, @var2 = 2 -> var := 1; var2 := 2)

4. Control Flow:
	- IF-BEGIN-END blocks to IF-THEN-END IF
	- Nested IF statements with ELSE branches
• CREATE PROCEDURE update_user_status
    @user_id INT,
    @status VARCHAR(10) OUTPUT,
    @last_updated DATETIME OUTPUT
AS
BEGIN
    DECLARE @count INT
    DECLARE @role VARCHAR(20)
    DECLARE @validation_status VARCHAR(20)
    DECLARE @user_status VARCHAR(20)

    SELECT @count = COUNT(*) FROM users WHERE id = @user_id

	IF EXISTS (SELECT 1 FROM users WHERE id = @user_id)
	BEGIN
		SELECT @var1 = col1, @var2 = col2 FROM employees
	
		IF @role = 'ADMIN'
		BEGIN
			SELECT @validation_status = "VALID", @user_status = "AUTH"
			SELECT @access = 'FULL'
			RETURN
		END
		ELSE
		BEGIN
			SELECT @validation_status = "INVALID", @user_status = "UNAUTH"
			SELECT @access = 'LIMITED'
			RETURN
		END
	END
	ELSE
	BEGIN
		SELECT @status = 'INACTIVE'
		SELECT @access = 'NONE'
		SELECT @current_time = GETDATE()
		RETURN
	END

    SELECT @status = 'NOT_FOUND'
    SELECT @last_updated = GETDATE()
    RETURN
END → CREATE OR REPLACE PROCEDURE update_user_status(
    user_id INT,
    INOUT status VARCHAR(10),
    INOUT last_updated TIMESTAMP
) AS $$
DECLARE
    count INTEGER;
    role VARCHAR(20);
    validation_status VARCHAR(20);
    user_status VARCHAR(20);
BEGIN
    SELECT COUNT(*) INTO count FROM users WHERE id = user_id;

    IF EXISTS (SELECT 1 FROM users WHERE id = user_id) THEN
        -- Table query uses SELECT INTO
        SELECT col1, col2 INTO var1, var2 FROM employees;
        
        IF role = 'ADMIN' THEN
            -- Multiple assignments use SELECT INTO
            SELECT 'VALID', 'AUTH' INTO validation_status, user_status;
            access := 'FULL';
        ELSE
            SELECT 'INVALID', 'UNAUTH' INTO validation_status, user_status;
            access := 'LIMITED';
			return
        END IF;
    ELSE
        status := 'INACTIVE';
        access := 'NONE';
        current_time_var := CURRENT_TIMESTAMP;
    END IF;

    status := 'NOT_FOUND';
    last_updated := CURRENT_TIMESTAMP;
    RETURN;
END;
$$ LANGUAGE plpgsql; (Convert stored procedure structure, parameters, variables and control flow)

Respond "Understood" if processed.
```

- Sybase output 出参改成 PostgreSQL inout
- 存储过程里变量名和系统保留字冲突.

[Sybase To PostgreSQL](./sybase_to_postgresql.yaml)

## 构造转换prompt

```text
Using the previously provided translation rules from {{.SourceType}} to {{.TargetType}},
please translate the following SQL. Follow these steps STRICTLY in order:

1. First and most importantly, scan the ENTIRE SQL for ALL variable names (including those in declarations, parameters, assignments, and conditions)
2. For each variable, check if it conflicts with any PostgreSQL keywords (like 'current_time', 'timestamp', 'date', etc.)
3. If a variable conflicts:
   - First try adding '_var' suffix
   - If that name exists, add numeric suffix (_var1, _var2, etc.)
   - Ensure to replace ALL occurrences of the variable consistently
4. Only after resolving ALL variable name conflicts, proceed with other translation rules
5. Return only the translated SQL statement

Source SQL:
{{.SQL}}

Available Metadata:
{{.Metadata}}
```

完整的转换prompt

```shell
Using the previously provided translation rules from sybase to postgresql,
please translate the following SQL. Follow these steps STRICTLY in order:

1. First and most importantly, scan the ENTIRE SQL for ALL variable names (including those in declarations, parameters, assignments, and conditions)
2. For each variable, check if it conflicts with any PostgreSQL keywords (like 'current_time', 'timestamp', 'date', etc.)
3. If a variable conflicts:
   - First try adding '_var' suffix
   - If that name exists, add numeric suffix (_var1, _var2, etc.)
   - Ensure to replace ALL occurrences of the variable consistently
4. Only after resolving ALL variable name conflicts, proceed with other translation rules
5. Return only the translated SQL statement

Source SQL:
create proc matrix_authProfileAccess(
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
end

Available Metadata:
Tables:

Procedures:
```
## Test

[matrix_authProfileAccess](./matrix_authProfileAccess.md)