SET SCHEMA adp;

ALTER TABLE t_sandbox_session
ADD COLUMN IF NOT EXISTS f_python_package_index_url VARCHAR(512 CHAR)
DEFAULT 'https://pypi.org/simple/';
