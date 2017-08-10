package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pborman/uuid"
)

const (
	StatisticTimeOut        = 5
	ClientCount             = 10
	WorkerCount             = 2
	ProblemsFrequency       = 50 * time.Millisecond
	ResponseTimeout         = 100 * time.Millisecond
	DelayAfterTaskExecution = 20 * time.Millisecond
	MainProcessTimeOut      = 1000 * time.Millisecond
	ThresholdFailureRate    = 0.2
	ThresholdWorkerRate     = 0.5
	DirName                 = "20_newsgroup/"
)

type Task struct {
	fileName  string
	wordCount int
	status    Status
}
type Status int

const (
	PENDING Status = iota
	PROCESSING
	DONE
)

type TaskManager struct {
	tasks map[string]Task
	done  chan bool
}

type MainThread struct {
	taskCount        int
	doneTaskCount    int
	failedTaskCount  int
	activeWorkerTime time.Duration
	workerCount      int
	clientCount      int
	clientActivity   time.Duration
}

func getAllFilenames(dirName string) []string {
	fileList := []string{}
	filepath.Walk(dirName, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return nil
	})
	return fileList
}

func getRandomInt(hightBound int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Intn(hightBound)
}

func getRandomFilename() string {
	files := getAllFilenames(DirName)
	index := getRandomInt(len(files))
	return files[index]
}

func getUniqueTaskId() string {
	return uuid.New()
}

func getWordCount(currentFilename string) int {
	b, err := ioutil.ReadFile(currentFilename)
	if err == nil {
		str := string(b)
		words := strings.Fields(str)
		return len(words)
	}
	return 0

}

func generateTask(taskManager *TaskManager, mainThread *MainThread, quit chan int) {
	for {
		select {
		case <-quit:
			return
		default:
			key := getUniqueTaskId()
			filename := getRandomFilename()
			taskManager.done <- true
			taskManager.tasks[key] = Task{filename, 0, PENDING}
			<-taskManager.done
			mainThread.taskCount++
			time.Sleep(ResponseTimeout)

			taskManager.done <- true
			task := taskManager.tasks[key]
			if task.status != DONE {
				mainThread.failedTaskCount++
			} else {
				mainThread.doneTaskCount++
			}
			delete(taskManager.tasks, key)
			<-taskManager.done
			time.Sleep(mainThread.clientActivity)
		}
	}
}
func processTask(taskManager *TaskManager, mainThread *MainThread, quit chan int) {
	var startProcessing time.Time
	var endProcessing time.Time
	for {
		select {
		case <-quit:
			return
		default:
			startProcessing = time.Now()
			taskManager.done <- true
			key, task := changeStatusInPendingTask(taskManager)
			<-taskManager.done
			if key != "notFound" {
				task.wordCount = getWordCount(task.fileName)
				task.status = DONE

				taskManager.done <- true
				taskManager.tasks[key] = task
				<-taskManager.done
				time.Sleep(DelayAfterTaskExecution)
			}
			endProcessing = time.Now()
			if task.status == DONE {
				mainThread.activeWorkerTime += endProcessing.Sub(startProcessing)
			}
		}
	}
}
func changeStatusInPendingTask(taskManager *TaskManager) (string, Task) {
	for key, value := range taskManager.tasks {
		if value.status == PENDING {
			task := taskManager.tasks[key]
			task.status = PROCESSING
			taskManager.tasks[key] = task
			return key, value
		}
	}
	return "notFound", Task{}
}
func manager(taskManager *TaskManager, mainThread *MainThread, workerQuit chan int, statistics *MainThread) {
	i := 0
	for {
		i++
		time.Sleep(MainProcessTimeOut)
		if i == StatisticTimeOut {
			printStatistics(statistics)
			resetStatistics(statistics)
			i = 0
		}
		if float64(mainThread.failedTaskCount) > float64(mainThread.taskCount)*ThresholdFailureRate {
			mainThread.workerCount++
			go processTask(taskManager, mainThread, workerQuit)
		}
		if mainThread.workerCount > 1 && (mainThread.activeWorkerTime.Seconds()/float64(mainThread.workerCount) <= ThresholdWorkerRate) {
			mainThread.workerCount--
			workerQuit <- 0
		}
		refreshStatistics(statistics, mainThread)
		resetCounters(mainThread)
	}
}
func resetStatistics(statistics *MainThread) {
	statistics.clientActivity = 0
	statistics.clientCount = 0
	statistics.doneTaskCount = 0
	statistics.failedTaskCount = 0
	statistics.doneTaskCount = 0
	statistics.taskCount = 0
	statistics.workerCount = 0
	statistics.activeWorkerTime = 0
}
func refreshStatistics(statistics, mainThread *MainThread) {
	statistics.clientCount = mainThread.clientCount
	statistics.workerCount = mainThread.workerCount
	statistics.clientActivity = mainThread.clientActivity
	statistics.taskCount += mainThread.taskCount
	statistics.doneTaskCount += mainThread.doneTaskCount
	statistics.failedTaskCount += mainThread.failedTaskCount
}
func printStatistics(statistics *MainThread) {
	fmt.Println("Statistics")
	fmt.Println("Worker count =", statistics.workerCount)
	fmt.Println("Client count =", statistics.clientCount)
	fmt.Println("Client activity =", statistics.clientActivity)
	fmt.Println("Generated task count =", statistics.taskCount)
	fmt.Println("Done task count =", statistics.doneTaskCount)
	fmt.Println("Failed task count =", statistics.failedTaskCount)
	if statistics.taskCount > 0 {
		fmt.Println("Done task percent =", statistics.doneTaskCount*100/statistics.taskCount)
		fmt.Println("Failed task percent =", statistics.failedTaskCount*100/statistics.taskCount)
	} else {
		fmt.Println("Done task percent =", 0)
		fmt.Println("Failed task percent =", 0)
	}
	fmt.Println("**************************************************")
}

