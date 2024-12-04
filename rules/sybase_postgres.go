package rules

import (
	"sql_translator/models"
	"strings"
)

// SybaseToPostgresRules defines translation rules from Sybase to PostgreSQL
type SybaseToPostgresRules struct{}

func (r *SybaseToPostgresRules) getTypeRules() []*TranslationRule {
	return []*TranslationRule{
		NewTranslationRuleFromTemplate(DataTypeRuleTemplate, "Convert Sybase data types to PostgreSQL data types").
			WithExample("datetime", "timestamp", "Convert datetime to timestamp").
			WithExample("text", "text", "Text type remains same").
			WithExample("money", "decimal(19,4)", "Convert money to decimal"),
	}
}

func (r *SybaseToPostgresRules) getSQLRules() []*TranslationRule {
	return []*TranslationRule{
		// 		NewTranslationRuleFromTemplate(BasicSQLRuleTemplate, "Basic SQL syntax conversions").
		// 			WithExample(
		// 				"SELECT TOP 10 * FROM employees",
		// 				"SELECT * FROM employees LIMIT 10",
		// 				"Convert TOP to LIMIT",
		// 			).
		// 			WithExample(
		// 				"SELECT GETDATE()",
		// 				"SELECT CURRENT_TIMESTAMP",
		// 				"Convert GETDATE to CURRENT_TIMESTAMP",
		// 			).
		// 			WithExample(
		// 				"SELECT DATEADD(day, 1, hire_date)",
		// 				"SELECT hire_date + INTERVAL '1 day'",
		// 				"Convert DATEADD to interval arithmetic",
		// 			),
		// 		NewTranslationRuleFromTemplate(RecursiveSQLRuleTemplate, "Convert recursive query syntax").
		// 			WithPriority(30).
		// 			WithExample(
		// 				`WITH RECURSIVE hierarchy AS (
		//     SELECT id, parent_id, name, 0 AS level
		//     FROM employees
		//     WHERE parent_id IS NULL
		//     UNION ALL
		//     SELECT e.id, e.parent_id, e.name, h.level + 1
		//     FROM employees e, hierarchy h
		//     WHERE e.parent_id = h.id
		// )`,
		// 				`WITH RECURSIVE hierarchy AS (
		//     SELECT id, parent_id, name, 0 AS level
		//     FROM employees
		//     WHERE parent_id IS NULL
		//     UNION ALL
		//     SELECT e.id, e.parent_id, e.name, h.level + 1
		//     FROM employees e
		//     INNER JOIN hierarchy h ON e.parent_id = h.id
		// )`,
		// 				"Convert recursive query and join syntax",
		// 			),
		// 		NewTranslationRuleFromTemplate(JoinSQLRuleTemplate, "Convert join syntax").
		// 			WithPriority(40).
		// 			WithExample(
		// 				`SELECT e.*, d.*
		// FROM employees e, departments d
		// WHERE e.dept_id *= d.id`,
		// 				`SELECT e.*, d.*
		// FROM employees e
		// LEFT JOIN departments d ON e.dept_id = d.id`,
		// 				"Convert Sybase outer join (*=) to ANSI LEFT JOIN",
		// 			),
	}
}

func (r *SybaseToPostgresRules) getFunctionRules() []*TranslationRule {
	return []*TranslationRule{
		// NewTranslationRuleFromTemplate(BuiltinFuncRuleTemplate, "Convert built-in function calls").
		// 	WithExample(
		// 		"ISNULL(column_name, 0)",
		// 		"COALESCE(column_name, 0)",
		// 		"Convert ISNULL to COALESCE",
		// 	).
		// 	WithExample(
		// 		"CONVERT(varchar(10), hire_date, 120)",
		// 		"TO_CHAR(hire_date, 'YYYY-MM-DD')",
		// 		"Convert CONVERT to TO_CHAR for dates",
		// 	).
		// 	WithExample(
		// 		"CHARINDEX('find', text)",
		// 		"POSITION('find' IN text)",
		// 		"Convert CHARINDEX to POSITION",
		// 	),
		NewTranslationRuleFromTemplate(DateFuncRuleTemplate, "Convert date/time functions").
			WithExample(
				"GETDATE()",
				"CURRENT_TIMESTAMP",
				"Convert GETDATE to CURRENT_TIMESTAMP",
				// ).
				// WithExample(
				// 	"DATEADD(month, 1, hire_date)",
				// 	"hire_date + INTERVAL '1 month'",
				// 	"Convert DATEADD to INTERVAL addition",
				// ).
				// WithExample(
				// 	"DATEDIFF(day, start_date, end_date)",
				// 	"(end_date - start_date)::integer",
				// 	"Convert DATEDIFF to date subtraction",
			),
	}
}

