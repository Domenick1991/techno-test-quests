package quest

import (
	"encoding/json"
	"errors"
	"fmt"
	dbx "github.com/go-ozzo/ozzo-dbx"
	"log/slog"
	"net/http"
	"strconv"
	storages "techno-test_quests/quests/storage"
)

// @Summary Добавить задание
// @Tags quests
// @Description Создает новое задание
// @id CreateQuest
// @Accept json
// @Procedure json
// @router /CreateQuest [POST]
// @param input body storage.NewQuest true "информация о задании"
// @Success 200 {object} storage.NewQuest
// @Security BasicAuth
func CreateQuest(storage *storages.Storage, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		storages.RequestTolog(r, logger)
		if r.Method == http.MethodPost {
			var quest storages.NewQuest
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&quest)

			if err != nil {
				storages.HttpResponse(w, http.StatusBadRequest, "Неверный формат запроса "+err.Error())
				return
			}

			questDB, errlist := quest.ConvertToDB()
			if len(errlist) > 0 {
				result, _ := json.MarshalIndent(errlist, "", "\t")
				storages.HttpResponseObject(w, http.StatusInternalServerError, result)
				return
			}

			var oldquestDB = storages.NewQuestDB{}
			err = storage.DB.Select().From(questDB.TableName()).Where(dbx.HashExp{"questname": questDB.Name}).One(&oldquestDB)
			if err != nil {
				err = storage.DB.Model(&questDB).Insert("Name", "Cost")
				if err != nil {
					storages.HttpResponse(w, http.StatusInternalServerError, "Не удалось добавить задание")
				} else {

					//Если передавалась информация о шагах - добавляем и шаги
					if quest.QuestSteps != nil {
						for _, questStep := range quest.QuestSteps {
							questStep.QuestId = questDB.Id
							questStepDB, errlist := questStep.ConvertToDB()
							if len(errlist) > 0 {
								result, _ := json.MarshalIndent(errlist, "", "\t")
								storages.HttpResponseObject(w, http.StatusInternalServerError, result)
								return
							}
							err = createQuestStep(storage, questStepDB)
							if err != nil {
								storages.HttpResponse(w, http.StatusBadRequest, "Ошибка при добавлении шага: "+err.Error())
								return
							}
						}
					}
					storages.HttpResponse(w, http.StatusOK, "Успешно")
				}
			} else {
				storages.HttpResponse(w, http.StatusInternalServerError, "Задание с таким именем существует, id :"+strconv.Itoa(oldquestDB.Id))
			}

		} else {
			storages.HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод POST")
		}
	}
}

// createQuestStep добавляет шаг к заданию.
func createQuestStep(storage *storages.Storage, questStepDB storages.NewQuestStepDB) error {
	//Проверяем, что существует questID.

	var quest storages.NewQuestDB
	err := storage.DB.Select().From(quest.TableName()).Where(dbx.HashExp{"id": questStepDB.QuestId}).One(&quest)
	if err == nil {
		//Проверяем, что существует шаг.
		var oldquestStepDB = storages.NewQuestStepDB{}
		err = storage.DB.Select().From(questStepDB.TableName()).Where(dbx.HashExp{"stepname": questStepDB.StepName, "questid": questStepDB.QuestId}).One(&oldquestStepDB)
		if err != nil {
			err = storage.DB.Model(&questStepDB).Insert("QuestId", "StepName", "Bonus", "IsMulti")
			if err != nil {
				return errors.New("Не удалось добавить задание")
			}
		} else {
			return errors.New(fmt.Sprint("Не удалось добавить шаг '", questStepDB.StepName, "' т.к. шаг с именем '", questStepDB.StepName, "' уже сщуествует"))
		}
	} else {
		//TODO по хорошему, нужно делать уведомление о том, что шаг не добавлен, но, как мне кажется, это должно обрабатываться на стороне фронтенда, поэтому тут просто будем игнорировать
		return errors.New(fmt.Sprint("Не удалось добавить шаг '", questStepDB.StepName, "' т.к. задание с id '", questStepDB.QuestId, "' не сщуествует"))
	}
	return nil
}

