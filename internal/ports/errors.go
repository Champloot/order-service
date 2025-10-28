package ports

import "errors"

var (
	ErrOrderNotFound = errors.New("Order not found")
	ErrTxBeginFailed = errors.New("Transaction begin failed")
	ErrTxCommitFailed = errors.New("Transaction commit failed")
)