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
    "description": "This is a swaggersocket chat demo.",
    "title": "swaggersocket chat demo API",
    "termsOfService": "http://www.sap.com/vora/terms/",
    "license": {
      "name": "Apache 2.0",
      "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
    },
    "version": "1.0.0"
  },
  "basePath": "/samples/chat",
  "paths": {
    "/v1/chat/{name}": {
      "post": {
        "description": "Chat with someone",
        "consumes": [
          "application/json"
        ],
        "produces": [
          "application/json"
        ],
        "summary": "chat with someone",
        "operationId": "chat",
        "parameters": [
          {
            "type": "string",
            "description": "author's name",
            "name": "name",
            "in": "path",
            "required": true
          },
          {
            "description": "A chat message",
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "type": "object",
              "properties": {
                "text": {
                  "description": "chat message",
                  "type": "string"
                }
              }
            }
          }
        ],
        "responses": {
          "200": {
            "description": "A chat acknowledgement",
            "schema": {
              "$ref": "#/definitions/message"
            }
          }
        }
      }
    },
    "/v1/members": {
      "get": {
        "description": "Get the list of chat memebers",
        "produces": [
          "application/json"
        ],
        "summary": "Get the list of chat members",
        "operationId": "members",
        "responses": {
          "200": {
            "description": "Memebers list",
            "schema": {
              "type": "array",
              "items": {
                "type": "string"
              }
            }
          }
        }
      }
    },
    "/v1/subscribe/{name}": {
      "get": {
        "description": "Subscribe to the chat events",
        "produces": [
          "application/json"
        ],
        "summary": "Subscribe to the chat events",
        "operationId": "subscribe",
        "parameters": [
          {
            "type": "string",
            "description": "author's name",
            "name": "name",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "Subscription information",
            "schema": {
              "type": "array",
              "items": {
                "$ref": "#/definitions/message"
              }
            }
          }
        }
      }
    },
    "/v1/unsubscribe/{sid}": {
      "delete": {
        "description": "Unsubscribe from the chat events",
        "produces": [
          "application/json"
        ],
        "summary": "Unsubscribe from the chat events",
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
            "description": "Chat summary",
            "schema": {
              "$ref": "#/definitions/summary"
            }
          }
        }
      }
    }
  },
  "definitions": {
    "message": {
      "type": "object",
      "properties": {
        "name": {
          "description": "chat author",
          "type": "string"
        },
        "text": {
          "description": "message text",
          "type": "string"
        },
        "type": {
          "description": "message type",
          "type": "string",
          "enum": [
            "message",
            "joined",
            "left"
          ]
        }
      }
    },
    "summary": {
      "type": "object",
      "properties": {
        "authors": {
          "description": "chatted persons",
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "total": {
          "description": "total chat message count",
          "type": "integer",
          "format": "int32"
        }
      }
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
    "description": "This is a swaggersocket chat demo.",
    "title": "swaggersocket chat demo API",
    "termsOfService": "http://www.sap.com/vora/terms/",
    "license": {
      "name": "Apache 2.0",
      "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
    },
    "version": "1.0.0"
  },
  "basePath": "/samples/chat",
  "paths": {
    "/v1/chat/{name}": {
      "post": {
        "description": "Chat with someone",
        "consumes": [
          "application/json"
        ],
        "produces": [
          "application/json"
        ],
        "summary": "chat with someone",
        "operationId": "chat",
        "parameters": [
          {
            "type": "string",
            "description": "author's name",
            "name": "name",
            "in": "path",
            "required": true
          },
          {
            "description": "A chat message",
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "type": "object",
              "properties": {
                "text": {
                  "description": "chat message",
                  "type": "string"
                }
              }
            }
          }
        ],
        "responses": {
          "200": {
            "description": "A chat acknowledgement",
            "schema": {
              "$ref": "#/definitions/message"
            }
          }
        }
      }
    },
    "/v1/members": {
      "get": {
        "description": "Get the list of chat memebers",
        "produces": [
          "application/json"
        ],
        "summary": "Get the list of chat members",
        "operationId": "members",
        "responses": {
          "200": {
            "description": "Memebers list",
            "schema": {
              "type": "array",
              "items": {
                "type": "string"
              }
            }
          }
        }
      }
    },
    "/v1/subscribe/{name}": {
      "get": {
        "description": "Subscribe to the chat events",
        "produces": [
          "application/json"
        ],
        "summary": "Subscribe to the chat events",
        "operationId": "subscribe",
        "parameters": [
          {
            "type": "string",
            "description": "author's name",
            "name": "name",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "Subscription information",
            "schema": {
              "type": "array",
              "items": {
                "$ref": "#/definitions/message"
              }
            }
          }
        }
      }
    },
    "/v1/unsubscribe/{sid}": {
      "delete": {
        "description": "Unsubscribe from the chat events",
        "produces": [
          "application/json"
        ],
        "summary": "Unsubscribe from the chat events",
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
            "description": "Chat summary",
            "schema": {
              "$ref": "#/definitions/summary"
            }
          }
        }
      }
    }
  },
  "definitions": {
    "message": {
      "type": "object",
      "properties": {
        "name": {
          "description": "chat author",
          "type": "string"
        },
        "text": {
          "description": "message text",
          "type": "string"
        },
        "type": {
          "description": "message type",
          "type": "string",
          "enum": [
            "message",
            "joined",
            "left"
          ]
        }
      }
    },
    "summary": {
      "type": "object",
      "properties": {
        "authors": {
          "description": "chatted persons",
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "total": {
          "description": "total chat message count",
          "type": "integer",
          "format": "int32"
        }
      }
    }
  }
}`))
}