package errors

import "errors"

var (
	ErrDoesNotExist         = errors.New("doesn't exist")
	ErrFailedToRollback     = errors.New("failed to rollback transaction")
	ErrFailedToExecuteQuery = errors.New("failed to execute query")
	ErrAlreadyExists        = errors.New("already exists")
	ErrFailedToBeginTx      = errors.New("failed to begin transaction")
	ErrFailedToRollbackTx   = errors.New("failed to rollback transaction")
	ErrFailedToCommitTx     = errors.New("failed to commit transaction")
)
