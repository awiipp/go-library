CREATE TABLE loans (
    id UUID PRIMARY KEY,
    book_id UUID NOT NULL REFERENCES books(id),
    user_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    borrowed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    due_at TIMESTAMPTZ NOT NULL,
    returned_at TIMESTAMPTZ
);

CREATE INDEX idx_loans_user_id ON loans(user_id);
CREATE INDEX idx_loans_book_id ON loans(book_id);
