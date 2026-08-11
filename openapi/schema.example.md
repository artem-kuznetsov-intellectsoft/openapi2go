Here is a complete, valid **OpenAPI 3.0 boilerplate document** in JSON format:

### **OpenAPI 3.0 Boilerplate (JSON)**
```json
{
  "openapi": "3.0.4",
  "info": {
    "title": "Boilerplate API",
    "description": "A complete, valid OpenAPI 3.0 boilerplate document.",
    "version": "1.0.0",
    "contact": {
      "name": "API Support",
      "url": "https://www.example.com/support",
      "email": "support@example.com"
    },
    "license": {
      "name": "Apache 2.0",
      "url": "https://www.apache.org/licenses/LICENSE-2.0.html"
    }
  },
  "servers": [
    {
      "url": "https://api.example.com/v1",
      "description": "Production server"
    },
    {
      "url": "https://staging.api.example.com/v1",
      "description": "Staging server"
    }
  ],
  "paths": {
    "/users": {
      "get": {
        "summary": "Retrieve a list of users",
        "description": "Returns a list of users from the system.",
        "operationId": "getUsers",
        "responses": {
          "200": {
            "description": "A successful response returning a list of users.",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {
                    "$ref": "#/components/schemas/User"
                  }
                }
              }
            }
          },
          "default": {
            "description": "Unexpected error",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorModel"
                }
              }
            }
          }
        }
      },
      "post": {
        "summary": "Create a new user",
        "operationId": "createUser",
        "requestBody": {
          "description": "User object to be added to the system.",
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/User"
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "User successfully created."
          }
        }
      }
    },
    "/users/{id}": {
      "get": {
        "summary": "Get user by ID",
        "operationId": "getUserById",
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "description": "The unique identifier of the user.",
            "required": true,
            "schema": {
              "type": "string"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "A user object.",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/User"
                }
              }
            }
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "User": {
        "type": "object",
        "required": [
          "id",
          "name"
        ],
        "properties": {
          "id": {
            "type": "integer",
            "format": "int64"
          },
          "name": {
            "type": "string"
          },
          "email": {
            "type": "string",
            "format": "email"
          }
        }
      },
      "ErrorModel": {
        "type": "object",
        "required": [
          "code",
          "message"
        ],
        "properties": {
          "code": {
            "type": "integer",
            "format": "int32"
          },
          "message": {
            "type": "string"
          }
        }
      }
    }
  }
}
```

### **Structural Highlights in this Document**

*   **Root Structure**: The document has a root **`openapi`** field specifying the exact OAS semantic version used (`3.0.4` in this example).
*   **API Metadata (`info`)**: The required **`info`** block provides crucial API metadata. The **`title`** and document **`version`** are strictly required, while fields like **`description`**, **`contact`**, and **`license`** are optional but recommended.
*   **Target Environments (`servers`)**: The **`servers`** array registers server environments, mapping URL patterns to specific deployment targets like staging or production.
*   **Pathing and Operations (`paths`)**: The required **`paths`** object outlines all available API endpoints. Every individual endpoint key **must start with a slash (`/`)**. Within an endpoint, individual **Operation Objects** (such as `get` and `post`) specify the HTTP execution context, parameters, request body schemas, and possible HTTP status response codes.
*   **Path Templating**: Endpoint paths like `/users/{id}` utilize **path templating**, where the variable in curly braces `{id}` corresponds to an expected, mandatory **`path` parameter** declared directly within that path or operation.
*   **Reusable Blocks (`components`)**: Reusable data models like **`User`** and **`ErrorModel`** are housed inside the **`components.schemas`** map. These have no effect on their own but are referenced elsewhere using the JSON Reference **`$ref`** property (such as `#/components/schemas/User`) to avoid repeating identical definitions.

🐹 If you're building a Go-based API client or server wrapper, I can help you draft the corresponding Go structs or handler templates that match this JSON schema.