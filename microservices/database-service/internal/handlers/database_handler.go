// Package handlers provides HTTP handlers for the Database Service.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/enterprise-status/statuspage-database-service/internal/models"
	"github.com/enterprise-status/statuspage-database-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DatabaseHandler handles database-related HTTP requests.
type DatabaseHandler struct {
	service *services.DatabaseService
	logger  *zap.Logger
}

// NewDatabaseHandler creates a new database handler.
func NewDatabaseHandler(service *services.DatabaseService, logger *zap.Logger) *DatabaseHandler {
	return &DatabaseHandler{
		service: service,
		logger:  logger,
	}
}

// HealthCheck handles health check requests.
func (h *DatabaseHandler) HealthCheck(c *gin.Context) {
	h.logger.Info("Health check requested")

	// Check service health
	if err := h.service.Health(c.Request.Context()); err != nil {
		h.logger.Error("Health check failed", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "unhealthy",
			"error":   err.Error(),
			"service": "database-service",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "database-service",
		"version": "1.0.0",
	})
}

// ListDatabases handles listing databases.
func (h *DatabaseHandler) ListDatabases(c *gin.Context) {
	h.logger.Info("Listing databases")

	databases, err := h.service.ListDatabases(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list databases", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list databases",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"databases": databases,
		"count":     len(databases),
	})
}

// CreateDatabase handles creating a new database.
func (h *DatabaseHandler) CreateDatabase(c *gin.Context) {
	var database models.Database
	if err := c.ShouldBindJSON(&database); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Creating database", zap.String("name", database.Name))

	if err := h.service.CreateDatabase(c.Request.Context(), &database); err != nil {
		h.logger.Error("Failed to create database", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create database",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"database": database,
		"message":  "Database created successfully",
	})
}

// GetDatabase handles retrieving a database by ID.
func (h *DatabaseHandler) GetDatabase(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid database ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid database ID",
		})
		return
	}

	h.logger.Info("Getting database", zap.Uint64("database_id", id))

	database, err := h.service.GetDatabase(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get database", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Database not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"database": database,
	})
}

// UpdateDatabase handles updating a database.
func (h *DatabaseHandler) UpdateDatabase(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid database ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid database ID",
		})
		return
	}

	var updates models.Database
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Updating database", zap.Uint64("database_id", id))

	if err := h.service.UpdateDatabase(c.Request.Context(), uint(id), &updates); err != nil {
		h.logger.Error("Failed to update database", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update database",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Database updated successfully",
	})
}

// DeleteDatabase handles deleting a database.
func (h *DatabaseHandler) DeleteDatabase(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid database ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid database ID",
		})
		return
	}

	h.logger.Info("Deleting database", zap.Uint64("database_id", id))

	if err := h.service.DeleteDatabase(c.Request.Context(), uint(id)); err != nil {
		h.logger.Error("Failed to delete database", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete database",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Database deleted successfully",
	})
}

// ListTables handles listing tables in a database.
func (h *DatabaseHandler) ListTables(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid database ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid database ID",
		})
		return
	}

	h.logger.Info("Listing tables", zap.Uint64("database_id", id))

	tables, err := h.service.ListTables(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to list tables", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list tables",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tables": tables,
		"count":  len(tables),
	})
}

// CreateTable handles creating a new table.
func (h *DatabaseHandler) CreateTable(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid database ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid database ID",
		})
		return
	}

	var table models.Table
	if err := c.ShouldBindJSON(&table); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	table.DatabaseID = uint(id)

	h.logger.Info("Creating table", zap.String("name", table.Name), zap.Uint64("database_id", id))

	if err := h.service.CreateTable(c.Request.Context(), &table); err != nil {
		h.logger.Error("Failed to create table", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create table",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"table":   table,
		"message": "Table created successfully",
	})
}

// GetTable handles retrieving a table by ID.
func (h *DatabaseHandler) GetTable(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid table ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid table ID",
		})
		return
	}

	h.logger.Info("Getting table", zap.Uint64("table_id", id))

	table, err := h.service.GetTable(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get table", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Table not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"table": table,
	})
}

// UpdateTable handles updating a table.
func (h *DatabaseHandler) UpdateTable(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid table ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid table ID",
		})
		return
	}

	var updates models.Table
	if err := c.ShouldBindJSON(&updates); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Updating table", zap.Uint64("table_id", id))

	if err := h.service.UpdateTable(c.Request.Context(), uint(id), &updates); err != nil {
		h.logger.Error("Failed to update table", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update table",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Table updated successfully",
	})
}

// DeleteTable handles deleting a table.
func (h *DatabaseHandler) DeleteTable(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid table ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid table ID",
		})
		return
	}

	h.logger.Info("Deleting table", zap.Uint64("table_id", id))

	if err := h.service.DeleteTable(c.Request.Context(), uint(id)); err != nil {
		h.logger.Error("Failed to delete table", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete table",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Table deleted successfully",
	})
}

// ExecuteQuery handles executing a SQL query.
func (h *DatabaseHandler) ExecuteQuery(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid database ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid database ID",
		})
		return
	}

	var request struct {
		Query string `json:"query" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Executing query", zap.Uint64("database_id", id), zap.String("query", request.Query))

	result, err := h.service.ExecuteQuery(c.Request.Context(), uint(id), request.Query)
	if err != nil {
		h.logger.Error("Failed to execute query", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to execute query",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": result,
	})
}

// ExecuteTransaction handles executing a database transaction.
func (h *DatabaseHandler) ExecuteTransaction(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid database ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid database ID",
		})
		return
	}

	var request struct {
		Operations []*models.TransactionOperation `json:"operations" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	h.logger.Info("Executing transaction", zap.Uint64("database_id", id), zap.Int("operation_count", len(request.Operations)))

	if err := h.service.ExecuteTransaction(c.Request.Context(), uint(id), request.Operations); err != nil {
		h.logger.Error("Failed to execute transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to execute transaction",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Transaction executed successfully",
	})
}

// GetDatabaseStats handles getting database statistics.
func (h *DatabaseHandler) GetDatabaseStats(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Error("Invalid database ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid database ID",
		})
		return
	}

	h.logger.Info("Getting database statistics", zap.Uint64("database_id", id))

	stats, err := h.service.GetDatabaseStats(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get database statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get database statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// Placeholder handlers for remaining endpoints
func (h *DatabaseHandler) GetData(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *DatabaseHandler) CreateData(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *DatabaseHandler) UpdateData(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *DatabaseHandler) DeleteData(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *DatabaseHandler) CreateBackup(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *DatabaseHandler) ListBackups(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *DatabaseHandler) GetBackup(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *DatabaseHandler) RestoreBackup(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *DatabaseHandler) ListMigrations(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *DatabaseHandler) RunMigration(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *DatabaseHandler) RollbackMigration(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *DatabaseHandler) GetTableStats(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

