package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type Response struct {
	Result string          `json:"result"`
	Data   json.RawMessage `json:"data,omitempty"`
	Error  string          `json:"error,omitempty"`
}

type Note struct {
	Name     string `json:"name,omitempty" sql.field:"name"`
	LastName string `json:"last_name,omitempty" sql.field:"last_name"`
	Text     string `json:"text,omitempty" sql.field:"text"`
}

func connentToServer() {
	for {
		var command string
		fmt.Print("Выберите, что хотите сделать [1 - Добавить записку, 2 - Обновить записку, 3 - Найти записку, 4 - Удалить записку]: ")
		fmt.Scanln(&command)
		if command == "1" {
			// confirm := "yes"
			// fmt.Print("Эта функция сохраняет записку в хранилище. Хотите продолжить? [yes]: ")
			// fmt.Scanln(&confirm)
			// if confirm != "yes" {
			// 	continue
			// }
			note := &Note{}
			fmt.Print("Введите имя: ")
			fmt.Scanln(&note.Name)
			fmt.Print("Введите фамилию: ")
			fmt.Scanln(&note.LastName)
			fmt.Println("Введите текст записки (чтобы завершить ввод, введите end в отдельной строке): ")
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				line := scanner.Text()
				// "end" - завершающая строка
				if line == "end" {
					break
				}
				// Добавляем считанную строку к общему тексту
				note.Text += line + "\n"
			}
			// Проверяем ошибки, возможные при сканировании
			if err := scanner.Err(); err != nil {
				log.Println("Error while scaning:", err)
				return
			}
			saveNote(note)
			continue
		}
		if command == "2" {
			// confirm := "yes"
			// fmt.Print("Эта функция обновяет записку. Хотите продолжить? [yes]: ")
			// fmt.Scanln(&confirm)
			// if confirm != "yes" {
			// 	continue
			// }
			note := &Note{}
			var id int64
			fmt.Print("Введите ID записки, которую хотите обновить: ")
			fmt.Scanln(&id)
			fmt.Println("Заполните те данные, которые хотите обновить")
			fmt.Print("Имя: ")
			fmt.Scanln(&note.Name)
			fmt.Print("Фамилия: ")
			fmt.Scanln(&note.LastName)
			fmt.Println("Текст записки (введите end, если не хотите обновлять текст заметки): ")
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				line := scanner.Text()
				if line == "end" {
					break
				}
				// Добавляем считанную строку к общему тексту
				note.Text += line + "\n"
			}
			// Проверяем ошибки, возможные при сканировании
			if err := scanner.Err(); err != nil {
				log.Println("Error while scaning:", err)
				return
			}
			updateNote(note, id)
			continue
		}
		if command == "3" {
			// confirm := "yes"
			// fmt.Print("Эта функция поиска записки. Хотите продолжить? [yes]: ")
			// fmt.Scanln(&confirm)
			// if confirm != "yes" {
			// 	continue
			// }
			var id []byte
			fmt.Print("Введите ID записки, которую хотите прочитать: ")
			fmt.Scanln(&id)
			readNote(id)
			continue
		}
		if command == "4" {
			// confirm := "yes"
			// fmt.Print("Эта функция удаления записки. Хотите продолжить? [yes]: ")
			// fmt.Scanln(&confirm)
			// if confirm != "yes" {
			// 	continue
			// }
			var id []byte
			fmt.Print("Введите ID записки, которую хотите удалить: ")
			fmt.Scanln(&id)
			deleteNote(id)
			continue
		}
		fmt.Println("Пожалуйста, введите корректную команду")
	}
}

func saveNote(note *Note) {
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintln(file, "Error oppening file:", err)
		log.Println("Ошибка при открытии файла logs.txt")
		return
	}
	jsonData, err := json.Marshal(note)
	if err != nil {
		fmt.Fprintln(file, "Error encoding JSON:", err)
		log.Println("Ошибка при сериализации заметки (подробности см. в файле logs.txt)")
		return
	}

	resp, err := http.Post("http://localhost:4040/save", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Fprintln(file, "Error sending POST request:", err)
		log.Println("Ошибка при отправке POST запроса (подробности см. в файле logs.txt)")
		return
	}
	defer resp.Body.Close()

	nodeData := make([]byte, resp.ContentLength)
	_, err = io.ReadFull(resp.Body, nodeData)
	if err != nil {
		fmt.Fprintln(file, "io.ReadFull(resp.Body, recordsData):", err)
		log.Println("Ошибка при чтении ответа от сервера (подробности см. в файле logs.txt)")
		return
	}
	var response Response
	err = json.Unmarshal(nodeData, &response)
	if err != nil {
		fmt.Fprintln(file, "json.Unmarshal(nodeData, &response):", err)
		log.Println("Ошибка при десериализации ответа (подробности см. в файле logs.txt)")
		return
	}
	if response.Result == "Error" {
		fmt.Fprintln(file, response.Error)
		log.Println("Не удалось создать заметку (подробности см. в файле logs.txt)")
		return
	}
	fmt.Println("Заметка успешно создана с ID: " + string(response.Data))
}

