USE adp;

ALTER TABLE `t_sandbox_session`
ADD COLUMN IF NOT EXISTS `f_python_package_index_url` varchar(512) NOT NULL
DEFAULT 'https://pypi.org/simple/'
AFTER `f_completed_at`;
