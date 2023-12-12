package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var (
	commandsExecuted []ExecutedCommand
	mutex            sync.Mutex
	whitelist        map[string]struct{}
	port             int
)

type ExecutedCommand struct {
	Command string    `json:"command"`
	Output  string    `json:"output"`
	Time    time.Time `json:"time"`
}

func init() {
	flag.IntVar(&port, "port", 8080, "Port number for the server")
	flag.Parse()
}

func loadWhitelist() {
	whitelist = make(map[string]struct{})

	file, err := os.Open("/usr/local/etc/gorun.conf")
	if err != nil {
		log.Fatal("Error opening whitelist file:", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		command := scanner.Text()
		whitelist[command] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("Error reading whitelist file:", err)
	}
}

func runCommandHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method. Only POST is allowed.", http.StatusMethodNotAllowed)
		return
	}

	loadWhitelist()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		log.Println("Error reading request body:", err)
		return
	}
	defer r.Body.Close()

	cmd := string(body)

	// 按空格分割命令
	cmdParts := strings.Fields(cmd)

	// 获取命令的名称（第一个部分）
	commandName := cmdParts[0]

	// 检查命令名称是否在白名单内
	if _, ok := whitelist[commandName]; !ok {
		http.Error(w, "Command not allowed.", http.StatusForbidden)
		log.Printf("Command not allowed: %s\n", cmd)
		return
	}

	startTime := time.Now()
	output, err := exec.Command("sh", "-c", cmd).Output()

	if err != nil {
		http.Error(w, fmt.Sprintf("Error executing command: %s", err), http.StatusInternalServerError)
		log.Printf("Error executing command: %s\n", err)
		return
	}

	// 记录参数和输出到日志文件
	log.Printf("Received command: %s\n", cmd)
	log.Printf("Command output: %s\n", output)

	// 记录执行的命令
	mutex.Lock()
	commandsExecuted = append(commandsExecuted, ExecutedCommand{
		Command: cmd,
		Output:  string(output),
		Time:    startTime,
	})
	mutex.Unlock()

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write(output)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	var executedCommandsWithTime []ExecutedCommand
	for _, cmd := range commandsExecuted {
		executedCommandsWithTime = append(executedCommandsWithTime, ExecutedCommand{
			Command: cmd.Command,
			Output:  cmd.Output,
			Time:    cmd.Time,
		})
	}

	result, err := json.Marshal(executedCommandsWithTime)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		http.Error(w, "Error generating JSON response", http.StatusInternalServerError)
		return
	}

	w.Write(result)
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Simple Command Execution Service")
	fmt.Fprintln(w, "Available Endpoints:")
	fmt.Fprintln(w, "- POST /run: Execute a command")
	fmt.Fprintln(w, "- GET /stats: Get executed commands statistics")
}

func main() {
	// 打开日志文件，创建一个带有时间戳的新日志文件
	logFile, err := os.OpenFile("/var/log/gorun.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	defer logFile.Close()

	// 设置日志输出到文件
	log.SetOutput(logFile)

	http.HandleFunc("/run", runCommandHandler)
	http.HandleFunc("/stats", statsHandler)
	http.HandleFunc("/", infoHandler)

	fmt.Printf("Server is running on :%d\n", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Printf("Error starting the server: %s\n", err)
	}
}