// @Summary Добавить шаг к заданию
// @Tags quests
// @Description Добавляет новые шаги к заданию
// @id CreateQuestSteps
// @Accept json
// @Procedure json
// @router /CreateQuestSteps [POST]
// @param input body storage.NewQuestSteps true "информация о шагах задания"
// @Success 200 {object} storage.NewQuestStep
// @Security BasicAuth
func CreateQuestSteps(storage *storages.Storage, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		storages.RequestTolog(r, logger)
		if r.Method == http.MethodPost {
			var questSteps storages.NewQuestSteps
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&questSteps)

			if err != nil {
				storages.HttpResponse(w, http.StatusBadRequest, "Неверный формат запроса "+err.Error())
				return
			}

			for _, questStep := range questSteps.QuestSteps {
				questStepDB, errlist := questStep.ConvertToDB()
				if len(errlist) > 0 {
					result, _ := json.MarshalIndent(errlist, "", "\t")
					storages.HttpResponseObject(w, http.StatusInternalServerError, result)
					return
				}
				err = createQuestStep(storage, questStepDB)
				if err != nil {
					storages.HttpResponse(w, http.StatusBadRequest, "Ошибка при добавлении шага: "+err.Error())
					return
				}
			}
			storages.HttpResponse(w, http.StatusOK, "Успешно")
		} else {
			storages.HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод POST")
		}
	}
}

// @Summary Выполнить шаг
// @Tags quests
// @Description Устанавливает признак выполнения шага у пользователя
// @id CompleteSteps
// @Accept json
// @Procedure json
// @router /CompleteSteps [POST]
// @param input body storage.NewQuestSteps true "обновленная информация о шагах задания"
// @Success 200 {object} storage.NewQuestStep
// @Security BasicAuth
func CompleteSteps(storage *storages.Storage, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		storages.RequestTolog(r, logger)
		if r.Method == http.MethodPost {
			var сompleteSteps storages.NewCompleteSteps
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&сompleteSteps)

			if err != nil {
				storages.HttpResponse(w, http.StatusBadRequest, "Неверный формат запроса "+err.Error())
				return
			}

			for _, сompleteStep := range сompleteSteps.CompleteSteps {
				сompleteStepDB, errlist := сompleteStep.ConvertToDB()
				if len(errlist) > 0 {
					result, _ := json.MarshalIndent(errlist, "", "\t")
					storages.HttpResponseObject(w, http.StatusInternalServerError, result)
					return
				}

				//Если задание можно выполнить - выполняем, если нет, то просто игнорируем
				if checkCompliteStep(storage, сompleteStep) {
					err = storage.DB.Model(&сompleteStepDB).Insert()
					if err != nil {
						storages.HttpResponse(w, http.StatusBadRequest, "Не удалось выполнить задание"+err.Error())
						return
					}
				}
			}
			storages.HttpResponse(w, http.StatusOK, "Успешно")
		} else {
			storages.HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод POST")
		}
	}
}

// checkCompliteStep возвращает true, если шаг доступен пользователю для выполнения
func checkCompliteStep(storage *storages.Storage, сompleteStep storages.CompleteStep) bool {
	//Получаем id всех выполненных заданий

	queryText := "SELECT s.id FROM public.queststeps as s where (s.ismulti = false and s.id not in (select h.stepid from history h where h.userid = " + strconv.Itoa(сompleteStep.Userid) + " )) or s.ismulti = true"
	query := storage.DB.NewQuery(queryText)
	rows, err := query.Rows()
	if err != nil {

	}
	var stepIds map[int]bool = make(map[int]bool)
	for rows.Next() {
		var id int
		rows.Scan(&id)
		stepIds[id] = true
	}
	_, inMap := stepIds[сompleteStep.Stepid]
	return inMap
}

