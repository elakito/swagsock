// Code generated by go-swagger; DO NOT EDIT.

package restapi

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"
)

var (
	// SwaggerJSON embedded version of the swagger document used at generation time
	SwaggerJSON json.RawMessage
	// FlatSwaggerJSON embedded flattened version of the swagger document used at generation time
	FlatSwaggerJSON json.RawMessage
)

func init() {
	SwaggerJSON = json.RawMessage([]byte(`{
  "schemes": [
    "http",
    "https"
  ],
  "swagger": "2.0",
  "info": {
    "description": "This is a greeter demo.",
    "title": "greeter demo API",
    "termsOfService": "http://www.github.com/elakito/swaggersocket/",
    "license": {
      "name": "Apache 2.0",
      "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
    },
    "version": "1.0.1"
  },
  "basePath": "/samples/greeter",
  "paths": {
    "/v1/echo": {
      "post": {
        "description": "Echo back the message",
        "consumes": [
          "text/plain"
        ],
        "produces": [
          "application/json"
        ],
        "summary": "Echo back the message",
        "operationId": "echo",
        "parameters": [
          {
            "description": "A message",
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "type": "string"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Returns an echo message",
            "schema": {
              "$ref": "#/definitions/echoReply"
            }
          }
        }
      }
    },
    "/v1/greet": {
      "get": {
        "description": "Show who has been greeted",
        "produces": [
          "application/json"
        ],
        "summary": "Show who has been greeted",
        "operationId": "getGreetSummary",
        "responses": {
          "200": {
            "description": "Greeting summary",
            "schema": {
              "$ref": "#/definitions/greetingSummary"
            }
          }
        }
      }
    },
    "/v1/greet/{name}": {
      "get": {
        "description": "Show when the person was last greeted",
        "produces": [
          "application/json"
        ],
        "summary": "Show when the person was greeted last",
        "operationId": "getGreetStatus",
        "parameters": [
          {
            "type": "string",
            "description": "greeter's name",
            "name": "name",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "Greeting status",
            "schema": {
              "$ref": "#/definitions/greetingStatus"
            }
          }
        }
      },
      "post": {
        "description": "Greet someone",
        "consumes": [
          "application/json"
        ],
        "produces": [
          "application/json"
        ],
        "summary": "Greet someone",
        "operationId": "greet",
        "parameters": [
          {
            "type": "string",
            "description": "greeter's name",
            "name": "name",
            "in": "path",
            "required": true
          },
          {
            "description": "A greeting message",
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/greeting"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "A greeting response",
            "schema": {
              "$ref": "#/definitions/greetingReply"
            }
          }
        }
      }
    },
    "/v1/greet/{name}/upload": {
      "post": {
        "description": "Upload a greeting card",
        "consumes": [
          "multipart/form-data"
        ],
        "produces": [
          "application/json"
        ],
        "summary": "Upload a greeting card",
        "operationId": "upload",
        "parameters": [
          {
            "type": "string",
            "description": "greeter's name",
            "name": "name",
            "in": "path",
            "required": true
          },
          {
            "type": "file",
            "description": "greeting card file to upload",
            "name": "file",
            "in": "formData"
          }
        ],
        "responses": {
          "200": {
            "description": "Upload status",
            "schema": {
              "$ref": "#/definitions/uploadStatus"
            }
          },
          "500": {
            "description": "Storage Full",
            "schema": {
              "type": "string"
            }
          }
        }
      }
    },
    "/v1/ping": {
      "get": {
        "description": "Ping the greeter service",
        "produces": [
          "application/json"
        ],
        "summary": "Ping the greeter service",
        "operationId": "ping",
        "responses": {
          "200": {
            "description": "Returns a pong",
            "schema": {
              "$ref": "#/definitions/pong"
            }
          }
        }
      }
    },
    "/v1/subscribe/{name}": {
      "get": {
        "description": "Subscribe to the greeteing events",
        "produces": [
          "application/json"
        ],
        "summary": "Subscribe to the greeting events",
        "operationId": "subscribe",
        "parameters": [
          {
            "type": "string",
            "description": "greeted name",
            "name": "name",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "Returns a subscription information",
            "schema": {
              "type": "array",
              "items": {
                "$ref": "#/definitions/greetingReply"
              }
            }
          }
        }
      }
    },
    "/v1/unsubscribe/{sid}": {
      "delete": {
        "description": "Unsubscribe from the greeting events",
        "produces": [
          "application/json"
        ],
        "summary": "Unsubscribe from the greeting events",
        "operationId": "unsubscribe",
        "parameters": [
          {
            "type": "string",
            "description": "subscription id",
            "name": "sid",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "Returns the subscription summary",
            "schema": {
              "$ref": "#/definitions/greetingSummary"
            }
          }
        }
      }
    }
  },
  "definitions": {
    "echoReply": {
      "type": "object",
      "properties": {
        "echo": {
          "description": "echo",
          "type": "string"
        }
      }
    },
    "greeting": {
      "type": "object",
      "properties": {
        "name": {
          "description": "greeting person",
          "type": "string"
        },
        "text": {
          "description": "greeting text",
          "type": "string"
        }
      }
    },
    "greetingReply": {
      "type": "object",
      "properties": {
        "from": {
          "description": "greeting person",
          "type": "string"
        },
        "name": {
          "description": "greeted person",
          "type": "string"
        },
        "text": {
          "description": "greeting reply text",
          "type": "string"
        }
      }
    },
    "greetingStatus": {
      "type": "object",
      "properties": {
        "count": {
          "description": "greeting count",
          "type": "integer",
          "format": "int32"
        },
        "name": {
          "description": "greeted person",
          "type": "string"
        }
      }
    },
    "greetingSummary": {
      "type": "object",
      "properties": {
        "greeted": {
          "description": "greeted persons",
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "total": {
          "description": "total greeting count",
          "type": "integer",
          "format": "int32"
        }
      }
    },
    "pong": {
      "type": "object",
      "properties": {
        "pong": {
          "description": "pong counter",
          "type": "integer",
          "format": "int32"
        }
      }
    },
    "uploadStatus": {
      "type": "object",
      "properties": {
        "name": {
          "description": "greeting name",
          "type": "string"
        },
        "size": {
          "description": "greeting card size",
          "type": "integer",
          "format": "int32"
        }
      }
    }
  },
  "securityDefinitions": {
    "UserSecurity": {
      "type": "basic"
    }
  }
}`))
	FlatSwaggerJSON = json.RawMessage([]byte(`{
  "schemes": [
    "http",
    "https"
  ],
  "swagger": "2.0",
  "info": {
    "description": "This is a greeter demo.",
    "title": "greeter demo API",
    "termsOfService": "http://www.github.com/elakito/swaggersocket/",
    "license": {
      "name": "Apache 2.0",
      "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
    },
    "version": "1.0.1"
  },
  "basePath": "/samples/greeter",
  "paths": {
    "/v1/echo": {
      "post": {
        "description": "Echo back the message",
        "consumes": [
          "text/plain"
        ],
        "produces": [
          "application/json"
        ],
        "summary": "Echo back the message",
        "operationId": "echo",
        "parameters": [
          {
            "description": "A message",
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "type": "string"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Returns an echo message",
            "schema": {
              "$ref": "#/definitions/echoReply"
            }
          }
        }
      }
    },
    "/v1/greet": {
      "get": {
        "description": "Show who has been greeted",
        "produces": [
          "application/json"
        ],
        "summary": "Show who has been greeted",
        "operationId": "getGreetSummary",
        "responses": {
          "200": {
            "description": "Greeting summary",
            "schema": {
              "$ref": "#/definitions/greetingSummary"
            }
          }
        }
      }
    },
    "/v1/greet/{name}": {
      "get": {
        "description": "Show when the person was last greeted",
        "produces": [
          "application/json"
        ],
        "summary": "Show when the person was greeted last",
        "operationId": "getGreetStatus",
        "parameters": [
          {
            "type": "string",
            "description": "greeter's name",
            "name": "name",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "Greeting status",
            "schema": {
              "$ref": "#/definitions/greetingStatus"
            }
          }
        }
      },
      "post": {
        "description": "Greet someone",
        "consumes": [
          "application/json"
        ],
        "produces": [
          "application/json"
        ],
        "summary": "Greet someone",
        "operationId": "greet",
        "parameters": [
          {
            "type": "string",
            "description": "greeter's name",
            "name": "name",
            "in": "path",
            "required": true
          },
          {
            "description": "A greeting message",
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/greeting"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "A greeting response",
            "schema": {
              "$ref": "#/definitions/greetingReply"
            }
          }
        }
      }
    },
    "/v1/greet/{name}/upload": {
      "post": {
        "description": "Upload a greeting card",
        "consumes": [
          "multipart/form-data"
        ],
        "produces": [
          "application/json"
        ],
        "summary": "Upload a greeting card",
        "operationId": "upload",
        "parameters": [
          {
            "type": "string",
            "description": "greeter's name",
            "name": "name",
            "in": "path",
            "required": true
          },
          {
            "type": "file",
            "description": "greeting card file to upload",
            "name": "file",
            "in": "formData"
          }
        ],
        "responses": {
          "200": {
            "description": "Upload status",
            "schema": {
              "$ref": "#/definitions/uploadStatus"
            }
          },
          "500": {
            "description": "Storage Full",
            "schema": {
              "type": "string"
            }
          }
        }
      }
    },
    "/v1/ping": {
      "get": {
        "description": "Ping the greeter service",
        "produces": [
          "application/json"
        ],
        "summary": "Ping the greeter service",
        "operationId": "ping",
        "responses": {
          "200": {
            "description": "Returns a pong",
            "schema": {
              "$ref": "#/definitions/pong"
            }
          }
        }
      }
    },
    "/v1/subscribe/{name}": {
      "get": {
        "description": "Subscribe to the greeteing events",
        "produces": [
          "application/json"
        ],
        "summary": "Subscribe to the greeting events",
        "operationId": "subscribe",
        "parameters": [
          {
            "type": "string",
            "description": "greeted name",
            "name": "name",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "Returns a subscription information",
            "schema": {
              "type": "array",
              "items": {
                "$ref": "#/definitions/greetingReply"
              }
            }
          }
        }
      }
    },
    "/v1/unsubscribe/{sid}": {
      "delete": {
        "description": "Unsubscribe from the greeting events",
        "produces": [
          "application/json"
        ],
        "summary": "Unsubscribe from the greeting events",
        "operationId": "unsubscribe",
        "parameters": [
          {
            "type": "string",
            "description": "subscription id",
            "name": "sid",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "Returns the subscription summary",
            "schema": {
              "$ref": "#/definitions/greetingSummary"
            }
          }
        }
      }
    }
  },
  "definitions": {
    "echoReply": {
      "type": "object",
      "properties": {
        "echo": {
          "description": "echo",
          "type": "string"
        }
      }
    },
    "greeting": {
      "type": "object",
      "properties": {
        "name": {
          "description": "greeting person",
          "type": "string"
        },
        "text": {
          "description": "greeting text",
          "type": "string"
        }
      }
    },
    "greetingReply": {
      "type": "object",
      "properties": {
        "from": {
          "description": "greeting person",
          "type": "string"
        },
        "name": {
          "description": "greeted person",
          "type": "string"
        },
        "text": {
          "description": "greeting reply text",
          "type": "string"
        }
      }
    },
    "greetingStatus": {
      "type": "object",
      "properties": {
        "count": {
          "description": "greeting count",
          "type": "integer",
          "format": "int32"
        },
        "name": {
          "description": "greeted person",
          "type": "string"
        }
      }
    },
    "greetingSummary": {
      "type": "object",
      "properties": {
        "greeted": {
          "description": "greeted persons",
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "total": {
          "description": "total greeting count",
          "type": "integer",
          "format": "int32"
        }
      }
    },
    "pong": {
      "type": "object",
      "properties": {
        "pong": {
          "description": "pong counter",
          "type": "integer",
          "format": "int32"
        }
      }
    },
    "uploadStatus": {
      "type": "object",
      "properties": {
        "name": {
          "description": "greeting name",
          "type": "string"
        },
        "size": {
          "description": "greeting card size",
          "type": "integer",
          "format": "int32"
        }
      }
    }
  },
  "securityDefinitions": {
    "UserSecurity": {
      "type": "basic"
    }
  }
}`))
}
