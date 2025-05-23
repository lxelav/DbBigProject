package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func handlePoolsAndSchemas(pools *AllPools, args []string) error {
	pools.ShowAll()
	switch args[0] {
	case "add-pool":
		if len(args) < 2 {
			return fmt.Errorf("недостаточно аргументов для команды add-pool")
		}
		pools.AddPools(args[1])
	case "remove-pool":
		if len(args) < 2 {
			return fmt.Errorf("недостаточно аргументов для команды remove-pool")
		}
		pools.RemovePools(args[1])
	case "add-schema":
		if len(args) < 3 {
			return fmt.Errorf("недостаточно аргументов для команды add-schema")
		}
		pool, err := pools.GetPools(args[1])
		if err != nil {
			return err
		}
		pool.AddSchema(args[2])
	case "remove-schema":
		if len(args) < 3 {
			return fmt.Errorf("недостаточно аргументов для команды remove-schema")
		}
		pool, err := pools.GetPools(args[1])
		if err != nil {
			return err
		}
		pool.RemoveSchema(args[2])
	case "add-collection":
		if len(args) < 5 {
			return fmt.Errorf("недостаточно аргументов для команды add-collection")
		}
		collectionType := args[4]
		pool, err := pools.GetPools(args[1])
		if err != nil {
			return err
		}
		treeCollection := NewTreeCollection(collectionType)
		if err = pool.AddCollection(args[2], args[3], *treeCollection); err != nil {
			return err
		}
	case "remove-collection":
		if len(args) < 4 {
			return fmt.Errorf("недостаточно аргументов для команды remove-collection")
		}
		pool, err := pools.GetPools(args[1])
		if err != nil {
			return err
		}
		schema, err := pool.GetSchema(args[2])
		if err != nil {
			return err
		}
		schema.RemoveCollection(args[3])
	default:
		return fmt.Errorf("неизвестная команда")
	}
	return nil
}

func RunCommand(pools *AllPools, command string, cr *ChainOfResponsibility) error {
	args := strings.Fields(command)
	if len(args) == 0 {
		return fmt.Errorf("не указана команда")
	}

	switch args[0] {
	case "add-pool", "remove-pool", "add-schema", "remove-schema", "add-collection", "remove-collection":
		return handlePoolsAndSchemas(pools, args)
	case "insert-data":
		if len(args) < 5 {
			return fmt.Errorf("недостаточно аргументов для команды insert-data")
		}
		data := TData{Key: args[3], Value: args[4]}
		insertCmd := &InsertCommand{InitialVersion: data}
		cr.AddHandler(insertCmd)
		fmt.Println("Команда вставки добавлена")
	case "update-data":
		if len(args) < 4 {
			return fmt.Errorf("недостаточно аргументов для команды update-data")
		}
		updateCmd := &UpdateCommand{UpdateExpression: args[3]}
		cr.AddHandler(updateCmd)
		fmt.Println("Команда обновления добавлена")
	case "delete-data":
		if len(args) < 3 {
			return fmt.Errorf("недостаточно аргументов для команды delete-data")
		}
		deleteCmd := &DisposeCommand{}
		cr.AddHandler(deleteCmd)
		fmt.Println("Команда удаления добавлена")
	case "get-data":
		if len(args) < 2 {
			return fmt.Errorf("недостаточно аргументов для команды get-data")
		}
		data, err := pools.GetPools(args[1])
		if err != nil {
			return err
		}
		fmt.Println("Полученные данные:", data)
	case "execute":
		var dataExists bool
		var data TData
		data.Timestamp = time.Now()
		cr.FirstHandler.Handle(&dataExists, &data, time.Now().Unix())
		HandleCommand(&data)
		fmt.Println("Команды выполнены, текущее состояние:", data.Timestamp.Format("2006-01-02 15:04:05"))
	case "save-state":
		if len(args) < 2 {
			return fmt.Errorf("недостаточно аргументов для команды save-state")
		}
		err := pools.SaveToFile(args[1])
		if err != nil {
			return err
		}
		fmt.Println("Состояние системы успешно сохранено в файл:", args[1])
	case "exit":
		return nil
	default:
		return fmt.Errorf("неизвестная команда")
	}
	return nil
}

func HandleCommand(data *TData) {
	data.Timestamp = time.Now()
}

var users = map[string]string{
	"admin": "password1234",
}

func authenticate(username, password string) bool {
	if pass, ok := users[username]; ok {
		return pass == password
	}
	return false
}

func main() {
	pools := InitPools()
	cr := &ChainOfResponsibility{}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		file, err := os.Open("login.html")
		if err != nil {
			http.Error(w, "Could not read HTML file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		w.Header().Set("Content-Type", "text/html")
		if _, err := io.Copy(w, file); err != nil {
			http.Error(w, "Failed to send HTML file", http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/authenticate", func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("username")
		password := r.URL.Query().Get("password")
		if authenticate(username, password) {
			http.SetCookie(w, &http.Cookie{
				Name:    "authenticated",
				Value:   "true",
				Path:    "/",
				Expires: time.Now().Add(24 * time.Hour),
			})
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"success": true}`)
		} else {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"success": false}`)
		}
	})

	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cookie, err := r.Cookie("authenticated"); err != nil || cookie.Value != "true" {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	http.Handle("/commands", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, err := os.Open("commands.html")
		if err != nil {
			http.Error(w, "Could not read HTML file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		w.Header().Set("Content-Type", "text/html")
		if _, err := io.Copy(w, file); err != nil {
			http.Error(w, "Failed to send HTML file", http.StatusInternalServerError)
		}
	})))

	http.HandleFunc("/run-command", func(w http.ResponseWriter, r *http.Request) {
		command := r.URL.Query().Get("command")
		if command == "" {
			http.Error(w, `{"error": "Missing command parameter"}`, http.StatusBadRequest)
			return
		}
		if err := RunCommand(pools, command, cr); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "Error executing command: %s"}`, err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "Command executed successfully"}`)
	})

	http.HandleFunc("/get-info", func(w http.ResponseWriter, r *http.Request) {
		type Info struct {
			Pools map[string][]string `json:"pools"`
		}
		info := Info{Pools: make(map[string][]string)}
		for poolName, pool := range pools.Pools {
			for schemaName := range pool.schema {
				info.Pools[poolName] = append(info.Pools[poolName], schemaName)
			}
		}
		data, err := json.Marshal(info)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "Error getting info: %s"}`, err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			file, err := os.Open("registration.html")
			if err != nil {
				http.Error(w, "Ошибка при чтении HTML файла", http.StatusInternalServerError)
				return
			}
			defer file.Close()

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			if _, err := io.Copy(w, file); err != nil {
				http.Error(w, "Ошибка при отправке HTML файла", http.StatusInternalServerError)
				return
			}
			return
		}

		if r.Method == http.MethodPost {
			username := r.FormValue("username")
			password := r.FormValue("password")
			users[username] = password
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
