package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// flags
var (
	operationFlag = flag.String("operation", "", "name of the operation")
	idflag        = flag.String("id", "", "id of the operation")
	itemFlag      = flag.String("item", "", "operational payload")
	fileNameFlag  = flag.String("fileName", "", "fileName of the operation")
)

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type Users []User

// errors
var (
	// Use when fileName is not specified
	fileNameError = errors.New("-fileName flag has to be specified")
	// Use when operation is not specified
	noOperationError = errors.New("-operation flag has to be specified")
	// Use when operation is wrong/not existing
	wrongOperationError = errors.New("-operation flag does not exist")
	// Use when item is not specified
	itemError = errors.New("-item flag has to be specified")
	// Use when id flag is not provided
	idFlagError = errors.New("-id flag has to be specified")
)

type Arguments map[string]string

func Perform(args Arguments, writer io.Writer) error {
	if args["fileName"] == "" {
		return fileNameError
	}

	if args["operation"] == "" {
		return noOperationError
	}
	fileName := args["fileName"]

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	operationErr := doOperation(args, file, writer)
	if operationErr != nil {
		return operationErr
	}

	return nil

}

func doOperation(args Arguments, file *os.File, writer io.Writer) error {
	switch args["operation"] {
	case "add":
		err := add(args["item"], file)
		return err
	case "list":
		err := list(file, writer)
		return err
	case "findById":
		err := findById(args["id"], file, writer)
		return err
	case "remove":
		err := remove(args["id"], file)
		return err
	default:
		return wrongOperationError
	}
}

func add(item string, file *os.File) error {
	var data Users
	var userItem User

	if item == "" {
		return itemError
	}

	content, readEr := ioutil.ReadAll(file)
	if readEr != nil {
		return fmt.Errorf("error while reading: %s", readEr)
	}

	if len(content) > 0 {
		e := json.Unmarshal(content, &data)
		if e != nil {
			return e
		}
	}

	itemParseErr := json.Unmarshal([]byte(item), &userItem)
	if itemParseErr != nil {
		return fmt.Errorf("error during item parsing: %s", itemParseErr)
	}

	data = append(data, userItem)

	res, marshErr := json.Marshal(data)
	if marshErr != nil {
		return marshErr
	}

	file.Truncate(0)
	_, err := file.Write(res)

	if err != nil {
		return err
	}

	return nil
}
func list(file *os.File, writer io.Writer) error {

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	_, wrErr := writer.Write(content)
	if wrErr != nil {
		return wrErr
	}

	return nil
}
func findById(id string, file *os.File, writer io.Writer) error {

	if id == "" {
		return idFlagError
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	var data Users
	err = json.Unmarshal(content, &data)
	if err != nil {
		return err
	}

	var foundUser User

	for _, i := range data {
		if i.Id == id {
			foundUser = i
		}
	}
	res, marshErr := json.Marshal(foundUser)
	if marshErr != nil {
		return marshErr
	}

	_, writeErr := writer.Write(res)
	if writeErr != nil {
		return writeErr
	}

	return nil
}
func remove(id string, file *os.File) error {
	if id == "" {
		return idFlagError
	}

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	var data Users
	var newData Users
	err = json.Unmarshal(content, &data)
	if err != nil {
		return err
	}

	for _, i := range data {
		if i.Id != id {
			newData = append(newData, i)
		}
	}

	if len(data) == len(newData) {
		return fmt.Errorf("Item with id %s not found", id)
	}

	res, marshErr := json.Marshal(newData)
	if marshErr != nil {
		return marshErr
	}

	file.Truncate(0)
	_, writeErr := file.Write(res)

	if writeErr != nil {
		return err
	}

	return nil
}

func parseArgs() Arguments {
	flag.Parse()
	args := Arguments{
		"id":        *idflag,
		"operation": *operationFlag,
		"item":      *itemFlag,
		"fileName":  *fileNameFlag,
	}
	return args
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}

}
