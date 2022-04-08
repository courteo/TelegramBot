package main

// сюда писать код

// https://api.telegram.org/bot5244227470:AAEModcsPOS8TxZehTmFoTwH5Kr3mctcMv0/getUpdates

import (
	//"gopkg.in/telegram-bot-api.v4"
	"context"
	"fmt"
	tgbotapi "github.com/skinass/telegram-bot-api/v5"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	// @BotFather в телеграме даст вам это
	BotToken = "5244227470:AAEModcsPOS8TxZehTmFoTwH5Kr3mctcMv0"

	// урл выдаст вам игрок или хероку
	WebhookURL = "https://telegrambotforgolang.herokuapp.com"
)

var AllTasks []Task
var AllUsers []User
var Inc int

func GetTaskId(id int) (int, error) {
	for i, task := range AllTasks {
		if task.Id == id {
			return i, nil
		}
	}
	err := fmt.Errorf("нет такой задачи")
	return -1, err
}

func IsTaskContain(TaskName string) bool {
	for _, task := range AllTasks {
		if task.Name == TaskName {
			return true
		}
	}
	return false
}

func GetIndexOfUser(UserName string) (int, error) {
	for i, user := range AllUsers {
		if user.UserName == UserName {
			return i, nil
		}
	}
	err := fmt.Errorf("нет пользователя")
	return -1, err
}

func GetUserId(UserName string) (int, error) {
	for i, user := range AllUsers {
		if user.UserName == UserName {
			return i, nil
		}
	}
	err := fmt.Errorf("нет пользователя")
	return -1, err
}

func GetUser(UserName string) (User, error) {
	for _, user := range AllUsers {
		if user.UserName == UserName {
			return user, nil
		}
	}
	err := fmt.Errorf("нет пользователя")
	return User{}, err
}

type User struct {
	UserName     string
	CreatedTasks []Task // Которые он создал
	UserTasks    []Task // Которые ему задали
	ChatId       int64
}

func (user *User) AddNewTask(NewTask Task) {
	user.CreatedTasks = append(user.CreatedTasks, NewTask)
}

func (user *User) DeleteTask(taskName string) {
	index := -1
	for i, task := range user.UserTasks {
		if task.Name == taskName {
			index = i
			break
		}
	}
	if index != -1 {
		user.UserTasks = append(user.UserTasks[:index], user.UserTasks[index+1:]...)
	}
}

func (user *User) DeleteCreatedTask(taskName string) {
	index := -1
	for i, task := range user.CreatedTasks {
		if task.Name == taskName {
			index = i
			break
		}
	}
	if index != -1 {
		user.CreatedTasks = append(user.CreatedTasks[:index], user.CreatedTasks[index+1:]...)
	}
}

func (user User) IsUserHasTask(TaskName string) bool {
	for _, userTask := range user.UserTasks {
		if userTask.Name == TaskName {
			return true
		}
	}
	return false
}

type Task struct {
	Name    string
	Owner   string
	Creator string
	Id      int
}

func PrintTaskWithOwner(CurrTask Task) string {
	return strconv.Itoa(CurrTask.Id) + ". " + CurrTask.Name + " by @" + CurrTask.Creator + "\n" +
		"/unassign_" + strconv.Itoa(CurrTask.Id) + " /resolve_" + strconv.Itoa(CurrTask.Id)
}

func PrintTaskWithoutOwner(CurrTask Task) string {
	return strconv.Itoa(CurrTask.Id) + ". " + CurrTask.Name + " by @" + CurrTask.Creator + "\n" +
		"/assign_" + strconv.Itoa(CurrTask.Id)
}

