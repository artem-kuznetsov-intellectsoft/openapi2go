An **OpenAPI 3.0 Document** is represented as a JSON (or YAML) object whose root structure (the **OpenAPI Object**) contains several required and optional fields. A generic JSON Schema that models this root structure—based on the requirements defined in the specification—is structured as follows:

### **Generic JSON Schema for an OpenAPI 3.0 Document**
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "OpenAPI Root Document Schema (Simplified)",
  "type": "object",
  "required": [
    "openapi",
    "info",
    "paths"
  ],
  "properties": {
    "openapi": {
      "type": "string",
      "description": "The version number of the OpenAPI Specification that the OpenAPI Document uses."
    },
    "info": {
      "type": "object",
      "description": "Provides metadata about the API.",
      "required": [
        "title",
        "version"
      ],
      "properties": {
        "title": {
          "type": "string",
          "description": "The title of the API."
        },
        "description": {
          "type": "string",
          "description": "A description of the API."
        },
        "termsOfService": {
          "type": "string",
          "format": "uri"
        },
        "contact": {
          "type": "object",
          "properties": {
            "name": { "type": "string" },
            "url": { "type": "string" },
            "email": { "type": "string" }
          }
        },
        "license": {
          "type": "object",
          "required": [
            "name"
          ],
          "properties": {
            "name": { "type": "string" },
            "url": { "type": "string" }
          }
        },
        "version": {
          "type": "string",
          "description": "The version of the OpenAPI Document itself."
        }
      }
    },
    "servers": {
      "type": "array",
      "description": "An array of Server Objects, which provide connectivity information to a target server.",
      "items": {
        "type": "object",
        "required": [
          "url"
        ],
        "properties": {
          "url": {
            "type": "string",
            "description": "A URL to the target host."
          },
          "description": {
            "type": "string"
          },
          "variables": {
            "type": "object"
          }
        }
      }
    },
    "paths": {
      "type": "object",
      "description": "The available paths and operations for the API.",
      "patternProperties": {
        "^/": {
          "type": "object",
          "description": "Describes the operations available on a single path."
        }
      },
      "additionalProperties": false
    },
    "components": {
      "type": "object",
      "description": "Holds a set of reusable objects for different aspects of the OAS.",
      "properties": {
        "schemas": { "type": "object" },
        "responses": { "type": "object" },
        "parameters": { "type": "object" },
        "examples": { "type": "object" },
        "requestBodies": { "type": "object" },
        "headers": { "type": "object" },
        "securitySchemes": { "type": "object" },
        "links": { "type": "object" },
        "callbacks": { "type": "object" }
      }
    },
    "security": {
      "type": "array",
      "description": "A declaration of which security mechanisms can be used across the API.",
      "items": {
        "type": "object",
        "additionalProperties": {
          "type": "array",
          "items": {
            "type": "string"
          }
        }
      }
    },
    "tags": {
      "type": "array",
      "items": {
        "type": "object",
        "required": [
          "name"
        ],
        "properties": {
          "name": {
            "type": "string"
          },
          "description": {
            "type": "string"
          },
          "externalDocs": {
            "type": "object"
          }
        }
      }
    },
    "externalDocs": {
      "type": "object",
      "required": [
        "url"
      ],
      "properties": {
        "description": { "type": "string" },
        "url": { "type": "string" }
      }
    }
  }
}
```

### **Key Structure Breakdown**

*   **`openapi`**: A **required string** indicating the semantic version of the specification used.
*   **`info`**: A **required Info Object** supplying metadata about the API. Within this object, **`title`** and **`version`** are strictly **required**.
*   **`paths`**: A **required Paths Object** mapped to individual operations. Every valid path key in this object **must begin with a forward slash (`/`)**.
*   **`servers`**: An optional **array of Server Objects**. Each server definition has a **required `url` field** (which supports server variables and relative paths).
*   **`components`**: An optional **Components Object** holding maps of reusable schemas, responses, parameters, and other definitions. These have no effect on the API unless explicitly referenced elsewhere in the description.
*   **`security`**: An optional **array of Security Requirement Objects** defining the authorization protocols.
*   **`tags`**: An optional **list of Tag Objects** used to group and categorize operations. Each tag **must have a unique `name` string**.
*   **`externalDocs`**: An optional **External Documentation Object** referencing external resources with a **required `url` field**.

💡 Would you like me to write a complete, valid OpenAPI 3.0 boilerplate document in JSON or YAML to see how these schema components are implemented in practice?