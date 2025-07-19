CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    login VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE sessions (
    session_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    token VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL
);

CREATE TABLE documents (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    mime VARCHAR(100),
    is_file BOOLEAN NOT NULL,
    is_public BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL,
    owner_id INTEGER NOT NULL REFERENCES users(user_id)
);

CREATE TABLE document_data (
    document_id VARCHAR(36) PRIMARY KEY,
    data TEXT NOT NULL
);

CREATE TABLE document_files (
    document_id VARCHAR(36) PRIMARY KEY,
    data BYTEA NOT NULL
);

CREATE TABLE document_grants (
    grant_id SERIAL PRIMARY KEY,
    document_id VARCHAR(36) NOT NULL REFERENCES documents(id),
    user_id INTEGER NOT NULL REFERENCES users(user_id),
    UNIQUE (document_id, user_id)
);

ALTER TABLE sessions ADD FOREIGN KEY (user_id) REFERENCES users(user_id);
ALTER TABLE documents ADD FOREIGN KEY (owner_id) REFERENCES users(user_id);
ALTER TABLE document_data ADD FOREIGN KEY (document_id) REFERENCES documents(id);
ALTER TABLE document_files ADD FOREIGN KEY (document_id) REFERENCES documents(id);
ALTER TABLE document_grants ADD FOREIGN KEY (document_id) REFERENCES documents(id);
ALTER TABLE document_grants ADD FOREIGN KEY (user_id) REFERENCES users(user_id);