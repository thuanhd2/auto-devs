package websocket

import "errors"

var (
	// ErrConnectionClosed indicates the connection is closed
	ErrConnectionClosed = errors.New("connection is closed")
	
	// ErrInvalidMessage indicates an invalid message format
	ErrInvalidMessage = errors.New("invalid message format")
	
	// ErrUnauthorized indicates the connection is not authorized
	ErrUnauthorized = errors.New("unauthorized connection")
	
	// ErrInvalidProjectID indicates an invalid project ID
	ErrInvalidProjectID = errors.New("invalid project ID")
	
	// ErrProjectNotFound indicates a project was not found
	ErrProjectNotFound = errors.New("project not found")
	
	// ErrTaskNotFound indicates a task was not found
	ErrTaskNotFound = errors.New("task not found")
	
	// ErrRateLimited indicates the connection is rate limited
	ErrRateLimited = errors.New("rate limited")
	
	// ErrProcessingFailed indicates message processing failed
	ErrProcessingFailed = errors.New("message processing failed")
)