func NewTask(TaskName string, Owner User) (res string) {
	if TaskName == "" {
		return "Название задачи не может быть пустой"
	}

	if IsTaskContain(TaskName) {
		return "the \"" + TaskName + "\" task already exists"
	}

	Inc++
	newTask := Task{
		Name:    TaskName,
		Creator: Owner.UserName,
		Id:      Inc,
	}
	Owner.AddNewTask(newTask)
	AllTasks = append(AllTasks, newTask)

	index, err := GetIndexOfUser(Owner.UserName)
	if err == nil {
		AllUsers = append(AllUsers[:index], AllUsers[index+1:]...)
	}
	AllUsers = append(AllUsers, Owner)

	//fmt.Println(AllUsers)
	return "Задача \"" + TaskName + "\" создана, id=" + strconv.Itoa(Inc)
}

func MyTask(user User) (res string) {
	for i, userTask := range user.UserTasks {
		res += PrintTaskWithOwner(userTask)
		if i != len(user.UserTasks)-1 {
			res += "\n"
		}
	}
	if len(user.UserTasks) == 0 {
		return "на вас нет задач"
	}
	return res
}

func OwnerTask(user User) (res string) {
	if len(user.CreatedTasks) == 0 {
		return "вы не создали задачи"
	}

	for i, userTask := range user.CreatedTasks {
		//fmt.Println("Task ", userTask)
		if userTask.Owner != "" {
			res += PrintTaskWithOwner(userTask)
		} else {
			res += PrintTaskWithoutOwner(userTask)
		}

		if i != (len(user.CreatedTasks) - 1) {
			res += "\n"
		}
	}
	//fmt.Println("result ", res, user)
	return res
}

func Assign(user User, id int) (res []string, chatId []int64, err error) {
	taskId, errorID := GetTaskId(id)
	if errorID != nil {
		err = fmt.Errorf("нет такой задачи")
		return []string{}, []int64{}, err
	}

	if AllTasks[taskId].Owner != "" {
		userId, errorUserID := GetUserId(AllTasks[taskId].Owner)
		if errorUserID != nil {
			return []string{}, []int64{}, errorUserID
		}

		AllUsers[userId].DeleteTask(AllTasks[taskId].Name)
		str := "Задача \"" + AllTasks[taskId].Name + "\" назначена на @" + user.UserName // сообщение новому владельцу задачи
		res = append(res, str)
		chatId = append(chatId, AllUsers[userId].ChatId)
	} else if AllTasks[taskId].Creator != user.UserName {
		userId, errorUserID := GetUserId(AllTasks[taskId].Creator)
		if errorUserID != nil {
			return []string{}, []int64{}, errorUserID
		}

		AllUsers[userId].DeleteTask(AllTasks[taskId].Name)
		str := "Задача \"" + AllTasks[taskId].Name + "\" назначена на @" + user.UserName // сообщение новому владельцу задачи
		res = append(res, str)
		chatId = append(chatId, AllUsers[userId].ChatId)
	}

	AllTasks[taskId].Owner = user.UserName

	userId, errorUserID := GetUserId(user.UserName)
	if errorUserID != nil {
		err = fmt.Errorf("")
		return []string{}, []int64{}, err
	}

	if !user.IsUserHasTask(AllTasks[taskId].Name) {
		AllUsers[userId].UserTasks = append(AllUsers[userId].UserTasks, AllTasks[taskId])
	}

	str := "Задача \"" + AllTasks[taskId].Name + "\" назначена на вас" // сообщение новому владельцу задачи
	res = append(res, str)
	chatId = append(chatId, user.ChatId)

	return res, chatId, nil
}

func UnAssign(user User, id int) (res []string, chatId []int64, err error) {
	taskId, errorID := GetTaskId(id)
	if errorID != nil {
		err = fmt.Errorf("нет такой задачи")
		return []string{}, []int64{}, err
	}

	if !user.IsUserHasTask(AllTasks[taskId].Name) {
		res = append(res, "Задача не на вас")
		chatId = append(chatId, user.ChatId)
		return res, chatId, nil
	}

	AllTasks[taskId].Owner = ""
	userId, errorUserID := GetUserId(user.UserName)
	if errorUserID != nil {
		return []string{}, []int64{}, errorUserID
	}

	AllUsers[userId].DeleteTask(AllTasks[taskId].Name)
	str := "Принято" // сняли задачу с пользователя
	res = append(res, str)
	chatId = append(chatId, AllUsers[userId].ChatId)

	userId, errorUserID = GetUserId(AllTasks[taskId].Creator)
	if errorUserID != nil {
		return []string{}, []int64{}, errorUserID
	}

	AllUsers[userId].DeleteTask(AllTasks[taskId].Name)
	str = "Задача \"" + AllTasks[taskId].Name + "\" осталась без исполнителя" // сообщение автору задачи

	res = append(res, str)
	chatId = append(chatId, AllUsers[userId].ChatId)
	return res, chatId, nil
}

