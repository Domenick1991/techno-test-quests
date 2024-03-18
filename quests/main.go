package main

import (
	_ "github.com/lib/pq"
	"github.com/swaggo/http-swagger"
	"net/http"
	"techno-test_quests/quests/config"
	"techno-test_quests/quests/handlers/history"
	"techno-test_quests/quests/handlers/quest"

	_ "techno-test_quests/quests/docs"
	"techno-test_quests/quests/handlers/auth"
	users "techno-test_quests/quests/handlers/user"
	slogpretty "techno-test_quests/quests/lib"
	storage2 "techno-test_quests/quests/storage"
)

// @title Задания пользователей API
// @version 1.0
// @description Фильмотека
// @host localhost:8080
// @securitydefinitions.basic BasicAuth
// @in header
// @name Authorization
func main() {
	//загружаем конфиг
	cfg := config.MustLoad()

	//инициализируем логер
	logger := slogpretty.SetupLogger()
	logger.Info("Logger is start")

	db, err := storage2.New(cfg.DbStorage)
	if err != nil {
		logger.Error("Database service is not start: ", err.Error())
		return
	}

	//создаем все необходимые таблицы
	err = db.Init()
	if err != nil {
		logger.Error("Initialization database complete with error: ", err.Error())
	}
	logger.Info("Initialization database complete")

	//роут
	mux := http.NewServeMux()
	mux.HandleFunc("/", auth.NonPage)
	mux.HandleFunc("/swagger/*", httpSwagger.Handler(httpSwagger.URL("http://localhost:8080/swagger/doc.json")))
	mux.HandleFunc("/GetAllUsers", auth.AdminAuth(users.GetAllUsers(db), db))
	mux.HandleFunc("/CreateUser", auth.AdminAuth(users.CreateUser(db), db))
	mux.HandleFunc("/DeleteUser", auth.AdminAuth(users.DeleteUser(db), db))
	mux.HandleFunc("/CreateQuest", auth.AdminAuth(quest.CreateQuest(db, logger), db))
	mux.HandleFunc("/CreateQuestSteps", auth.AdminAuth(quest.CreateQuestSteps(db, logger), db))
	mux.HandleFunc("/CompleteSteps", auth.AdminAuth(history.CompleteSteps(db, logger), db))
	mux.HandleFunc("/UpdateQuestSteps", auth.AdminAuth(quest.UpdateQuestSteps(db, logger), db))
	mux.HandleFunc("/GetHistory", auth.AdminAuth(history.GetHistory(db, logger), db))
	mux.HandleFunc("/GetQuests", auth.AdminAuth(quest.GetQuests(db, logger), db))

	//запуск сервера
	server := &http.Server{
		Addr:         cfg.HttpServer.Address,
		Handler:      mux,
		ReadTimeout:  cfg.HttpServer.TimeoutRequest,
		WriteTimeout: cfg.HttpServer.IdleTimeout,
		IdleTimeout:  cfg.HttpServer.IdleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Error("Server does not started")
	}

}
