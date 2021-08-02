package main

import (
	"fmt"
	"github.com/djumanoff/amqp"
	"github.com/joho/godotenv"
	setdata_common "github.com/kirigaikabuto/setdata-common"
	setdata_questionnaire_store "github.com/kirigaikabuto/setdata-questionnaire-store"
	"github.com/urfave/cli"
	"os"
	"strconv"
)

var (
	configPath           = ".env"
	version              = "0.0.0"
	amqpHost             = ""
	amqpPort             = 0
	postgresUser         = "oaxlkqvpikdard"
	postgresPassword     = "79a272cdf4249041aa90183895ff92d9b2d1e6bd69cd5165552f98c6f0e634bd"
	postgresDatabaseName = "dd4k5rjp3rmvg1"
	postgresHost         = "ec2-44-194-54-123.compute-1.amazonaws.com"
	postgresPort         = 5432
	postgresParams       = ""
	flags                = []cli.Flag{
		&cli.StringFlag{
			Name:        "config, c",
			Usage:       "path to .env config file",
			Destination: &configPath,
		},
	}
)

func parseEnvFile() {
	// Parse config file (.env) if path to it specified and populate env vars
	if configPath != "" {
		godotenv.Overload(configPath)
	}
	amqpHost = os.Getenv("RABBIT_HOST")
	amqpPortStr := os.Getenv("RABBIT_PORT")
	amqpPort, _ = strconv.Atoi(amqpPortStr)
	if amqpPort == 0 {
		amqpPort = 5672
	}
	if amqpHost == "" {
		amqpHost = "localhost"
	}
	//postgresUser = os.Getenv("POSTGRES_USER")
	//postgresPassword = os.Getenv("POSTGRES_PASSWORD")
	//postgresDatabaseName = os.Getenv("POSTGRES_DATABASE")
	//postgresParams = os.Getenv("POSTGRES_PARAMS")
	//portStr := os.Getenv("POSTGRES_PORT")
	//postgresPort, _ = strconv.Atoi(portStr)
	//postgresHost = os.Getenv("POSTGRES_HOST")
	//if postgresHost == "" {
	//	postgresHost = "localhost"
	//}
	//if postgresPort == 0 {
	//	postgresPort = 5432
	//}
	//if postgresUser == "" {
	//	postgresUser = "setdatauser"
	//}
	//if postgresPassword == "" {
	//	postgresPassword = "123456789"
	//}
	//if postgresDatabaseName == "" {
	//	postgresDatabaseName = "setdata"
	//}
	//if postgresParams == "" {
	//	//postgresParams = "sslmode=disable"
	//}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	app := cli.NewApp()
	app.Name = "Set data questions api"
	app.Description = ""
	app.Usage = "set data run"
	app.UsageText = "set data run"
	app.Version = version
	app.Flags = flags
	app.Action = run

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}

func run(c *cli.Context) error {
	parseEnvFile()
	rabbitConfig := amqp.Config{
		AMQPUrl:  "amqps://futohrkk:Qq4imfTpgcDawG6bzuSnJALRg-a6xqZl@toad.rmq.cloudamqp.com/futohrkk",
		LogLevel: 5,
	}
	serverConfig := amqp.ServerConfig{
		ResponseX: "response",
		RequestX:  "request",
	}
	sess := amqp.NewSession(rabbitConfig)
	err := sess.Connect()
	if err != nil {
		return err
	}
	srv, err := sess.Server(serverConfig)
	if err != nil {
		return err
	}
	cfg := setdata_questionnaire_store.PostgresConfig{
		Host:             postgresHost,
		Port:             postgresPort,
		User:             postgresUser,
		Password:         postgresPassword,
		Database:         postgresDatabaseName,
		Params:           postgresParams,
		ConnectionString: "",
	}
	questionPostgreStore, err := setdata_questionnaire_store.NewQuestionsPostgresStore(cfg)
	if err != nil {
		return err
	}
	questionService := setdata_questionnaire_store.NewQuestionsService(questionPostgreStore)
	questionsAmqpEndpoints := setdata_questionnaire_store.NewQuestionsAmqpEndpoints(setdata_common.NewCommandHandler(questionService))
	srv.Endpoint("questions.create", questionsAmqpEndpoints.MakeCreateQuestionAmqpEndpoint())
	srv.Endpoint("questions.delete", questionsAmqpEndpoints.MakeDeleteQuestionAmqpEndpoint())
	srv.Endpoint("questions.update", questionsAmqpEndpoints.MakeUpdateQuestionAmqpEndpoint())
	srv.Endpoint("questions.list", questionsAmqpEndpoints.MakeListQuestionAmqpEndpoint())
	srv.Endpoint("questions.get", questionsAmqpEndpoints.MakeGetQuestionAmqpEndpoint())
	err = srv.Start()
	if err != nil {
		return err
	}
	return nil
}
