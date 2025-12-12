package main

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/ONSdigital/log.go/v2/log"

	"github.com/ONSdigital/dis-migration-service/features/steps"
	componentTest "github.com/ONSdigital/dp-component-test"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var componentFlag = flag.Bool("component", false, "perform component tests")

const mongoVersion = "6.0.4"
const databaseName = "testing"

type ComponentTest struct {
	Mongo *componentTest.MongoFeature
	Auth  *componentTest.AuthorizationFeature
}

func (f *ComponentTest) InitializeScenario(godogCtx *godog.ScenarioContext) {
	ctx := context.Background()

	migrationComponent, err := steps.NewMigrationComponent(f.Mongo, f.Auth)
	if err != nil {
		log.Error(ctx, "error occurred while creating a new migrationComponent", err)
		os.Exit(1)
	}

	apiFeature := migrationComponent.InitAPIFeature()

	godogCtx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		f.Mongo.Reset()
		migrationComponent.SeedDatabase()
		apiFeature.Reset()
		f.Auth.Reset()

		return ctx, nil
	})

	godogCtx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		migrationComponent.Close()
		f.Mongo.Reset()
		apiFeature.Reset()
		f.Auth.Reset()

		return ctx, nil
	})

	apiFeature.RegisterSteps(godogCtx)
	f.Mongo.RegisterSteps(godogCtx)
	migrationComponent.RegisterSteps(godogCtx)
	f.Auth.RegisterSteps(godogCtx)
}

func (f *ComponentTest) InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctxBackground := context.Background()

	ctx.BeforeSuite(func() {
		mongoOptions := componentTest.MongoOptions{
			MongoVersion: mongoVersion,
			DatabaseName: databaseName,
		}
		f.Mongo = componentTest.NewMongoFeature(mongoOptions)
		f.Auth = componentTest.NewAuthorizationFeature()
	})
	ctx.AfterSuite(func() {
		err := f.Mongo.Close()
		if err != nil {
			log.Error(ctxBackground, "error occurred while closing the MongoFeature", err)
			os.Exit(1)
		}

		f.Auth.Close()
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
