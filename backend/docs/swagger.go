package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "swagger": "2.0",
    "info": {
        "description": "Pintuotuo - AI Token Trading Platform API Documentation",
        "title": "Pintuotuo API",
        "contact": {
            "email": "support@pintuotuo.com"
        },
        "version": "1.0"
    },
    "host": "localhost:8080",
    "basePath": "/api/v1",
    "schemes": ["http", "https"],
    "paths": {
        "/users/register": {
            "post": {
                "description": "Register a new user account",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["users"],
                "summary": "Register user",
                "parameters": [
                    {
                        "description": "User registration data",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/RegisterRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {"description": "Created"},
                    "400": {"description": "Bad Request"}
                }
            }
        },
        "/users/login": {
            "post": {
                "description": "Login with email and password",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["users"],
                "summary": "Login user",
                "parameters": [
                    {
                        "description": "Login credentials",
                        "name": "credentials",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/LoginRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {"description": "OK"},
                    "401": {"description": "Unauthorized"}
                }
            }
        },
        "/products": {
            "get": {
                "description": "Get list of products with pagination",
                "produces": ["application/json"],
                "tags": ["products"],
                "summary": "List products",
                "parameters": [
                    {"type": "integer", "description": "Page number", "name": "page", "in": "query"},
                    {"type": "integer", "description": "Items per page", "name": "per_page", "in": "query"},
                    {"type": "string", "description": "Filter by status", "name": "status", "in": "query"}
                ],
                "responses": {
                    "200": {"description": "OK"}
                }
            }
        },
        "/orders": {
            "post": {
                "security": [{"Bearer": []}],
                "description": "Create a new order",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["orders"],
                "summary": "Create order",
                "parameters": [
                    {
                        "description": "Order data",
                        "name": "order",
                        "in": "body",
                        "required": true,
                        "schema": {"$ref": "#/definitions/CreateOrderRequest"}
                    }
                ],
                "responses": {
                    "201": {"description": "Created"},
                    "400": {"description": "Bad Request"},
                    "401": {"description": "Unauthorized"}
                }
            }
        },
        "/proxy/chat": {
            "post": {
                "security": [{"ApiKeyAuth": []}],
                "description": "Proxy chat request to AI providers",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["api-proxy"],
                "summary": "Proxy chat request",
                "parameters": [
                    {
                        "description": "Chat request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {"$ref": "#/definitions/APIProxyRequest"}
                    }
                ],
                "responses": {
                    "200": {"description": "OK"},
                    "401": {"description": "Unauthorized"},
                    "402": {"description": "Insufficient balance"}
                }
            }
        },
        "/health": {
            "get": {
                "description": "Get system health status",
                "produces": ["application/json"],
                "tags": ["health"],
                "summary": "Health check",
                "responses": {
                    "200": {"description": "OK"}
                }
            }
        },
        "/health/ready": {
            "get": {
                "description": "Check if service is ready",
                "produces": ["application/json"],
                "tags": ["health"],
                "summary": "Readiness check",
                "responses": {
                    "200": {"description": "Ready"},
                    "503": {"description": "Not ready"}
                }
            }
        },
        "/health/live": {
            "get": {
                "description": "Check if service is alive",
                "produces": ["application/json"],
                "tags": ["health"],
                "summary": "Liveness check",
                "responses": {
                    "200": {"description": "Alive"}
                }
            }
        },
        "/consumption/records": {
            "get": {
                "security": [{"Bearer": []}],
                "description": "Get consumption records with pagination",
                "produces": ["application/json"],
                "tags": ["consumption"],
                "summary": "Get consumption records",
                "parameters": [
                    {"type": "integer", "name": "page", "in": "query"},
                    {"type": "integer", "name": "page_size", "in": "query"},
                    {"type": "string", "name": "start_date", "in": "query"},
                    {"type": "string", "name": "end_date", "in": "query"},
                    {"type": "string", "name": "provider", "in": "query"}
                ],
                "responses": {
                    "200": {"description": "OK"},
                    "401": {"description": "Unauthorized"}
                }
            }
        },
        "/consumption/stats": {
            "get": {
                "security": [{"Bearer": []}],
                "description": "Get consumption statistics",
                "produces": ["application/json"],
                "tags": ["consumption"],
                "summary": "Get consumption stats",
                "responses": {
                    "200": {"description": "OK"},
                    "401": {"description": "Unauthorized"}
                }
            }
        }
    },
    "definitions": {
        "RegisterRequest": {
            "type": "object",
            "required": ["email", "name", "password"],
            "properties": {
                "email": {"type": "string"},
                "name": {"type": "string"},
                "password": {"type": "string"}
            }
        },
        "LoginRequest": {
            "type": "object",
            "required": ["email", "password"],
            "properties": {
                "email": {"type": "string"},
                "password": {"type": "string"}
            }
        },
        "CreateOrderRequest": {
            "type": "object",
            "required": ["product_id", "quantity"],
            "properties": {
                "product_id": {"type": "integer"},
                "quantity": {"type": "integer"},
                "group_id": {"type": "integer"}
            }
        },
        "APIProxyRequest": {
            "type": "object",
            "required": ["provider", "model", "messages"],
            "properties": {
                "provider": {"type": "string", "enum": ["openai", "anthropic", "google"]},
                "model": {"type": "string"},
                "messages": {"type": "array", "items": {"$ref": "#/definitions/ChatMessage"}},
                "temperature": {"type": "number"},
                "max_tokens": {"type": "integer"}
            }
        },
        "ChatMessage": {
            "type": "object",
            "properties": {
                "role": {"type": "string"},
                "content": {"type": "string"}
            }
        }
    },
    "securityDefinitions": {
        "Bearer": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header",
            "description": "Bearer token authentication"
        },
        "ApiKeyAuth": {
            "type": "apiKey",
            "name": "X-API-Key",
            "in": "header",
            "description": "API key authentication"
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

var SwaggerInfo = swaggerInfo{
	Version:     "1.0",
	Host:        "localhost:8080",
	BasePath:    "/api/v1",
	Schemes:     []string{"http", "https"},
	Title:       "Pintuotuo API",
	Description: "Pintuotuo - AI Token Trading Platform API Documentation",
}

type s struct{}

func (s *s) ReadDoc() string {
	return docTemplate
}

func init() {
	swag.Register("swagger", &s{})
}
