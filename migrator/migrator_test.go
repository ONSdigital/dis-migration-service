package migrator

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMigratorShutdown(t *testing.T) {
	Convey("Given a migrator with ongoing migrations", t, func() {
		mig := &migrator{}
		mig.wg.Add(1)

		// Simulate background task completion
		go func() {
			time.Sleep(10 * time.Millisecond)
			mig.wg.Done()
		}()

		ctx := context.Background()
		Convey("When Shutdown is called", func() {
			err := mig.Shutdown(ctx)
			Convey("Then it waits for migrations to complete and returns nil error", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given a migrator with ongoing migrations that do not complete before context timeout", t, func() {
		mig := &migrator{}
		mig.wg.Add(1) // Simulate ongoing migration

		// Create a context that will timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		Convey("When Shutdown is called", func() {
			err := mig.Shutdown(ctx)
			Convey("Then it returns a timeout error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "timed out waiting for background tasks to complete")
			})
		})
	})
}