func Resolve(user User, id int) (res []string, chatId []int64, err error) {
	taskId, errorID := GetTaskId(id)
	if errorID != nil {
		err = fmt.Errorf("нет такой задачи")
		return []string{}, []int64{}, err
	}

	owner, errorUser := GetUserId(AllTasks[taskId].Owner)
	if errorUser != nil {
		errorUser = fmt.Errorf("Нет пользователя, которому задали эту задачу")
		return []string{}, []int64{}, errorUser
	}

	if AllUsers[owner].UserName != user.UserName {
		err = fmt.Errorf("у вас нет доступка к этому")
		return []string{}, []int64{}, err
	}
	AllUsers[owner].DeleteTask(AllTasks[taskId].Name) // удаляем задачу у исполнителя
	str := "Задача \"" + AllTasks[taskId].Name + "\" выполнена"
	res = append(res, str)
	chatId = append(chatId, AllUsers[owner].ChatId)

	creator, errorUser := GetUserId(AllTasks[taskId].Creator)
	if errorUser != nil {
		return []string{}, []int64{}, errorUser
	}

	AllUsers[creator].DeleteCreatedTask(AllTasks[taskId].Name) // удаляем задачу у создателя
	str = "Задача \"" + AllTasks[taskId].Name + "\" выполнена @" + AllUsers[owner].UserName
	res = append(res, str)
	chatId = append(chatId, AllUsers[creator].ChatId)

	AllTasks = append(AllTasks[:taskId], AllTasks[taskId+1:]...)
	//fmt.Println(AllTasks)
	return res, chatId, nil
}

func PrintAllTasks(user User) (res string, err error) {
	if len(AllTasks) == 0 {
		err = fmt.Errorf("Нет задач")
		return "", err
	}

	for i, task := range AllTasks {
		str := strconv.Itoa(task.Id) + ". " + task.Name + " by @" + task.Creator + "\n"
		if task.Owner != "" { // задачу кто-то взял
			if task.Owner == user.UserName {
				str += "assignee: я\n"
				str += "/unassign_" + strconv.Itoa(task.Id) + " /resolve_" + strconv.Itoa(task.Id)
			} else {
				str += "assignee: @" + task.Owner
			}

		} else { // задачу никто не взял
			str += "/assign_" + strconv.Itoa(task.Id)
		}
		res += str
		if i != len(AllTasks)-1 {
			res += "\n" + "\n"
		}
	}
	return res, nil
}

func BotSend(bot tgbotapi.BotAPI, currUser User, taskId int, update tgbotapi.Update, name string) {
	var msgs []string
	var chatId []int64
	var err error
	switch name {
	case "assign":
		msgs, chatId, err = Assign(currUser, taskId)
	case "unassign":
		msgs, chatId, err = UnAssign(currUser, taskId)
	case "resolve":
		msgs, chatId, err = Resolve(currUser, taskId)
	}

	if err != nil {
		bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			err.Error(),
		))
		return
	}
	//fmt.Println(msgs, chatId)
	if len(msgs) == 2 {
		bot.Send(tgbotapi.NewMessage(
			chatId[0],
			msgs[0],
		))

		bot.Send(tgbotapi.NewMessage(
			chatId[1],
			msgs[1],
		))

	} else {
		bot.Send(tgbotapi.NewMessage(
			chatId[0],
			msgs[0],
		))
	}
}

