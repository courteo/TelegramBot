package main

// сюда писать код

// https://api.telegram.org/bot5244227470:AAEModcsPOS8TxZehTmFoTwH5Kr3mctcMv0/getUpdates

import (
	"context"
	"fmt"
	tgbotapi "github.com/skinass/telegram-bot-api/v5"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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

func IsTaskContain(taskName string) bool {
	for _, task := range AllTasks {
		if task.Name == taskName {
			return true
		}
	}
	return false
}

func GetUserId(userName string) (int, error) {
	for i, user := range AllUsers {
		if user.UserName == userName {
			return i, nil
		}
	}
	err := fmt.Errorf("нет пользователя")
	return -1, err
}

func GetUser(userName string) (User, error) {
	for _, user := range AllUsers {
		if user.UserName == userName {
			return user, nil
		}
	}
	err := fmt.Errorf("нет пользователя")
	return User{}, err
}

type User struct {
	UserName     string
	CreatedTasks []int // Которые он создал
	UserTasks    []int // Которые ему задали
	ChatId       int64
}

func (user *User) AddNewTask(newTask Task) {
	user.CreatedTasks = append(user.CreatedTasks, newTask.Id)
}

func (user *User) DeleteTask(taskName int) {
	index := -1
	for i, task := range user.UserTasks {
		if task == taskName {
			index = i
			break
		}
	}
	if index != -1 {
		user.UserTasks = append(user.UserTasks[:index], user.UserTasks[index+1:]...)
	}
}

func (user *User) DeleteCreatedTask(taskName int) {
	index := -1
	for i, task := range user.CreatedTasks {
		if task == taskName {
			index = i
			break
		}
	}
	if index != -1 {
		user.CreatedTasks = append(user.CreatedTasks[:index], user.CreatedTasks[index+1:]...)
	}
}

func (user User) IsUserHasTask(taskName int) bool {
	for _, userTask := range user.UserTasks {
		if userTask == taskName {
			return true
		}
	}
	return false
}

type Task struct {
	Name     string
	Assignee string
	Creator  string
	Id       int
}

func PrintTaskWithAssignee(currTask Task) string {
	return strconv.Itoa(currTask.Id) + ". " + currTask.Name + " by @" + currTask.Creator + "\n" +
		"/unassign_" + strconv.Itoa(currTask.Id) + " /resolve_" + strconv.Itoa(currTask.Id)
}

func PrintTaskWithoutAssignee(currTask Task) string {
	return strconv.Itoa(currTask.Id) + ". " + currTask.Name + " by @" + currTask.Creator + "\n" +
		"/assign_" + strconv.Itoa(currTask.Id)
}

func NewTask(taskName string, creator User) (res string) {
	if taskName == "" {
		return "Название задачи не может быть пустой"
	}

	if IsTaskContain(taskName) {
		return "the \"" + taskName + "\" task already exists"
	}

	Inc++
	newTask := Task{
		Name:    taskName,
		Creator: creator.UserName,
		Id:      Inc,
	}
	creator.AddNewTask(newTask)
	AllTasks = append(AllTasks, newTask)

	index, err := GetUserId(creator.UserName)
	if err == nil {
		AllUsers = append(AllUsers[:index], AllUsers[index+1:]...)
	}
	AllUsers = append(AllUsers, creator)

	return "Задача \"" + taskName + "\" создана, id=" + strconv.Itoa(Inc)
}

func MyTask(user User) (res string) {
	for i, userTask := range user.UserTasks {
		task, err := GetTaskId(userTask)
		if err != nil {
			return "нет такой задачи"
		}

		res += PrintTaskWithAssignee(AllTasks[task])
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
		taskId, err := GetTaskId(userTask)
		if err != nil {
			return "нет такой задачи"
		}

		if AllTasks[taskId].Assignee != "" {
			res += PrintTaskWithAssignee(AllTasks[taskId])
		} else {
			res += PrintTaskWithoutAssignee(AllTasks[taskId])
		}

		if i != (len(user.CreatedTasks) - 1) {
			res += "\n"
		}
	}
	return res
}

func Assign(user User, id int) (res []string, chatId []int64, err error) {
	taskId, errorID := GetTaskId(id)
	if errorID != nil {
		err = fmt.Errorf("нет такой задачи")
		return []string{}, []int64{}, err
	}

	if AllTasks[taskId].Assignee != "" || AllTasks[taskId].Creator != user.UserName {
		var userId int
		var errorUserID error

		if AllTasks[taskId].Assignee != "" {
			userId, errorUserID = GetUserId(AllTasks[taskId].Assignee)
		} else {
			userId, errorUserID = GetUserId(AllTasks[taskId].Creator)
		}

		if errorUserID != nil {
			return []string{}, []int64{}, errorUserID
		}

		AllUsers[userId].DeleteTask(AllTasks[taskId].Id)
		str := "Задача \"" + AllTasks[taskId].Name + "\" назначена на @" + user.UserName // сообщение новому владельцу задачи
		res = append(res, str)
		chatId = append(chatId, AllUsers[userId].ChatId)
	}
	AllTasks[taskId].Assignee = user.UserName

	userId, errorUserID := GetUserId(user.UserName)
	if errorUserID != nil {
		err = fmt.Errorf("")
		return []string{}, []int64{}, err
	}

	if !user.IsUserHasTask(AllTasks[taskId].Id) {
		AllUsers[userId].UserTasks = append(AllUsers[userId].UserTasks, AllTasks[taskId].Id)
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

	if !user.IsUserHasTask(AllTasks[taskId].Id) {
		res = append(res, "Задача не на вас")
		chatId = append(chatId, user.ChatId)
		return res, chatId, nil
	}

	AllTasks[taskId].Assignee = ""
	userId, errorUserID := GetUserId(user.UserName)
	if errorUserID != nil {
		return []string{}, []int64{}, errorUserID
	}

	AllUsers[userId].DeleteTask(AllTasks[taskId].Id)
	str := "Принято" // сняли задачу с пользователя
	res = append(res, str)
	chatId = append(chatId, AllUsers[userId].ChatId)

	userId, errorUserID = GetUserId(AllTasks[taskId].Creator)
	if errorUserID != nil {
		return []string{}, []int64{}, errorUserID
	}

	AllUsers[userId].DeleteTask(AllTasks[taskId].Id)
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

	Assignee, errorUser := GetUserId(AllTasks[taskId].Assignee)
	if errorUser != nil {
		errorUser = fmt.Errorf("Нет пользователя, которому задали эту задачу")
		return []string{}, []int64{}, errorUser
	}

	if AllUsers[Assignee].UserName != user.UserName {
		err = fmt.Errorf("у вас нет доступка к этому")
		return []string{}, []int64{}, err
	}
	AllUsers[Assignee].DeleteTask(AllTasks[taskId].Id) // удаляем задачу у исполнителя
	str := "Задача \"" + AllTasks[taskId].Name + "\" выполнена"
	res = append(res, str)
	chatId = append(chatId, AllUsers[Assignee].ChatId)

	creator, errorUser := GetUserId(AllTasks[taskId].Creator)
	if errorUser != nil {
		return []string{}, []int64{}, errorUser
	}

	if creator == Assignee {
		return res, chatId, nil
	}
	AllUsers[creator].DeleteCreatedTask(AllTasks[taskId].Id) // удаляем задачу у создателя
	str = "Задача \"" + AllTasks[taskId].Name + "\" выполнена @" + AllUsers[Assignee].UserName
	res = append(res, str)
	chatId = append(chatId, AllUsers[creator].ChatId)

	AllTasks = append(AllTasks[:taskId], AllTasks[taskId+1:]...)
	return res, chatId, nil
}

func PrintAllTasks(user User) (res string, err error) {
	if len(AllTasks) == 0 {
		err = fmt.Errorf("Нет задач")
		return "", err
	}

	for i, task := range AllTasks {
		str := strconv.Itoa(task.Id) + ". " + task.Name + " by @" + task.Creator + "\n"
		if task.Assignee != "" { // задачу кто-то взял
			if task.Assignee == user.UserName {
				str += "assignee: я\n"
				str += "/unassign_" + strconv.Itoa(task.Id) + " /resolve_" + strconv.Itoa(task.Id)
			} else {
				str += "assignee: @" + task.Assignee
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
		_, err2 := bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"нет такой задачи",
		))
		if err2 != nil {
			fmt.Println("ошибка при отправке")
			return
		}
		return
	}

	for i := range msgs {
		_, err1 := bot.Send(tgbotapi.NewMessage(
			chatId[i],
			msgs[i],
		))
		if err1 != nil {
			fmt.Println("ошибка при отправке")
			return
		}
	}
}

func Help(bot tgbotapi.BotAPI, currUser User, update tgbotapi.Update) {
	str := "Существующие команды:\n \t /tasks - выводит текущие задачи\n \t /new XXX - вы создаете новую задачу\n" +
		"\t /assign_$ID  - назначаете пользователя исполнителем задачи\n \t /unassign_$ID - снимаете задачу с текущего пользователя\n" +
		"\t /resolve_$ID - выполняется задача\n \t /my - выводит задачи, которые назначили на меня\n \t /owner - показывает задачи, созданные мной"

	_, err := bot.Send(tgbotapi.NewMessage(
		update.Message.Chat.ID,
		str,
	))
	if err != nil {
		fmt.Println("ошибка при отправке")
		return
	}
}

func ForCommand(bot tgbotapi.BotAPI, currUser User, update tgbotapi.Update) {
	var msg, command, body string
	var taskId int
	var errorConv error
	index := strings.Index(update.Message.Text, " ")
	if index != -1 {
		command = update.Message.Text[1:index]
		body = update.Message.Text[index+1:]
	} else {

		command = update.Message.Text[1:]
		taskIdTemp := strings.Index(command, "_")

		if taskIdTemp != -1 {
			taskId, errorConv = strconv.Atoi(command[taskIdTemp+1:])
			command = command[:taskIdTemp]

			if errorConv != nil {
				_, err := bot.Send(tgbotapi.NewMessage(
					update.Message.Chat.ID,
					"Следует вводить номер задачи",
				))
				if err != nil {
					fmt.Println("ошибка при отправке")
					return
				}
				return
			}
		}
	}
	var err1 error
	switch command {
	case "new":
		msg = NewTask(body, currUser)
		_, err1 = bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			msg,
		))

	case "my":
		msg = MyTask(currUser)
		_, err1 = bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			msg,
		))
	case "owner":
		msg = OwnerTask(currUser)
		_, err1 = bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			msg,
		))
	case "assign":
		BotSend(bot, currUser, taskId, update, "assign")
	case "unassign":
		BotSend(bot, currUser, taskId, update, "unassign")
	case "resolve":
		BotSend(bot, currUser, taskId, update, "resolve")
	case "tasks":
		msg1, err := PrintAllTasks(currUser)
		if err != nil {
			msg1 = "Нет задач"
		}

		_, err1 = bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			msg1,
		))
	case "start":
		_, err1 = bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"Введите /help"))

	case "help":
		Help(bot, currUser, update)
	default:
		_, err1 = bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"Команды не существует",
		))

	}
	if err1 != nil {
		fmt.Println("ошибка при отправке")
		return
	}
}

func startTaskBot(ctx context.Context) error {
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
		_, err1 := w.Write([]byte("all is working"))
		if err1 != nil {
			fmt.Println("не работает")
			return
		}
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
		if update.Message == nil {
			continue
		}

		currUser, err := GetUser(update.Message.From.UserName)
		if err != nil {
			currUser = User{UserName: update.Message.From.UserName, ChatId: update.Message.Chat.ID}
			AllUsers = append(AllUsers, currUser)
		}

		if update.Message.IsCommand() {
			ForCommand(*bot, currUser, update)
		} else {
			_, err1 := bot.Send(tgbotapi.NewMessage(
				update.Message.Chat.ID,
				"Привет, напиши /help для команд",
			))
			if err1 != nil {
				fmt.Println("ошибка при отправке")
				return err1
			}
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
