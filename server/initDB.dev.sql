USE bbsgo_db;
SET NAMES utf8mb4;

INSERT INTO t_user (`id`, `username`, `nickname`, `avatar`, `email`, `password`, `status`, `create_time`, `update_time`,
                    `roles`, `description`, `topic_count`, `comment_count`, `score`)
SELECT 1,
       'admin',
       'bbsgo站长',
       '',
       'a@example.com',
       '$2a$10$ofA39bAFMpYpIX/Xiz7jtOMH9JnPvYfPRlzHXqAtLPFpbE/cLdjmS',
       0,
       (UNIX_TIMESTAMP(now()) * 1000),
       (UNIX_TIMESTAMP(now()) * 1000),
       '高管',
       '轻轻地我走了，正如我轻轻的来。',
       0,
       0,
       0
FROM DUAL
WHERE NOT EXISTS(SELECT * FROM `t_user` WHERE `id` = 1);
