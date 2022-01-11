package errors

import (
	"fmt"
	"strings"
)

const (
	ConfigValidateFailed = "ConfigValidateFailed"
	APIRequestFailed     = "APIRequestFailed"
	NotCompletedError    = "NotCompletedError"
)

type MigrateError struct {
	errorCode string
	message   string
}

func NewAPIRequestFailedError(err error) error {
	return &MigrateError{
		errorCode: APIRequestFailed,
		message:   fmt.Sprintf("[API Request Error]: %s", err),
	}
}

func IsAPIRequestFailedError(err error) bool {
	if e, ok := err.(*MigrateError); ok &&
		(e.ErrorCode() == APIRequestFailed || strings.Contains(strings.ToLower(e.Message()), APIRequestFailed)) {
		return true
	}

	return false
}

func (err *MigrateError) Error() string {
	if IsNotCompleteError(err) {
		return fmt.Sprintf(" Migrate Cube to UHost: %s", err.message)
	}
	return fmt.Sprintf("[ERROR] Migrate Cube to UHost got err: %s", err.message)
}

func (err *MigrateError) ErrorCode() string {
	return err.errorCode
}

func (err *MigrateError) Message() string {
	return err.message
}

func NewConfigValidateFailedError(err error) error {
	return &MigrateError{
		errorCode: ConfigValidateFailed,
		message:   fmt.Sprintf("[ConfigValidateFailed]: %s", err),
	}
}

func IsConfigValidateFailedError(err error) bool {
	if e, ok := err.(*MigrateError); ok &&
		(e.ErrorCode() == ConfigValidateFailed || strings.Contains(strings.ToLower(e.Message()), ConfigValidateFailed)) {
		return true
	}

	return false
}

func NewNotCompletedError(err error) error {
	return &MigrateError{
		errorCode: NotCompletedError,
		message:   fmt.Sprintf("[NotCompleted]: %s", err),
	}
}

func IsNotCompleteError(err error) bool {
	if e, ok := err.(*MigrateError); ok &&
		(e.ErrorCode() == NotCompletedError || strings.Contains(strings.ToLower(e.Message()), NotCompletedError)) {
		return true
	}
	return false
}
