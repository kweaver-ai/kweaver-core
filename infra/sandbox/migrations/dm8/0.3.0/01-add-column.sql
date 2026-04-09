USE adp;

DECLARE
    v_column_count INT;
BEGIN
    SELECT COUNT(*)
    INTO v_column_count
    FROM USER_TAB_COLUMNS
    WHERE TABLE_NAME = 'T_SANDBOX_SESSION'
      AND COLUMN_NAME = 'F_PYTHON_PACKAGE_INDEX_URL';

    IF v_column_count = 0 THEN
        EXECUTE IMMEDIATE '
            ALTER TABLE t_sandbox_session
            ADD f_python_package_index_url VARCHAR(512 CHAR) NOT NULL
            DEFAULT ''https://pypi.org/simple/''
        ';

        EXECUTE IMMEDIATE '
            COMMENT ON COLUMN t_sandbox_session.f_python_package_index_url
            IS ''Python软件包仓库地址''
        ';
    END IF;
END;
/
