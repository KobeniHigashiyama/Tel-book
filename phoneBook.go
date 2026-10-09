package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Entry struct {
	Name       string
	Surname    string
	Tel        string
	LastAccess string
}

type PhoneBook []Entry

// Len, Less и Swap реализуют sort.Interface для PhoneBook.
func (a PhoneBook) Len() int {
	return len(a)
}

func (a PhoneBook) Less(i, j int) bool {
	if a[i].Surname == a[j].Surname {
		return a[i].Name < a[j].Name
	}
	return a[i].Surname < a[j].Surname
}

func (a PhoneBook) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// CSVFILE is relative to the directory from which the program is run.
var CSVFILE = "csv.data"

var data = PhoneBook{}
var index map[string]int

func setCSVFile() error {
	filepath := os.Getenv("PHONEBOOk")
	if filepath != "" {
		CSVFILE = filepath
	}
	_, err := os.Stat(CSVFILE)
	if err != nil {
		fmt.Println("Creating", CSVFILE)
		f, err := os.Create(CSVFILE)
		if err != nil {
			f.Close()
			return err
		}
		f.Close()
	}
	fileInfo, err := os.Stat(CSVFILE)
	mode := fileInfo.Mode()
	if !mode.IsRegular() {
		return fmt.Errorf("%s not a regular file", CSVFILE)
	}
	return nil

}

func readCSVFile(filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = 4
	lines, err := reader.ReadAll()
	if err != nil {
		return err
	}

	entries := make([]Entry, 0, len(lines))
	for _, line := range lines {
		entries = append(entries, Entry{
			Name:       line[0],
			Surname:    line[1],
			Tel:        line[2],
			LastAccess: line[3],
		})
	}
	data = entries
	return nil
}

func saveCSVFile(filepath string) error {
	csvfile, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer csvfile.Close()

	csvwriter := csv.NewWriter(csvfile)
	for _, row := range data {
		temp := []string{row.Name, row.Surname, row.Tel, row.LastAccess}
		if err := csvwriter.Write(temp); err != nil {
			return err
		}
	}
	csvwriter.Flush()
	if err := csvwriter.Error(); err != nil {
		return err
	}
	return csvfile.Close()
}

func createIndex() error {
	newIndex := make(map[string]int)
	for i, entry := range data {
		if _, ok := newIndex[entry.Tel]; ok {
			return fmt.Errorf("%s already exists", entry.Tel)
		}
		newIndex[entry.Tel] = i
	}
	index = newIndex
	return nil
}

func initS(N, S, T string) *Entry {
	if T == "" || S == "" {
		return nil
	}
	lastAccess := strconv.FormatInt(time.Now().Unix(), 10)
	return &Entry{Name: N, Surname: S, Tel: T, LastAccess: lastAccess}
}

func insert(pS *Entry) error {
	if pS == nil {
		return fmt.Errorf("contact must not be nil")
	}
	if _, ok := index[pS.Tel]; ok {
		return fmt.Errorf("%s already exists", pS.Tel)
	}
	data = append(data, *pS)
	if err := createIndex(); err != nil {
		return err
	}
	return saveCSVFile(CSVFILE)
}

func deleteEntry(key string) error {
	i, ok := index[key]
	if !ok {
		return fmt.Errorf("%s cannot be found", key)
	}
	data = append(data[:i], data[i+1:]...)
	// Removing an entry shifts the positions of the following contacts.
	if err := createIndex(); err != nil {
		return err
	}
	return saveCSVFile(CSVFILE)
}

func search(key string) (*Entry, error) {
	i, ok := index[key]
	if !ok {
		return nil, nil
	}
	data[i].LastAccess = strconv.FormatInt(time.Now().Unix(), 10)
	if err := saveCSVFile(CSVFILE); err != nil {
		return nil, err
	}
	return &data[i], nil
}

func list() {
	sort.Sort(data)
	// После сортировки позиции контактов в срезе изменились.
	if err := createIndex(); err != nil {
		fmt.Println("Cannot create index:", err)
		return
	}
	for _, entry := range data {
		fmt.Println(entry)
	}
}

func matchTel(s string) bool {
	re := regexp.MustCompile(`^\d+$`)
	return re.MatchString(s)
}

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Usage: insert|delete|search|list <arguments>")
		return
	}

	fileInfo, err := os.Stat(CSVFILE)
	if os.IsNotExist(err) {
		fmt.Println("Creating", CSVFILE)
		f, err := os.Create(CSVFILE)
		if err != nil {
			fmt.Println(err)
			return
		}
		if err := f.Close(); err != nil {
			fmt.Println(err)
			return
		}
	} else if err != nil {
		fmt.Println(err)
		return
	} else if !fileInfo.Mode().IsRegular() {
		fmt.Println(CSVFILE, "not a regular file!")
		return
	}

	if err := readCSVFile(CSVFILE); err != nil {
		fmt.Println(err)
		return
	}
	if err := createIndex(); err != nil {
		fmt.Println("Cannot create index:", err)
		return
	}

	switch arguments[1] {
	case "insert":
		if len(arguments) != 5 {
			fmt.Println("Usage: insert Name Surname Telephone")
			return
		}
		t := strings.ReplaceAll(arguments[4], "-", "")
		if !matchTel(t) {
			fmt.Println("Not a valid telephone number:", t)
			return
		}
		temp := initS(arguments[2], arguments[3], t)
		if temp == nil {
			fmt.Println("Surname and telephone must not be empty")
			return
		}
		if err := insert(temp); err != nil {
			fmt.Println(err)
		}
	case "delete":
		if len(arguments) != 3 {
			fmt.Println("Usage: delete Number")
			return
		}
		t := strings.ReplaceAll(arguments[2], "-", "")
		if !matchTel(t) {
			fmt.Println("Not a valid telephone number:", t)
			return
		}
		if err := deleteEntry(t); err != nil {
			fmt.Println(err)
		}
	case "search":
		if len(arguments) != 3 {
			fmt.Println("Usage: search Number")
			return
		}
		t := strings.ReplaceAll(arguments[2], "-", "")
		if !matchTel(t) {
			fmt.Println("Not a valid telephone number:", t)
			return
		}
		temp, err := search(t)
		if err != nil {
			fmt.Println(err)
			return
		}
		if temp == nil {
			fmt.Println("Number not found:", t)
			return
		}
		fmt.Println(*temp)
	case "list":
		list()
	default:
		fmt.Println("Not a valid option")
	}
}
