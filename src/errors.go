package main

import "fmt"

type HalcyonError struct {
	Op      string
	Message string
}

func (err HalcyonError) Error() string {
	if err.Op == "" {
		return err.Message
	}
	return fmt.Sprintf("%s: %s", err.Op, err.Message)
}

func fail(op string, format string, args ...any) error {
	return HalcyonError{Op: op, Message: fmt.Sprintf(format, args...)}
}

func require(condition bool, op string, format string, args ...any) error {
	if condition {
		return nil
	}
	return fail(op, format, args...)
}

func firstError(errors ...error) error {
	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}
