package main

import (
	"context"
	"flag"
	"github.com/ONSdigital/log.go/v2/log"
	"os"
	"testing"

	"github.com/ONSdigital/dis-migration-service/features/steps"
	componentTest "github.com/ONSdigital/dp-component-test"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var componentFlag = flag.Bool("component", false, "perform component tests")

const mongoVersion = "4.4.8"
const databaseName = "testing"
const replicaSetName = "rs0"

type ComponentTest struct {
	Mongo *componentTest.MongoFeature
}

func (f *ComponentTest) InitializeScenario(godogCtx *godog.ScenarioContext) {

	ctx := context.Background()
	//if f.MongoFeature == nil {
	//	mongoOpts := componentTest.MongoOptions{MongoVersion: mongoVersion, DatabaseName: databaseName}
	//	f.MongoFeature = componentTest.NewMongoFeature(mongoOpts)
	//	f.MongoFeature.Client.Database(databaseName).CreateCollection(context.Background(), "jobs")
	//	f.MongoFeature.Client.Database(databaseName).CreateCollection(context.Background(), "events")
	//	f.MongoFeature.Client.Database(databaseName).CreateCollection(context.Background(), "tasks")
	//}

	migrationComponent, err := steps.NewMigrationComponent(f.Mongo)
	if err != nil {
		log.Error(ctx, "error occurred while creating a new migrationComponent", err)
		os.Exit(1)
	}

	apiFeature := migrationComponent.InitAPIFeature()

	godogCtx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		//if f.MongoFeature == nil {
		//	f.MongoFeature = componentTest.NewMongoFeature(componentTest.MongoOptions{MongoVersion: mongoVersion, DatabaseName: databaseName})
		//}
		//migrationComponent.Reset()
		f.Mongo.Reset()
		apiFeature.Reset()

		return ctx, nil
	})

	godogCtx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		//if f.MongoFeature != nil {x
		//	if closeErr := f.MongoFeature.Close(); closeErr != nil {
		//		log.Error(context.Background(), "error occurred while closing the MongoFeature", closeErr)
		//		os.Exit(1)
		//	}
		//}
		//migrationComponent.Reset()
		f.Mongo.Reset()
		apiFeature.Reset()

		return ctx, nil
	})

	//f.MongoFeature.RegisterSteps(godogCtx)
	apiFeature.RegisterSteps(godogCtx)
	migrationComponent.RegisterSteps(godogCtx)
}

func (f *ComponentTest) InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctxBackground := context.Background()

	ctx.BeforeSuite(func() {
		mongoOptions := componentTest.MongoOptions{
			MongoVersion: mongoVersion,
			DatabaseName: databaseName,
		}
		f.Mongo = componentTest.NewMongoFeature(mongoOptions)
	})
	ctx.AfterSuite(func() {
		err := f.Mongo.Close()
		if err != nil {
			log.Error(ctxBackground, "error occurred while closing the MongoFeature", err)
			os.Exit(1)
		}
	})
}

func TestComponent(t *testing.T) {
	if *componentFlag {
		status := 0

		var opts = godog.Options{
			Output: colors.Colored(os.Stdout),
			Format: "pretty",
			Paths:  flag.Args(),
			Strict: true,
		}

		f := &ComponentTest{}

		status = godog.TestSuite{
			Name:                 "feature_tests",
			ScenarioInitializer:  f.InitializeScenario,
			TestSuiteInitializer: f.InitializeTestSuite,
			Options:              &opts,
		}.Run()

		if status > 0 {
			t.Fail()
		}
	} else {
		t.Skip("component flag required to run component tests")
	}
}
