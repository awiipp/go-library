package mocks

import "context"

type FakeTransactor struct {
	ForceError error
}

// FakeTransactor immediate execution closure without an actual transaction database.
func (f *FakeTransactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if f.ForceError != nil {
		return f.ForceError
	}

	return fn(ctx)
}
