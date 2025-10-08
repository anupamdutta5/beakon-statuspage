// Package handlers provides HTTP request handlers and helper utilities for the Tenant Admin Service.
//
// This file contains reusable helper functions for common handler operations such as
// parameter parsing and standardized response formatting. These helpers reduce code
// duplication and ensure consistent behavior across all API endpoints.
package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

// ParseUUIDParam extracts and parses a UUID from a URL parameter.
//
// This helper eliminates code duplication across handlers that need to parse UUID
// parameters from URL paths. If parsing fails, it automatically sends an error
// response to the client and returns false.
//
// Parameters:
//   - c: The Gin context containing the request
//   - paramName: The name of the URL parameter to parse (e.g., "tenant_id", "id")
//
// Returns:
//   - uuid.UUID: The parsed UUID value (uuid.Nil if parsing failed)
//   - bool: true if parsing succeeded, false if it failed (error response already sent)
//
// Example usage:
//
//	tenantID, ok := ParseUUIDParam(c, "tenant_id")
//	if !ok {
//	    return // Error response already sent
//	}
//	// Use tenantID...
func ParseUUIDParam(c *gin.Context, paramName string) (uuid.UUID, bool) {
	paramValue := c.Param(paramName)
	id, err := uuid.Parse(paramValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid " + paramName,
		})
		return uuid.Nil, false
	}
	return id, true
}

// ParseUintParam extracts and parses an unsigned integer from a URL parameter.
//
// Similar to ParseUUIDParam, this helper standardizes uint parsing across handlers.
// It automatically handles parsing errors and sends appropriate error responses.
//
// Parameters:
//   - c: The Gin context containing the request
//   - paramName: The name of the URL parameter to parse (e.g., "id", "admin_id")
//
// Returns:
//   - uint: The parsed uint value (0 if parsing failed)
//   - bool: true if parsing succeeded, false if it failed (error response already sent)
//
// Example usage:
//
//	adminID, ok := ParseUintParam(c, "id")
//	if !ok {
//	    return // Error response already sent
//	}
//	// Use adminID...
func ParseUintParam(c *gin.Context, paramName string) (uint, bool) {
	paramValue := c.Param(paramName)
	id, err := strconv.ParseUint(paramValue, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid " + paramName,
		})
		return 0, false
	}
	return uint(id), true
}

// RespondWithError sends a standardized JSON error response.
//
// This helper ensures all error responses follow a consistent format across the API.
// It reduces code duplication and makes error handling more maintainable.
//
// Parameters:
//   - c: The Gin context for the HTTP response
//   - statusCode: The HTTP status code (e.g., http.StatusBadRequest, http.StatusInternalServerError)
//   - message: The error message to send to the client
//
// Response format:
//
//	{
//	    "error": "message"
//	}
//
// Example usage:
//
//	RespondWithError(c, http.StatusBadRequest, "Invalid tenant ID")
//	RespondWithError(c, http.StatusInternalServerError, "Database connection failed")
func RespondWithError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"error": message,
	})
}

// RespondWithSuccess sends a standardized JSON success response.
//
// This helper provides a consistent format for successful API responses.
// It can optionally include data in the response payload.
//
// Parameters:
//   - c: The Gin context for the HTTP response
//   - statusCode: The HTTP status code (typically http.StatusOK or http.StatusCreated)
//   - message: A success message describing the operation
//   - data: Optional data to include in the response (can be nil)
//
// Response format (with data):
//
//	{
//	    "message": "Operation successful",
//	    "data": { ... }
//	}
//
// Response format (without data):
//
//	{
//	    "message": "Operation successful"
//	}
//
// Example usage:
//
//	RespondWithSuccess(c, http.StatusOK, "Tenant created successfully", tenant)
//	RespondWithSuccess(c, http.StatusOK, "Tenant deleted successfully", nil)
func RespondWithSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	response := gin.H{
		"message": message,
	}
	if data != nil {
		response["data"] = data
	}
	c.JSON(statusCode, response)
}
