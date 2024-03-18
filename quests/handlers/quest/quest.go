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

// Quests model info
// @Description Quests json информация о заданиях и их шагов
type Quests struct {
	Id        string  `json:"Id" db:"id"`               //ИД задания
	QuestName string  `json:"QuestName" db:"questname"` //Имя выполненного задания пользователем
	Steps     []Steps `json:"Steps" db:"-`              //Шаги задания
}

func (quest *Quests) TableName() string {
	return "quests"
}

type Steps struct {
	StepName string `json:"StepName" db:"stepname"` //Имя шага
	Id       int    `json:"Id" db:"id"`             //ИД шага
	Bonus    int    `json:"Bonus" db:"bonus"`       //Бонус за выполнение шага
	IsMulti  bool   `json:"isMulti" db:"ismulti"`   //Признак того, что шаг можно выполнять повторно
}

// @Summary Обновить шаг к заданию
// @Tags quests
// @Description Создает новое задание
// @id GetQuests
// @Accept json
// @Procedure json
// @router /GetQuests [GET]
// @Success 200 {object} Quests
// @Security BasicAuth
func GetQuests(storage *storages.Storage, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		storages.RequestTolog(r, logger)
		if r.Method == http.MethodGet {

			var quests []Quests
			err := storage.DB.Select().From("quests").All(&quests)
			if err != nil {
				storages.HttpResponse(w, http.StatusInternalServerError, "Ошибка при получении данных о заданиях")
				return
			}
			for i, quest := range quests {
				var steps []Steps
				err = storage.DB.Select().From("queststeps").Where(dbx.HashExp{"questid": quest.Id}).All(&steps)
				if err != nil {
					storages.HttpResponse(w, http.StatusInternalServerError, "Ошибка при получении данных о заданиях")
					return
				}
				quests[i].Steps = steps
			}
			result, _ := json.MarshalIndent(quests, "", "\t")
			storages.HttpResponseObject(w, http.StatusInternalServerError, result)
		} else {
			storages.HttpResponse(w, http.StatusMethodNotAllowed, "Метод не поддерживается, используйте метод POST")
		}
	}
}

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