func updateNote(note *Note, id int64) {
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintln(file, "Error oppening file:", err)
		log.Println("Ошибка при открытии файла logs.txt")
		return
	}
	var requestData struct {
		Index int64 `json:"index"`
		Data  Note  `json:"data"`
	}

	requestData.Index = id
	requestData.Data = *note

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		fmt.Fprintln(file, "Error encoding JSON:", err)
		log.Println("Ошибка при сериализации заметки (подробности см. в файле logs.txt)")
		return
	}

	resp, err := http.Post("http://localhost:4040/update", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Fprintln(file, "Error sending POST request:", err)
		log.Println("Ошибка при отправке POST запроса (подробности см. в файле logs.txt)")
		return
	}
	defer resp.Body.Close()
	noteData := make([]byte, resp.ContentLength)
	_, err = io.ReadFull(resp.Body, noteData)
	if err != nil {
		fmt.Fprintln(file, "io.ReadFull(resp.Body, noteData):", err)
		log.Println("Ошибка при чтении ответа от сервера (подробности см. в файле logs.txt)")
		return
	}
	var response Response
	err = json.Unmarshal(noteData, &response)
	if err != nil {
		fmt.Fprintln(file, "json.Unmarshal(noteData, &resp):", err)
		log.Println("Ошибка при десериализации ответа (подробности см. в файле logs.txt)")
		return
	}

	if response.Result == "Error" {
		fmt.Fprintln(file, response.Error)
		log.Println("Не удалось обновить заметку (подробности см. в файле logs.txt)")
		return
	}
	fmt.Println("Заметка успешно обновлена. Новый ID: " + string(response.Data))
}

func deleteNote(id []byte) {
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintln(file, "Error oppening file:", err)
		log.Println("Ошибка при открытии файла logs.txt")
		return
	}
	resp, err := http.Post("http://localhost:4040/delete", "text/plain", bytes.NewBuffer(id))
	if err != nil {
		fmt.Fprintln(file, "Error sending POST request:", err)
		log.Println("Ошибка при отправке POST запроса (подробности см. в файле logs.txt)")
		return
	}
	defer resp.Body.Close()

	nodeData := make([]byte, resp.ContentLength)
	_, err = io.ReadFull(resp.Body, nodeData)
	if err != nil {
		fmt.Fprintln(file, "io.ReadFull(resp.Body, nodeData):", err)
		log.Println("Ошибка при чтении ответа от сервера (подробности см. в файле logs.txt)")
		return
	}

	var response Response
	err = json.Unmarshal(nodeData, &response)
	if err != nil {
		fmt.Fprintln(file, "json.Unmarshal(nodeData, &response):", err)
		log.Println("Ошибка при десериализации ответа (подробности см. в файле logs.txt)")
		return
	}
	if response.Result == "Error" {
		log.Println(response.Error)
		log.Println("Не удалось удалить заметку (подробности см. в файле logs.txt)")
		return
	}
	fmt.Println("Записка успешно удалена")
}

func readNote(id []byte) {
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintln(file, "Error oppening file:", err)
		log.Println("Ошибка при открытии файла logs.txt")
		return
	}
	resp, err := http.Post("http://localhost:4040/read", "text/plain", bytes.NewBuffer(id))
	if err != nil {
		fmt.Fprintln(file, "Error sending POST request:", err)
		log.Println("Ошибка при отправке POST запроса (подробности см. в файле logs.txt)")
		return
	}
	defer resp.Body.Close()

	noteData := make([]byte, resp.ContentLength)
	_, err = io.ReadFull(resp.Body, noteData)
	if err != nil {
		fmt.Fprintln(file, "io.ReadFull(resp.Body, noteData):", err)
		log.Println("Ошибка при чтении ответа от сервера (подробности см. в файле logs.txt)")
		return
	}
	var response Response
	err = json.Unmarshal(noteData, &response)
	if err != nil {
		fmt.Fprintln(file, "json.Unmarshal(noteData, &resp):", err)
		log.Println("Ошибка при десериализации ответа (подробности см. в файле logs.txt)")
		return
	}

	var note Note
	err = json.Unmarshal(response.Data, &note)
	if err != nil {
		fmt.Fprintln(file, "json.Unmarshal(response.Data, &note):", err)
		log.Println("Ошибка при десериализации тела заметки (подробности см. в файле logs.txt)")
		return
	}
	if response.Result == "Error" {
		fmt.Fprintln(file, response.Error)
		log.Println("Не удалось прочитать заметку (подробности см. в файле logs.txt)")
		return
	}
	fmt.Println("Имя:" + note.Name)
	fmt.Println("Фамилия:" + note.LastName)
	fmt.Println("Текст заметки:")
	fmt.Println(note.Text)
}

func main() {
	connentToServer()
}
