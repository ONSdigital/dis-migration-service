package main

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/ONSdigital/dis-migration-service/features/steps"
	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var componentFlag = flag.Bool("component", false, "perform component tests")

const mongoVersion = "4.4.8"
const databaseName = "testing"
const replicaSetName = "rs0"

type ComponentTest struct {
	MongoFeature *componenttest.MongoFeature
}

func (f *ComponentTest) InitializeScenario(ctx *godog.ScenarioContext) {
	mongoOpts := componenttest.MongoOptions{MongoVersion: mongoVersion, DatabaseName: databaseName, ReplicaSetName: replicaSetName}
	f.MongoFeature = componenttest.NewMongoFeature(mongoOpts)

	component, err := steps.NewMigrationComponent(f.MongoFeature)
	if err != nil {
		panic(err)
	}

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		if f.MongoFeature == nil {
			f.MongoFeature = componenttest.NewMongoFeature(mongoOpts)
		}
		component.Reset()

		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if closeErr := component.Close(); closeErr != nil {
			panic(closeErr)
		}

		return ctx, nil
	})

	f.MongoFeature.RegisterSteps(ctx)
	component.RegisterSteps(ctx)
}

func (f *ComponentTest) InitializeTestSuite(ctx *godog.TestSuiteContext) {

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