func resetCounters(mainThread *MainThread) {
	mainThread.activeWorkerTime = 0
	mainThread.doneTaskCount = 0
	mainThread.failedTaskCount = 0
	mainThread.taskCount = 0
}

func simulator(taskManager *TaskManager) {
	for {
		<-taskManager.done
		taskManager.done <- true
	}
}
func clientManager(taskManager *TaskManager, mainThread *MainThread, clientQuit chan int) {
	var input string
	for {
		fmt.Scan(&input)
		switch input {
		case "+":
			mainThread.clientCount++
			go generateTask(taskManager, mainThread, clientQuit)
			fmt.Println("The client count has been increased by 1. Now it is", mainThread.clientCount)
		case "-":
			mainThread.clientCount--
			clientQuit <- 0
			fmt.Println("The client count has been reduced by 1. Now it is", mainThread.clientCount)
		case "<":
			newActivity := time.Duration(mainThread.clientActivity.Nanoseconds() - mainThread.clientActivity.Nanoseconds()*10/100)
			mainThread.clientActivity = newActivity
			fmt.Println("The client activity has been reduced by 10%. Now it is", mainThread.clientActivity)
		case ">":
			newActivity := time.Duration(mainThread.clientActivity.Nanoseconds() + mainThread.clientActivity.Nanoseconds()*10/100)
			mainThread.clientActivity = newActivity
			fmt.Println("The client activity has been increased by 10%. Now it is", mainThread.clientActivity)
		}
	}
}
func main() {
	taskManager := TaskManager{make(map[string]Task), make(chan bool)}
	mainThread := MainThread{0, 0, 0, 0, WorkerCount, ClientCount, ProblemsFrequency}
	statistics := MainThread{0, 0, 0, 0, 0, 0, 0}
	workerQuit := make(chan int, 1)
	clientQuit := make(chan int, 1)
	for i := 0; i < ClientCount; i++ {
		go generateTask(&taskManager, &mainThread, clientQuit)
	}
	for i := 0; i < WorkerCount; i++ {
		go processTask(&taskManager, &mainThread, workerQuit)
	}
	go simulator(&taskManager)
	go manager(&taskManager, &mainThread, workerQuit, &statistics)
	clientManager(&taskManager, &mainThread, clientQuit)
}