func (r *SybaseToPostgresRules) getProcedureRules() []*TranslationRule {
	return []*TranslationRule{
		NewTranslationRuleFromTemplate(ProcSyntaxRuleTemplate, "Convert stored procedure syntax and structure").
			WithDescription(`Handles the following Sybase to PostgreSQL stored procedure conversions:
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
	- Nested IF statements with ELSE branches`).
			WithExample(
				`CREATE PROCEDURE update_user_status
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
END`,
				`CREATE OR REPLACE PROCEDURE update_user_status(
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
        SELECT col1, col2 INTO var1, var2 FROM employees;
        
        IF role = 'ADMIN' THEN
			validation_status := 'VALID';
			user_status := 'AUTH';
            access := 'FULL';
        ELSE
			validation_status := 'INVALID';
			user_status := 'UNAUTH';
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
$$ LANGUAGE plpgsql;`,
				"Convert stored procedure structure, parameters, variables and control flow",
			),
		// 		NewTranslationRuleFromTemplate(ProcFlowRuleTemplate, "Convert control flow and loop structures").
		// 			WithDescription(`Handles the following control flow conversions:
		// 1. Loop structures:
		//    - WHILE loops with BREAK/CONTINUE
		//    - Cursor-based loops
		//    - Loop variable declarations and updates
		//    - Loop condition evaluations
		//
		// 2. Flow control:
		//    - GOTO statements to labeled blocks
		//    - RETURN with/without values
		//    - CASE statements and expressions
		//    - Conditional execution paths
		//
		// 3. Cursor handling:
		//    - Cursor declarations and options
		//    - Fetch operations and status checks
		//    - Dynamic cursor management
		//    - Result set processing`).
		// 			WithExample(
		// 				`DECLARE cur CURSOR FOR
		//     SELECT id, name FROM users
		//     WHERE status = 'ACTIVE'
		//
		// OPEN cur
		// FETCH NEXT FROM cur INTO @id, @name
		//
		// WHILE @@FETCH_STATUS = 0
		// BEGIN
		//     IF @name IS NULL
		//         CONTINUE
		//
		//     UPDATE profiles SET last_seen = GETDATE() WHERE user_id = @id
		//
		//     IF @@ERROR <> 0
		//         BREAK
		//
		//     FETCH NEXT FROM cur INTO @id, @name
		// END
		//
		// CLOSE cur
		// DEALLOCATE cur`,
		// 				`DECLARE
		//     cur CURSOR FOR
		//         SELECT id, name FROM users
		//         WHERE status = 'ACTIVE';
		// BEGIN
		//     FOR record IN cur LOOP
		//         IF record.name IS NULL THEN
		//             CONTINUE;
		//         END IF;
		//
		//         UPDATE profiles SET last_seen = CURRENT_TIMESTAMP WHERE user_id = record.id;
		//
		//         IF NOT FOUND THEN
		//             EXIT;
		//         END IF;
		//     END LOOP;
		// END;`,
		// 				"Convert cursor loops and flow control",
		// 			),
		// 		NewTranslationRuleFromTemplate(ProcErrorRuleTemplate, "Convert error handling and exception blocks").
		// 			WithDescription(`Handles the following error handling conversions:
		// 1. Exception handling:
		//    - TRY-CATCH blocks to BEGIN-EXCEPTION
		//    - Error message handling and propagation
		//    - Custom error codes and states
		//    - Transaction rollback in error blocks
		//
		// 2. Error status:
		//    - @@ERROR to SQLSTATE
		//    - Error code mapping
		//    - Error message formatting
		//
		// 3. Error raising:
		//    - RAISERROR to RAISE EXCEPTION
		//    - Dynamic error messages
		//    - Error severity levels`).
		// 			WithExample(
		// 				`BEGIN TRY
		//     INSERT INTO users (id, name) VALUES (@id, @name)
		//     IF @@ERROR <> 0
		//         RAISERROR('Insert failed', 16, 1)
		// END TRY
		// BEGIN CATCH
		//     SELECT @ErrorMessage = ERROR_MESSAGE()
		//     SELECT @ErrorSeverity = ERROR_SEVERITY()
		//     RAISERROR(@ErrorMessage, @ErrorSeverity, 1)
		//     ROLLBACK TRANSACTION
		// END CATCH`,
		// 				`BEGIN
		//     INSERT INTO users (id, name) VALUES (id, name);
		//     IF FOUND THEN
		//         NULL;
		//     ELSE
		//         RAISE EXCEPTION 'Insert failed';
		//     END IF;
		// EXCEPTION WHEN OTHERS THEN
		//     GET STACKED DIAGNOSTICS error_message = MESSAGE_TEXT,
		//                           error_detail = PG_EXCEPTION_DETAIL,
		//                           error_hint = PG_EXCEPTION_HINT;
		//     RAISE EXCEPTION '%', error_message
		//         USING DETAIL = error_detail,
		//               HINT = error_hint;
		//     ROLLBACK;
		// END;`,
		// 				"Convert error handling blocks and statements",
		// 			),
	}
}

// GetRules returns all translation rules from Sybase to PostgreSQL
func (r *SybaseToPostgresRules) GetRules(sourceDB, targetDB string) []*TranslationRule {
	if !strings.EqualFold(sourceDB, models.DBTypeSybase) || !strings.EqualFold(targetDB, models.DBTypePostgreSQL) {
		return nil
	}

	rules := make([]*TranslationRule, 0)
	rules = append(rules, r.getTypeRules()...)
	rules = append(rules, r.getSQLRules()...)
	rules = append(rules, r.getFunctionRules()...)
	rules = append(rules, r.getProcedureRules()...)
	return rules
}

// GetRulesByCategory returns rules filtered by category
func (r *SybaseToPostgresRules) GetRulesByCategory(sourceDB, targetDB string, category RuleCategory) []*TranslationRule {
	if !strings.EqualFold(sourceDB, models.DBTypeSybase) || !strings.EqualFold(targetDB, models.DBTypePostgreSQL) {
		return nil
	}

	switch category {
	case TypeRule:
		return r.getTypeRules()
	case SQLRule:
		return r.getSQLRules()
	case FunctionRule:
		return r.getFunctionRules()
	case ProcedureRule:
		return r.getProcedureRules()
	default:
		return nil
	}
}
