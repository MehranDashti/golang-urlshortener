CREATE TABLE IF NOT EXISTS urls (
    id           VARCHAR(36)  NOT NULL PRIMARY KEY,
    user_id      VARCHAR(36)  NOT NULL,
    original_url TEXT         NOT NULL,
    short_code   VARCHAR(10)  NOT NULL,
    clicks       INT          NOT NULL DEFAULT 0,
    expires_at   DATETIME(3)  NULL,
    created_at   DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at   DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
                              ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE INDEX idx_urls_short_code (short_code),
    INDEX        idx_urls_user_id    (user_id),
    CONSTRAINT   fk_urls_user_id
        FOREIGN KEY (user_id) REFERENCES users(id)
        ON DELETE CASCADE
);