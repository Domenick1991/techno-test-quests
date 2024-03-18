package history

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	storages "techno-test_quests/quests/storage"
)

// UserBonus model info
// @Description UserBonus json для получения история выполнения заданий и их шагов
type UserBonus struct {
	TotalBonus      int                  `json:"TotalBonus"`      //Общий бонусный счет пользователя
	CompletedQuests []UserCompletedQuest `json:"ComplitedQuests"` //Список заданий в которых участвовал пользователь
}

type UserCompletedQuest struct {
	QuestId             string               `json:"QuestId"`             //ИД задания
	QuestName           string               `json:"QuestName"`           //Имя выполненного задания пользователем
	Bonus               int                  `json:"Bonus"`               //Сумма Бонусов за выполненные задания
	CompletedStepsCount int                  `json:"CompletedStepsCount"` //Кол-во выполненных шагов заданий пользователем
	AllStepsCount       int                  `json:"AllStepsCount"`       //Кол-во шагов, доступное в задании
	CompletedSteps      []UserCompletedSteps `json:"CompletedSteps"`      //Выполненные шаги пользователем
}
type UserCompletedSteps struct {
	StepName      string `json:"StepName"`      //Имя выполненного шага
	Count         int    `json:"Count"`         //Кол-во выполнений шага
	UserBonusStep int    `json:"UserBonusStep"` //Бонус пользователя за выполнение шага
}

// ExicuteCountSumQuery Обёртка предназначена для запросов, которые возвращают одно целое значение
func ExicuteCountSumQuery(storage *storages.Storage, queryText string) int {
	result := 0
	query := storage.DB.NewQuery(queryText)
	rows, err := query.Rows()
	if err != nil {
		return result
	}
	for rows.Next() {
		rows.Scan(&result)
	}
	return result
}

// @Summary Выполнить шаг
// @Tags history
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
// @Tags history
// @Description Создает новое задание
// @id GetHistory
// @Accept json
// @Procedure json
// @router /GetHistory [GET]
// @Success 200 {object} UserBonus
// @Security BasicAuth
func GetHistory(storage *storages.Storage, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		storages.RequestTolog(r, logger)
		if r.Method == http.MethodGet {
			userId := r.Header.Get("userid")
			if userId != "" {
				questIds := getCompletedQuestId(storage, userId)
				if len(questIds) > 0 {
					userBonus := UserBonus{}

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
func GetCompletedQuestForUser(storage *storages.Storage, userId, questId string) UserCompletedQuest {
	UserCompletedQuest := UserCompletedQuest{}

	UserCompletedQuest.QuestId = questId
	//TODO UserCompletedQuest.QuestName

	//Всего заданий
	queryText := `	SELECT count(*)
					FROM public.queststeps
					where questid = ` + questId
	UserCompletedQuest.AllStepsCount = ExicuteCountSumQuery(storage, queryText)

	//Всего заданий выполненных пользователей
	queryText = `SELECT count(*)
					FROM public.queststeps as s
					left join history as h on s.id = h.stepid
					left join quests as q on s.questid = q.id
					where h.userid = ` + userId + ` and q.id = ` + questId
	UserCompletedQuest.CompletedStepsCount = ExicuteCountSumQuery(storage, queryText)

	//Сумма бонуса за выполненные шаги задания
	queryText = `SELECT Sum(s.bonus)
					FROM public.queststeps as s
					left join history as h on s.id = h.stepid
					left join quests as q on s.questid = q.id
					where h.userid = ` + userId + ` and q.id = ` + questId
	UserCompletedQuest.Bonus = ExicuteCountSumQuery(storage, queryText)

	//region пройдемся по каждому выполненному шагу пользователя и посчитаем сколько раз был выполнен каждый шаг и сумму бонусов за это
	queryText = `SELECT distinct (s.id)
					FROM public.queststeps as s
					left join history as h on s.id = h.stepid
					left join quests as q on s.questid = q.id
					where h.userid = ` + userId + ` and q.id = ` + questId
	query := storage.DB.NewQuery(queryText)
	rowsSteps, _ := query.Rows()
	CompletedStepsCount := 0
	type stepInfo struct{ count, bonus int }
	for rowsSteps.Next() {
		var stepid string
		rowsSteps.Scan(&stepid)

		queryText = `SELECT s.stepname, s.bonus
					FROM public.queststeps as s
					left join history as h on s.id = h.stepid
					left join quests as q on s.questid = q.id
					where h.userid = ` + userId + ` and q.id = ` + questId + ` and s.id =` + stepid
		query = storage.DB.NewQuery(queryText)
		rows, _ := query.Rows()

		var stepsInfo map[string]stepInfo = make(map[string]stepInfo)
		for rows.Next() {
			var stepname string
			var bonus int
			rows.Scan(&stepname, &bonus)
			if value, inMap := stepsInfo[stepname]; inMap {
				value.count++
				value.bonus += bonus
				stepsInfo[stepname] = value
			} else {
				stepsInfo[stepname] = stepInfo{1, bonus}
			}

		}
		for key, value := range stepsInfo {
			UserCompletedQuest.CompletedSteps = append(UserCompletedQuest.CompletedSteps, UserCompletedSteps{key, value.count, value.bonus})
		}

		CompletedStepsCount++
	}
	//endregion

	UserCompletedQuest.CompletedStepsCount = CompletedStepsCount

	return UserCompletedQuest
}