func Help(bot tgbotapi.BotAPI, currUser User, update tgbotapi.Update) {
	str := "Существующие команды:\n \t /tasks - выводит текущие задачи\n \t /new XXX - вы создаете новую задачу\n" +
		"\t /assign_$ID  - назначаете пользователя исполнителем задачи\n \t /unassign_$ID - снимаете задачу с текущего пользователя\n" +
		"\t /resolve_$ID - выполняется задача\n \t /my - выводит задачи, которые назначили на меня\n \t /owner - показывает задачи, созданные мной"

	bot.Send(tgbotapi.NewMessage(
		update.Message.Chat.ID,
		str,
	))
}

func startTaskBot(ctx context.Context) error {
	rand.Seed(time.Now().UnixNano())
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		log.Fatalf("NewBotAPI failed: %s", err)
	}

	bot.Debug = true
	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	wh, err := tgbotapi.NewWebhook(WebhookURL)
	if err != nil {
		log.Fatalf("NewWebhook failed: %s", err)
	}

	_, err = bot.Request(wh)
	if err != nil {
		log.Fatalf("SetWebhook failed: %s", err)
	}

	updates := bot.ListenForWebhook("/")

	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("all is working"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	go func() {
		log.Fatalln("http err:", http.ListenAndServe(":"+port, nil))
	}()
	fmt.Println("start listen :" + port)

	// получаем все обновления из канала updates
	for update := range updates {
		currUser, err := GetUser(update.Message.From.UserName)
		if err != nil {
			currUser = User{UserName: update.Message.From.UserName, ChatId: update.Message.Chat.ID}
			AllUsers = append(AllUsers, currUser)
		}

		if update.Message.IsCommand() {
			var msg, command, body string
			var taskId int
			index := strings.Index(update.Message.Text, " ")

			if index != -1 {
				command = update.Message.Text[1:index]
				body = update.Message.Text[index+1:]

			} else {
				command = update.Message.Text[1:]

				taskIdTemp := strings.Index(command, "_")

				//fmt.Println("ASSIGN  ", taskIdTemp, command)
				if taskIdTemp != -1 {
					taskId, _ = strconv.Atoi(command[taskIdTemp+1:])
					command = command[:taskIdTemp]
					//fmt.Println("ASSIGN  ", taskIdTemp, command)

				}
				//fmt.Println("ASSIGN  ", taskId, command)
			}

			switch command {
			case "new":
				msg = NewTask(body, currUser)
				bot.Send(tgbotapi.NewMessage(
					update.Message.Chat.ID,
					msg,
				))
			case "my":
				msg = MyTask(currUser)
				bot.Send(tgbotapi.NewMessage(
					update.Message.Chat.ID,
					msg,
				))
			case "owner":
				msg = OwnerTask(currUser)
				bot.Send(tgbotapi.NewMessage(
					update.Message.Chat.ID,
					msg,
				))
			case "assign":
				BotSend(*bot, currUser, taskId, update, "assign")
			case "unassign":
				BotSend(*bot, currUser, taskId, update, "unassign")
			case "resolve":
				BotSend(*bot, currUser, taskId, update, "resolve")
			case "tasks":
				msg, err := PrintAllTasks(currUser)
				if err != nil {
					msg = err.Error()
				}

				bot.Send(tgbotapi.NewMessage(
					update.Message.Chat.ID,
					msg,
				))
			case "start":
				bot.Send(tgbotapi.NewMessage(
					update.Message.Chat.ID,
					"Введите /help"))
			case "help":
				Help(*bot, currUser, update)
			default:
				msg = "none"
			}

			if msg == "none" {
				bot.Send(tgbotapi.NewMessage(
					update.Message.Chat.ID,
					"Команды не существует",
				))
			}

		} else {
			bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				"Привет, напиши /help для команд",
			))
		}
	}
	return nil
}

func main() {
	err := startTaskBot(context.Background())
	if err != nil {
		panic(err)
	}
}
