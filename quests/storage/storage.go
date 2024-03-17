package storage

import (
	"encoding/json"
	"fmt"
	dbx "github.com/go-ozzo/ozzo-dbx"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Storage struct {
	DB *dbx.DB
}

// типы для выполнения шагов
// region
// NewCompleteSteps model info
// @Description NewCompleteSteps  json для отметки о выполнении шага задания пользователем
type NewCompleteSteps struct {
	CompleteSteps []CompleteStep `json:"CompleteSteps"` //Идентификатор задания
}

type CompleteStep struct {
	//TODO вообще правильно ИД пользователя не передавать, а брать из авторизации, но тогда будет сложно тестировать, с учетом того, что это тестовое задание, то будем передавать
	Stepid int `json:"stepid"` //Идентификатор шага
	Userid int `json:"userid"` //Идентификатор пользователя выполневшего шаг
}

type CompleteStepDB struct {
	Stepid int `json:"stepid"` //Идентификатор шага
	Userid int `json:"userid"` //Идентификатор пользователя выполневшего шаг
}

func (quest *CompleteStepDB) TableName() string {
	return "history"
}

func (complete *CompleteStep) ConvertToDB() (CompleteStepDB, []ErrorList) {
	var errlist []ErrorList

	completeDB := CompleteStepDB{}
	if complete.Stepid == 0 {
		errlist = append(errlist, ErrorList{"Идентификатор шага должен быть больше 0"})
	}
	if complete.Userid == 0 {
		errlist = append(errlist, ErrorList{"Идентификатор пользователя должен быть больше 0"})
	}
	completeDB.Stepid = complete.Stepid
	completeDB.Userid = complete.Userid

	if len(errlist) > 0 {
		return completeDB, errlist
	} else {
		return completeDB, nil
	}
}

//endregion

// NewQuest model info
// @Description NewQuest json для создания задания с шагами
type NewQuest struct {
	Id         int            `json:"id"`         //идентификатор задания
	Name       string         `json:"Name"`       //Имя задания
	Cost       int            `json:"Cost"`       //Стоимость задания
	QuestSteps []NewQuestStep `json:"QuestSteps"` //Шаги задания
}

type NewQuestDB struct {
	Id   int    `json:"id" db:"id"`          //идентификатор задания
	Name string `json:"Name" db:"questname"` //Имя задания
	Cost int    `json:"Cost" db:"cost"`      //Стоимость задания
}

func (quest *NewQuestDB) TableName() string {
	return "quests"
}

func (quest *NewQuest) ConvertToDB() (NewQuestDB, []ErrorList) {
	var errlist []ErrorList

	questdb := NewQuestDB{}
	questdb.Id = quest.Id

	if quest.Name == "" {
		errlist = append(errlist, ErrorList{"Имя задания должно содержать от 1 до 200 символов"})
	}
	if quest.Cost <= 0 {
		errlist = append(errlist, ErrorList{"Cтоимость задания должна быть больше 0"})
	}
	questdb.Name = quest.Name
	questdb.Cost = quest.Cost

	if len(errlist) > 0 {
		return questdb, errlist
	} else {
		return questdb, nil
	}
}

// NewQuestSteps model info
// @Description NewQuestStep json для создания шага задания
type NewQuestSteps struct {
	QuestSteps []NewQuestStep `json:"QuestSteps"` //Идентификатор задания
}

type NewQuestStep struct {
	Id       int    `json:"id"`       //Идентификатор задания
	QuestId  int    `json:"QuestId"`  //Идентификатор задания. При создании методом CreateQuest, значение будет проигнорировано, т.к. будет подставляться идентификатор создаваемого задания
	StepName string `json:"StepName"` //Описание шага
	Bonus    int    `json:"Bonus"`    //Бонус за задание
	IsMulti  *bool  `json:"IsMulti"`  //Признак того, что шаг можно выполнять несколько раз
}

type UpdateQuestSteps struct {
	QuestSteps []UpdateQuestStep `json:"QuestSteps"` //Идентификатор задания
}

type UpdateQuestStep struct {
	Id      int   `json:"id"`      //Идентификатор задания
	Bonus   int   `json:"Bonus"`   //Бонус за задание
	IsMulti *bool `json:"IsMulti"` //Признак того, что шаг можно выполнять несколько раз
}

func (questStep *UpdateQuestStep) ConvertToDB() (NewQuestStepDB, []ErrorList) {
	var errlist []ErrorList

	questStepDB := NewQuestStepDB{}
	questStepDB.Id = questStep.Id

	if questStep.Id == 0 {
		errlist = append(errlist, ErrorList{"Не указан идентификатор шага, который необходимо обновить"})
	}
	if questStep.IsMulti == nil {
		errlist = append(errlist, ErrorList{"Укажите признак многократного выполнения"})
	}
	questStepDB.Bonus = questStep.Bonus
	questStepDB.IsMulti = *questStep.IsMulti

	if len(errlist) > 0 {
		return questStepDB, errlist
	} else {
		return questStepDB, nil
	}
}

type NewQuestStepDB struct {
	Id       int    `json:"id" db:"id"`
	QuestId  int    `json:"QuestId" db:"questid"`
	StepName string `json:"StepName" db:"stepname"`
	Bonus    int    `json:"Bonus" db:"bonus"`
	IsMulti  bool   `json:"IsMulti" db:"ismulti"`
}

func (quest *NewQuestStepDB) TableName() string {
	return "queststeps"
}

// GetUpdatesData Функция возвращает Мар содержащую только измененные поля, которые необходимо записать в БД. Если поле не было передано для обновления, то и записываться в базу оно не будет
func (questStep *NewQuestStepDB) GetUpdatesData() map[string]interface{} {
	var data = make(map[string]interface{})
	if questStep.Bonus > 0 {
		data["bonus"] = questStep.Bonus
	}
	data["ismulti"] = questStep.IsMulti
	return data
}

func (questStep *NewQuestStep) ConvertToDB() (NewQuestStepDB, []ErrorList) {
	var errlist []ErrorList

	questStepDB := NewQuestStepDB{}
	questStepDB.Id = questStep.Id

	if questStep.StepName == "" {
		errlist = append(errlist, ErrorList{"Не указано описание шага"})
	}
	if questStep.QuestId <= 0 {
		errlist = append(errlist, ErrorList{"Не указан идентификатор задания, к которому относится шаг"})
	}

	if questStep.Bonus < 0 {
		errlist = append(errlist, ErrorList{"Бонус не может быть меньше 0"})
	}

	if questStep.IsMulti == nil {
		questStepDB.IsMulti = false
	}

	questStepDB.QuestId = questStep.QuestId
	questStepDB.Bonus = questStep.Bonus
	questStepDB.StepName = questStep.StepName

	if len(errlist) > 0 {
		return questStepDB, errlist
	} else {
		return questStepDB, nil
	}
}

type ErrorList struct {
	Error string
}

func HttpResponse(w http.ResponseWriter, status int, text string) {
	w.WriteHeader(status)
	result, _ := json.MarshalIndent(text, "", "\t")
	w.Write([]byte(result))
}

func HttpResponseObject(w http.ResponseWriter, status int, text []byte) {
	w.WriteHeader(status)
	w.Write(text)
}

// New возвразает соединение с БД
func New(storagePath string) (*Storage, error) {

	db, err := dbx.MustOpen("postgres", storagePath)

	if err != nil {
		return nil, err
	}
	return &Storage{DB: db}, nil
}

// Init иницилизирует БД
func (storage *Storage) Init() error {
	//region Создаем таблицу Пользователей, индекс и пользователя администратора
	queryText := `CREATE TABLE IF NOT EXISTS users (
								id integer PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,
								userName varchar(20) NOT NULL,
								password varchar(20) NOT NULL,
								isAdmin boolean	 NOT NULL
								)`
	q := storage.DB.NewQuery(queryText)
	_, err := q.Execute()
	if err != nil {
		return fmt.Errorf("create table 'Users' complete with error: %s", err.Error())
	}

	//проверяем существует ли администратор, если нет - создаем.
	queryText = `Select count(*) as count from USERS where username ='admin'`
	q = storage.DB.NewQuery(queryText)
	resultRows, err := q.Rows()
	if err != nil {
		return fmt.Errorf("check Admin user complete with error: %s", err)
	}
	count := 0
	for resultRows.Next() {
		resultRows.Scan(&count)
	}

	if count == 0 {
		queryText = "INSERT INTO USERS(username, password, isAdmin) VALUES ( 'admin', " + "'" + EncodePassword("123") + "', true )"
		q = storage.DB.NewQuery(queryText)
		_, err = q.Execute()
		if err != nil {
			return fmt.Errorf("create Admin user complete with error: %s", err)
		}
	}

	//endregion

	//region TODO Создаем таблицу Заданий
	queryText = `CREATE TABLE IF NOT EXISTS quests (
								id integer GENERATED BY DEFAULT AS IDENTITY,
								questName varchar(200) NOT NULL,
    							PRIMARY KEY (questName)
								)`
	q = storage.DB.NewQuery(queryText)
	_, err = q.Execute()
	if err != nil {
		return fmt.Errorf("create table 'quests' complete with error: %s", err.Error())
	}
	//endregion

	//region TODO Создаем таблицу questSteps
	queryText = `CREATE TABLE IF NOT EXISTS questSteps (
								id integer PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,
								questID integer NOT NULL,
								stepName varchar(200) NOT NULL,
								bonus integer,
    							isMulti bool NOT NULL
								)`
	q = storage.DB.NewQuery(queryText)
	_, err = q.Execute()
	if err != nil {
		return fmt.Errorf("create questSteps 'films' complete with error: %s", err.Error())
	}

	//region TODO Создаем таблицу history
	queryText = `CREATE TABLE IF NOT EXISTS history (
								stepId integer NOT NULL,
								userId integer NOT NULL
								)`
	q = storage.DB.NewQuery(queryText)
	_, err = q.Execute()
	if err != nil {
		return fmt.Errorf("create table 'history' complete with error: %s", err.Error())
	}
	//endregion

	return nil
}

// EncodePassword функция возвращает хэш для пароля
func EncodePassword(password string) string {
	// Для пэт проекта сделаем просто, добавим в конец к паролю @1 и сдвинем каждый символы на 1
	passwordNew := password + "@1"
	bs := []byte(passwordNew)
	for i := range bs {
		bs[i] = bs[i] + 1
	}
	return string(bs)
}

// DecodePassword Функция возвращает пароль по хэшу
func DecodePassword(password string) string {
	bs := []byte(password)
	for i := range bs {
		bs[i] = bs[i] - 1
	}
	pass := string(bs)
	return pass[0:strings.LastIndex(pass, "@1")]
}

// TODO Тело запроса
func RequestTolog(r *http.Request, logger *slog.Logger) {
	username, _, ok := r.BasicAuth()
	if !ok {
		username = ""
	}
	logger.Debug(
		"incoming request",
		"url", r.URL.Path,
		"method", r.Method,
		"user", username,
		"time", time.Now().Format("02-01-2006 15:04:05"),
	)
}

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

// ExicuteCountSumQuery Обертка, которая предназначена для запросов, которые возвращают одно целое значение
func ExicuteCountSumQuery(storage *Storage, queryText string) int {
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

//TODO нужен еще запрос, которые показывает вообще все задания и шаги
