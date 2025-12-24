@Job @GetJob
Feature: Get a Job

  Rule: User that is authorised and authenticated
    Background:
      Given an admin user has the "migrations:read" permission
      And I am an admin user
      And the migration service is running

    Scenario: Get a job which exists
      Given the following document exists in the "jobs" collection:
        """
        {
          "_id": "1",
          "job_number": 19,
          "label": "Labour Market statistics",
          "last_updated": "2025-11-19T13:28:00Z",
          "state": "migrating",
          "config": {
            "source_id": "test-source-id",
            "target_id": "test-target-id",
            "type": "test-type"
          }
        }
        """
      When I GET "/v1/migration-jobs/19"
      Then I should receive the following JSON response with status "200":
        """
        {
          "job_number": 19,
          "label": "Labour Market statistics",
          "last_updated": "2025-11-19T13:28:00Z",
          "state": "migrating",
          "config": {
            "source_id": "test-source-id",
            "target_id": "test-target-id",
            "type": "test-type"
          },
          "links": {}
        }
        """

    Scenario: Get a job which doesn't exist
      When I GET "/v1/migration-jobs/4000"
      Then I should receive the following JSON response with status "404":
        """
        {
          "errors": [
            {
              "code": 404,
              "description": "job not found"
            }
          ]
        }
        """

  @Auth
  Rule: Users that are not authorised or authenticated
    Background:
      Given an admin user has the "incorrect" permission
      And the migration service is running

    Scenario: User that is not authenticated
      Given I am not authorised
      And the migration service is running
      When I GET "/v1/migration-jobs/1"
      Then the HTTP status code should be "401"

    Scenario: User that is not authorised
      Given I am an admin user
      When I GET "/v1/migration-jobs/1"
      Then the HTTP status code should be "403"


