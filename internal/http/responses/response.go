package responses

import (
	"encoding/json"
	"net/http"
)

// APIResponse representa a estrutura padronizada de resposta da API
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// SuccessResponse representa uma resposta de sucesso para Swagger
type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Operação realizada com sucesso"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse representa uma resposta de erro para Swagger
type ErrorResponse struct {
	Success bool      `json:"success" example:"false"`
	Message string    `json:"message" example:"Erro na operação"`
	Error   *APIError `json:"error,omitempty"`
}

// CreatedResponse representa uma resposta de criação para Swagger
type CreatedResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Recurso criado com sucesso"`
	Data    interface{} `json:"data,omitempty"`
}

// APIError representa detalhes de erro na resposta
type APIError struct {
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

// WriteJSON escreve uma resposta JSON padronizada
func WriteJSON(w http.ResponseWriter, statusCode int, success bool, message string, data interface{}, err *APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := APIResponse{
		Success: success,
		Message: message,
		Data:    data,
		Error:   err,
	}

	json.NewEncoder(w).Encode(response)
}

// Success escreve uma resposta de sucesso
func Success(w http.ResponseWriter, message string, data interface{}) {
	WriteJSON(w, http.StatusOK, true, message, data, nil)
}

// Created escreve uma resposta de recurso criado
func Created(w http.ResponseWriter, message string, data interface{}) {
	WriteJSON(w, http.StatusCreated, true, message, data, nil)
}

// BadRequest escreve uma resposta de erro de requisição inválida
func BadRequest(w http.ResponseWriter, message string, details string) {
	WriteJSON(w, http.StatusBadRequest, false, message, nil, &APIError{
		Code:    "BAD_REQUEST",
		Details: details,
	})
}

// NotFound escreve uma resposta de recurso não encontrado
func NotFound(w http.ResponseWriter, message string) {
	WriteJSON(w, http.StatusNotFound, false, message, nil, &APIError{
		Code: "NOT_FOUND",
	})
}

// Conflict escreve uma resposta de conflito
func Conflict(w http.ResponseWriter, message string, details string) {
	WriteJSON(w, http.StatusConflict, false, message, nil, &APIError{
		Code:    "CONFLICT",
		Details: details,
	})
}

// InternalError escreve uma resposta de erro interno
func InternalError(w http.ResponseWriter, message string) {
	WriteJSON(w, http.StatusInternalServerError, false, message, nil, &APIError{
		Code: "INTERNAL_ERROR",
	})
}

// TooManyRequests escreve uma resposta de rate limit excedido
func TooManyRequests(w http.ResponseWriter, message string) {
	WriteJSON(w, http.StatusTooManyRequests, false, message, nil, &APIError{
		Code: "RATE_LIMIT_EXCEEDED",
	})
}

// Error500 escreve uma resposta de erro interno (compatibilidade)
func Error500(w http.ResponseWriter, message, code, details string) {
	WriteJSON(w, http.StatusInternalServerError, false, message, nil, &APIError{
		Code:    code,
		Details: details,
	})
}

// Success200 escreve uma resposta de sucesso (200)
func Success200(w http.ResponseWriter, message string, data interface{}) {
	WriteJSON(w, http.StatusOK, true, message, data, nil)
}

// Success201 escreve uma resposta de recurso criado (201)
func Success201(w http.ResponseWriter, message string, data interface{}) {
	WriteJSON(w, http.StatusCreated, true, message, data, nil)
}

// Error400 escreve uma resposta de erro de requisição inválida (400)
func Error400(w http.ResponseWriter, message, code, details string) {
	WriteJSON(w, http.StatusBadRequest, false, message, nil, &APIError{
		Code:    code,
		Details: details,
	})
}

// Error404 escreve uma resposta de recurso não encontrado (404)
func Error404(w http.ResponseWriter, message, code, details string) {
	WriteJSON(w, http.StatusNotFound, false, message, nil, &APIError{
		Code:    code,
		Details: details,
	})
}

// Error409 escreve uma resposta de conflito (409)
func Error409(w http.ResponseWriter, message, code, details string) {
	WriteJSON(w, http.StatusConflict, false, message, nil, &APIError{
		Code:    code,
		Details: details,
	})
}
