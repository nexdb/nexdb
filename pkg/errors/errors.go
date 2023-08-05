package errors

import "strconv"

// ErrorCode is a custom error code type.
type ErrorCode int

func (c ErrorCode) ToString() string {
	return strconv.Itoa(int(c))
}

// Error is a custom error type that implements the error interface.
type Error struct {
	err string
	ErrorCode
}

// Error returns the error message.
func (e *Error) Error() string {
	return e.err
}

// Code returns the error code.
func (e *Error) Code() ErrorCode {
	return e.ErrorCode
}

// New returns a new error.
func New(code ErrorCode) *Error {
	return &Error{
		ErrorCode: code,
		err:       errorMessageFromCode(code),
	}
}

// errorMessageFromCode returns the error message for the given error code.
func errorMessageFromCode(code ErrorCode) string {
	switch code {
	case ErrCollectionNameIsEmpty:
		return "collection name is empty"
	case ErrCollectionNameIsInvalid:
		return "collection name is invalid, must match the follow regex ^[a-z]*$"
	case ErrUnauthorized:
		return "unauthorized"
	case ErrDocumentNotFound:
		return "document not found"
	default:
		return "unknown error"
	}
}

// Error codes.
const (
	// ErrCollectionNameIsEmpty is returned when the collection name is empty.
	ErrCollectionNameIsEmpty ErrorCode = 1000 + iota
	// ErrCollectionNameIsInvalid is returned when the collection name is invalid.
	ErrCollectionNameIsInvalid
)

const (
	// ErrUnauthorized is returned when a user is not authorized to perform an action.
	ErrUnauthorized ErrorCode = 2000 + iota
)

const (
	// ErrDocumentNotFound is returned when a document is not found.
	ErrDocumentNotFound ErrorCode = 3000 + iota
)
