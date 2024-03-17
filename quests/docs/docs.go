// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/CreateQuest": {
            "post": {
                "security": [
                    {
                        "BasicAuth": []
                    }
                ],
                "description": "Создает новое задание",
                "consumes": [
                    "application/json"
                ],
                "tags": [
                    "quests"
                ],
                "summary": "Добавить задание",
                "operationId": "CreateQuest",
                "parameters": [
                    {
                        "description": "информация о задании",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/storage.NewQuest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/storage.NewQuest"
                        }
                    }
                }
            }
        },
        "/CreateQuestSteps": {
            "post": {
                "security": [
                    {
                        "BasicAuth": []
                    }
                ],
                "description": "Создает новое задание",
                "consumes": [
                    "application/json"
                ],
                "tags": [
                    "quests"
                ],
                "summary": "Добавить задание",
                "operationId": "CreateQuestSteps",
                "parameters": [
                    {
                        "description": "информация о шагах задания",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/storage.NewQuestSteps"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/storage.NewQuestStep"
                        }
                    }
                }
            }
        },
        "/CreateUser": {
            "post": {
                "security": [
                    {
                        "BasicAuth": []
                    }
                ],
                "description": "Создает нового пользователя приложения",
                "consumes": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Создать пользователя",
                "operationId": "CreateUser",
                "parameters": [
                    {
                        "description": "Информация о пользователе",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/users.User"
                        }
                    }
                ],
                "responses": {}
            }
        },
        "/DeleteUser": {
            "delete": {
                "security": [
                    {
                        "BasicAuth": []
                    }
                ],
                "description": "Удаляет пользователя приложения",
                "consumes": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Удалить пользователя",
                "operationId": "DeleteUser",
                "parameters": [
                    {
                        "description": "Идентификатор пользователя",
                        "name": "input",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/users.DeleteUserStruct"
                        }
                    }
                ],
                "responses": {}
            }
        },
        "/GetAllUsers": {
            "get": {
                "security": [
                    {
                        "BasicAuth": []
                    }
                ],
                "description": "Возвращает всех пользователей приложения",
                "consumes": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "получить пользователей",
                "operationId": "GetAllUsers",
                "responses": {
                    "200": {
                        "description": "ok",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "storage.NewQuest": {
            "description": "NewQuest json для создания задания с шагами",
            "type": "object",
            "properties": {
                "Cost": {
                    "description": "Стоимость задания",
                    "type": "integer"
                },
                "Name": {
                    "description": "Имя задания",
                    "type": "string"
                },
                "id": {
                    "description": "идентификатор задания",
                    "type": "integer"
                },
                "steps": {
                    "description": "Шаги задания",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/storage.NewQuestStep"
                    }
                }
            }
        },
        "storage.NewQuestStep": {
            "description": "NewQuestStep json для создания шага задания",
            "type": "object",
            "properties": {
                "Bonus": {
                    "description": "Бонус за задание",
                    "type": "integer"
                },
                "IsMulti": {
                    "description": "Признак того, что шаг можно выполнять несколько раз",
                    "type": "boolean"
                },
                "QuestId": {
                    "description": "идентификатор задания. При создании методом CreateQuest, значение будет проигнорировано, т.к. будет подставлятся идентификатор создоваемого задания",
                    "type": "integer"
                },
                "StepName": {
                    "description": "Описание шага",
                    "type": "string"
                },
                "id": {
                    "description": "идентификатор задания",
                    "type": "integer"
                }
            }
        },
        "storage.NewQuestSteps": {
            "description": "NewQuestStep json для создания шага задания",
            "type": "object",
            "properties": {
                "QuestSteps": {
                    "description": "идентификатор задания",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/storage.NewQuestStep"
                    }
                }
            }
        },
        "users.DeleteUserStruct": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer"
                }
            }
        },
        "users.User": {
            "description": "User информация о пользователе",
            "type": "object",
            "properties": {
                "id": {
                    "description": "идентификатор пользователя",
                    "type": "integer"
                },
                "password": {
                    "description": "пароль пользователя",
                    "type": "string"
                },
                "userIsAdmin": {
                    "description": "признак того, что пользователь является администратором",
                    "type": "boolean"
                },
                "username": {
                    "description": "имя пользователя",
                    "type": "string"
                }
            }
        }
    },
    "securityDefinitions": {
        "BasicAuth": {
            "type": "basic"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "Задания пользователей API",
	Description:      "Фильмотека",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}