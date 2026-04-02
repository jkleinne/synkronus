package main

import "errors"

// ErrOperationAborted indicates the user chose not to proceed with a destructive operation.
var ErrOperationAborted = errors.New("operation aborted by the user")