// @Summary Обновить шаг к заданию
// @Tags quests
// @Description Создает новое задание
// @id UpdateQuestSteps
// @Accept json
// @Procedure json
// @router /UpdateQuestSteps [POST]
// @param input body storage.NewQuestSteps true "обновленная информация о шагах задания"
// @Success 200 {object} storage.NewQuestStep
// @Security BasicAuth
func UpdateQuestSteps(storage *storages.Storage, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		storages.RequestTolog(r, logger)
		if r.Method == http.MethodPost {
			var updateQuestSteps storages.UpdateQuestSteps
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&updateQuestSteps)

			if err != nil {
				storages.HttpResponse(w, http.StatusBadRequest, "Неверный формат запроса "+err.Error())
				return
			}

			for _, questStep := range updateQuestSteps.QuestSteps {
				questStepDB, errlist := questStep.ConvertToDB()
				if len(errlist) > 0 {
					result, _ := json.MarshalIndent(errlist, "", "\t")
					storages.HttpResponseObject(w, http.StatusInternalServerError, result)
					return
				}
				params := questStepDB.GetUpdatesData()
				if len(params) > 0 {
					_, err = storage.DB.Update(questStepDB.TableName(), params, dbx.HashExp{"id": questStepDB.Id}).Execute()
					if err != nil {
						storages.HttpResponse(w, http.StatusBadRequest, "не удалось обновить задание"+err.Error())
						return
					}
				}
			}
			storages.HttpResponse(w, http.StatusOK, "Успешно")
		} else {
			storages.HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод POST")
		}
	}
}

// @Summary Обновить шаг к заданию
// @Tags quests
// @Description Создает новое задание
// @id GetHistory
// @Accept json
// @Procedure json
// @router /GetHistory [POST]
// @param input body storage.NewQuestSteps true "обновленная информация о шагах задания"
// @Success 200 {object} storage.UserBonus
// @Security BasicAuth
func GetHistory(storage *storages.Storage, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		storages.RequestTolog(r, logger)
		if r.Method == http.MethodGet {
			userId := r.Header.Get("userid")
			if userId != "" {
				questIds := getCompletedQuestId(storage, userId)
				if len(questIds) > 0 {
					userBonus := storages.UserBonus{}

					for _, questId := range questIds {
						CompletedQuest := GetCompletedQuestForUser(storage, userId, questId)
						userBonus.CompletedQuests = append(userBonus.CompletedQuests, CompletedQuest)
						userBonus.TotalBonus += CompletedQuest.Bonus
					}
					result, _ := json.MarshalIndent(userBonus, "", "\t")
					storages.HttpResponseObject(w, http.StatusInternalServerError, result)
				} else {
					storages.HttpResponse(w, http.StatusOK, "Пользователь еще не выполнял задания")
				}
			} else {
				storages.HttpResponse(w, http.StatusBadRequest, "Неверный формат запроса, укажите 'userid")
				return
			}

		} else {
			storages.HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод POST")
		}
	}
}

// getCompletedQuestId возвращает ИД заданий в которых участвовал пользователь
func getCompletedQuestId(storage *storages.Storage, userId string) []string {
	var questIds []string

	queryText := `SELECT distinct q.id
						FROM public.queststeps as s
						left join history as h on s.id = h.stepid
						left join quests as q on s.questid = q.id
						where h.userid = ` + userId
	query := storage.DB.NewQuery(queryText)
	rows, err := query.Rows()
	if err != nil {
		return questIds
	}

	for rows.Next() {
		var id string
		rows.Scan(&id)
		questIds = append(questIds, id)
	}
	return questIds
}

// GetCompletedQuestForUser Возвращает информацию по заданию для пользователя
func GetCompletedQuestForUser(storage *storages.Storage, userId, questId string) storages.UserCompletedQuest {
	UserCompletedQuest := storages.UserCompletedQuest{}

	UserCompletedQuest.QuestId = questId
	//TODO UserCompletedQuest.QuestName

	//Всего заданий
	queryText := `	SELECT count(*)
					FROM public.queststeps
					where questid = ` + questId
	UserCompletedQuest.AllStepsCount = storages.ExicuteCountSumQuery(storage, queryText)

	//Всего заданий выполненных пользователей
	queryText = `SELECT count(*)
					FROM public.queststeps as s
					left join history as h on s.id = h.stepid
					left join quests as q on s.questid = q.id
					where h.userid = ` + userId + ` and q.id = ` + questId
	UserCompletedQuest.CompletedStepsCount = storages.ExicuteCountSumQuery(storage, queryText)

	//Сумма бонуса за выполненные шаги задания
	queryText = `SELECT Sum(s.bonus)
					FROM public.queststeps as s
					left join history as h on s.id = h.stepid
					left join quests as q on s.questid = q.id
					where h.userid = ` + userId + ` and q.id = ` + questId
	UserCompletedQuest.Bonus = storages.ExicuteCountSumQuery(storage, queryText)

	//TODO UserCompletedQuest.CompletedSteps берем как сумму из CompletedStepsCount ?

	return UserCompletedQuest
}
