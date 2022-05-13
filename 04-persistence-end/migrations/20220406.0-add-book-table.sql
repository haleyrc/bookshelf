CREATE TABLE books (
    id         BIGSERIAL PRIMARY KEY,
    title      TEXT NOT NULL CHECK (length(title) > 0),
    author     TEXT NOT NULL CHECK (length(author) > 0),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

COMMENT ON TABLE books IS 'User-submitted books that can be rated';
COMMENT ON COLUMN books.id IS 'A unique identifier for the book';
COMMENT ON COLUMN books.title IS 'The title of the book';
COMMENT ON COLUMN books.author IS 'The author of the book';
COMMENT ON COLUMN books.created_at IS 'The time the book was added to the database';