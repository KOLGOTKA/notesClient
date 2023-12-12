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
	// ID       int64  `json:"-" sql.field:"id"`
	Name     string `json:"name,omitempty" sql.field:"name"`
	LastName string `json:"last_name,omitempty" sql.field:"last_name"`
	Text     string `json:"text,omitempty" sql.field:"text"`
}

func connentToServer() {
	for {
		var command int
		fmt.Print("Выберите, что хотите сделать [1 - Добавить записку, 2 - Обновить записку, 3 - Найти записку, 4 - Удалить записку]: ")
		fmt.Scanln(&command)
		if command == 1 {
			confirm := "yes"
			fmt.Print("Эта функция сохраняет записку в хранилище. Хотите продолжить? [yes]: ")
			fmt.Scanln(&confirm)
			if confirm != "yes" {
				continue
			}
			note := &Note{}
			fmt.Print("Введите имя: ")
			fmt.Scanln(&note.Name)
			fmt.Print("Введите фамилию: ")
			fmt.Scanln(&note.LastName)
			fmt.Print("Введите текст записки (чтобы завершить ввод, введите end в отдельной строке): ")
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
		if command == 2 {
			confirm := "yes"
			fmt.Print("Эта функция обновяет записку. Хотите продолжить? [yes]: ")
			fmt.Scanln(&confirm)
			if confirm != "yes" {
				continue
			}
			note := &Note{}
			var id int64
			fmt.Print("Введите ID записки, которую хотите обновить: ")
			fmt.Scanln(&id)
			fmt.Println("Заполните те данные, которые хотите обновить")
			fmt.Print("Имя: ")
			fmt.Scanln(&note.Name)
			fmt.Print("Фамилия: ")
			fmt.Scanln(&note.LastName)
			fmt.Print("Текст записки (введите end, если не хотите обновлять текст заметки): ")
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
		if command == 3 {
			confirm := "yes"
			fmt.Print("Эта функция поиска записки. Хотите продолжить? [yes]: ")
			fmt.Scanln(&confirm)
			if confirm != "yes" {
				continue
			}
			var id []byte
			fmt.Print("Введите ID записки, которую хотите прочитать: ")
			fmt.Scanln(&id)
			readNote(id)
			continue
		}
		if command == 4 {
			confirm := "yes"
			fmt.Print("Эта функция удаления записки. Хотите продолжить? [yes]: ")
			fmt.Scanln(&confirm)
			if confirm != "yes" {
				continue
			}
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
	jsonData, err := json.Marshal(note)
	if err != nil {
		log.Println("Error encoding JSON:", err)
		return
	}

	resp, err := http.Post("http://localhost:4040/save", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error sending POST request:", err)
		return
	}
	defer resp.Body.Close()

	nodeData := make([]byte, resp.ContentLength)
	_, err = io.ReadFull(resp.Body, nodeData)
	if err != nil {
		log.Println("io.ReadFull(resp.Body, recordsData):", err)
		return
	}

	fmt.Println("Response Status:", resp.Status)
	var response Response
	err = json.Unmarshal(nodeData, &response)
	if err != nil {
		log.Println("json.Unmarshal(nodeData, &response):", err)
		return
	}
	if response.Result == "Error" {
		log.Println(response.Error)
		return
	}
	fmt.Println("Note successfully created with id: " + string(response.Data))
}

func updateNote(note *Note, id int64) {
	var requestData struct {
		Index int64 `json:"index"`
		Data  Note  `json:"data"`
	}

	requestData.Index = id
	requestData.Data = *note

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		log.Println("Error encoding JSON:", err)
		return
	}

	resp, err := http.Post("http://localhost:4040/update", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error sending POST request:", err)
		return
	}
	defer resp.Body.Close()
	noteData := make([]byte, resp.ContentLength)
	_, err = io.ReadFull(resp.Body, noteData)
	if err != nil {
		log.Println("io.ReadFull(resp.Body, noteData):", err)
		return
	}
	var response Response
	err = json.Unmarshal(noteData, &response)
	if err != nil {
		log.Println("json.Unmarshal(noteData, &resp):", err)
		return
	}
	
	if response.Result == "Error" {
		log.Println(response.Error)
		return
	}
	fmt.Println("Record successfully updated. New id is: " + string(response.Data))
}

func deleteNote(id []byte) {
	resp, err := http.Post("http://localhost:4040/delete", "application/text", bytes.NewBuffer(id))
	if err != nil {
		log.Println("Error sending POST request:", err)
		return
	}
	defer resp.Body.Close()

	nodeData := make([]byte, resp.ContentLength)
	_, err = io.ReadFull(resp.Body, nodeData)
	if err != nil {
		log.Println("io.ReadFull(resp.Body, nodeData):", err)
		return
	}

	fmt.Println("Response Status:", resp.Status)
	var response Response
	err = json.Unmarshal(nodeData, &response)
	if err != nil {
		log.Println("json.Unmarshal(nodeData, &response):", err)
		return
	}
	if response.Result == "Error" {
		log.Println(response.Error)
		return
	}
	fmt.Println("Record successfully deleted")
}

func readNote(id []byte) {
	resp, err := http.Post("http://localhost:4040/read", "application/json", bytes.NewBuffer(id))
	if err != nil {
		log.Println("Error sending POST request:", err)
		return
	}
	defer resp.Body.Close()

	noteData := make([]byte, resp.ContentLength)
	_, err = io.ReadFull(resp.Body, noteData)
	if err != nil {
		log.Println("io.ReadFull(resp.Body, noteData):", err)
		return
	}
	var response Response
	err = json.Unmarshal(noteData, &response)
	if err != nil {
		log.Println("json.Unmarshal(noteData, &resp):", err)
		return
	}

	var note Note
	err = json.Unmarshal(response.Data, &note)
	if err != nil {
		log.Println("json.Unmarshal(response.Data, &note):", err)
		return
	}
	if response.Result == "Error" {
		log.Println(response.Error)
		return
	}
	fmt.Println("\tResult: ")
	// fmt.Println("-->")
	fmt.Println("Name:" + note.Name)
	fmt.Println("Lastname:" + note.LastName)
	fmt.Println("Text:" + note.Text)
}

func main() {
	connentToServer()
	// var response Response
	// bs := []byte{85, 110, 101, 120, 112, 101, 99, 116, 101, 100, 32, 116, 121, 112, 101, 10, 123, 34, 114, 101, 115, 117, 108, 116, 34, 58, 34, 69, 114, 114, 111, 114, 34, 44, 34, 100, 97, 116, 97, 34, 58, 123, 125, 44, 34, 101, 114, 114, 111, 114, 34, 58, 34, 85, 110, 101, 120, 112, 101, 99, 116, 101, 100, 32, 116, 121, 112, 101, 32, 105, 110, 32, 100, 97, 116, 97, 46, 40, 42, 100, 116, 111, 46, 78, 111, 116, 101, 41, 58, 32, 112, 97, 99, 107, 97, 103, 101, 32, 115, 116, 100, 104, 116, 116, 112, 58, 32, 102, 117, 110, 99, 32, 40, 104, 115, 32, 42, 67, 111, 110, 116, 114, 111, 108, 108, 101, 114, 41, 32, 78, 111, 116, 101, 85, 112, 100, 97, 116, 101, 72, 97, 110, 100, 108, 101, 114, 40, 119, 32, 104, 116, 116, 112, 46, 82, 101, 115, 112, 111, 110, 115, 101, 87, 114, 105, 116, 101, 114, 44, 32, 114, 101, 113, 32, 42, 104, 116, 116, 112, 46, 82, 101, 113, 117, 101, 115, 116, 41, 34, 125}
	// sbs := string(bs)
	// fmt.Println(sbs)
	// err := json.Unmarshal([]byte(sbs), &response)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println("success")
	// fmt.Println(response)
}
