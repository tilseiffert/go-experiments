package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/oklog/ulid/v2"
	"github.com/spf13/viper"
	"github.com/tilseiffert/go-tools-config/config"
)

const (
	ProjectID = "231001"
	TableName = "go-experiments-dynamodb"
)

// Todo represents a todo item within a specific project.
type Todo struct {
	// ProjectID is the identifier for the project to which this todo belongs.
	ProjectID string `json:"project_id" dynamodbav:"ProjectID"`

	// TodoID is the unique identifier for this todo item.
	TodoID string `json:"todo_id" dynamodbav:"TodoID"`

	// Task describes what needs to be done.
	Task string `json:"task" dynamodbav:"Task"`

	// Complete indicates whether the task has been completed.
	Complete bool `json:"complete" dynamodbav:"Complete"`
}

func (t Todo) String() string {
	return fmt.Sprintf("%s/%s: %s (%v)", t.ProjectID, t.TodoID, t.Task, t.Complete)
}

// initConfig initializes the configuration
func initConfig() (map[string]*config.Option, error) {

	// prepare config
	conf := config.New()

	// init a map of options
	options := make(map[string]*config.Option)

	options["TABLENAME"] = conf.NewOption("TABLENAME", TableName, true, "")

	// initialize
	err := config.Init(conf)

	if err != nil {
		return nil, err
	}

	viper.SetTypeByDefaultValue(true)

	return options, nil
}

// addTodo adds a todo item to the database
func addTodo(svc *dynamodb.DynamoDB, table string, todo Todo) (*dynamodb.PutItemOutput, error) {
	input := &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item: map[string]*dynamodb.AttributeValue{
			"ProjectID": {S: aws.String(todo.ProjectID)},
			"TodoID":    {S: aws.String(todo.TodoID)},
			"Task":      {S: aws.String(todo.Task)},
			"Complete":  {BOOL: aws.Bool(todo.Complete)},
		},
	}

	res, err := svc.PutItem(input)
	if err != nil {
		return res, fmt.Errorf("failed to put item: %w", err)
	}

	return res, nil
}

// getAllTodos returns all todos for a given project
func getAllTodos(svc *dynamodb.DynamoDB, table string, projectID string) ([]Todo, error) {

	input := &dynamodb.QueryInput{
		TableName: aws.String(table),
		KeyConditions: map[string]*dynamodb.Condition{
			"ProjectID": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(projectID),
					},
				},
			},
		},
	}

	result, err := svc.Query(input)

	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}

	// Konvertiere result.Items in []Todo
	var todos []Todo
	for _, i := range result.Items {
		todo := Todo{}

		fmt.Println(i)

		err = dynamodbattribute.UnmarshalMap(i, &todo)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal items: %w", err)
		}

		todos = append(todos, todo)
	}

	return todos, nil
}

// Entry point of the program
func main() {

	fmt.Println("Hello 👋")
	defer fmt.Println("Bye 👋")

	options, err := initConfig()
	if err != nil {
		panic(err)
	}

	table := fmt.Sprint(options["TABLENAME"].Get())

	sess := session.Must(session.NewSession(&aws.Config{}))
	svc := dynamodb.New(sess)

	res, err := addTodo(svc, table, Todo{
		ProjectID: ProjectID,
		TodoID:    ulid.Make().String(),
		Task:      "Task " + time.Now().Format(time.RFC3339),
		Complete:  false,
	})

	if err != nil {
		panic(err)
	}

	fmt.Println("PutItem:", res)

	todos, err := getAllTodos(svc, table, ProjectID)

	if err != nil {
		panic(err)
	}

	fmt.Println("Todos:")
	for _, todo := range todos {
		fmt.Println(todo)
	}
}
