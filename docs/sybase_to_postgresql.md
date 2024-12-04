# Sybase To PostgreSQL


## Procedure

- 出参 改成 inout

    ```sql
    create proc example_output(
        @validation_status char(30) output, 
        @user_status  char(30) output,
        @user_id char(30)) as 
    begin 
    end;
    ```

    ```sql
    CREATE OR REPLACE PROCEDURE example_output(
        INOUT validation_status char(30), 
        INOUT user_status char(30),
        user_id char(30)
    ) AS $$
    DECLARE
    BEGIN
    END; $$ LANGUAGE plpgsql;
    ```