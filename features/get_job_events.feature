@Job @GetJobEvents
Feature: Get list of job events

  Rule: User that is authorised and authenticated
    Background:
      Given an admin user has the "migrations:read" permission
      And I am an admin user
      And the migration service is running

    Scenario: Get a list of 1 event with defaults for existing job
      Given the following document exists in the "jobs" collection:
        """
        {
          "_id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
          "last_updated": "2025-11-19T13:28:00Z",
          "links": {
            "self": {
              "href": "http://localhost:30100/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
            },
            "events": {
              "href": "http://localhost:30100/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/events"
            },
            "tasks": {
              "href": "http://localhost:30100/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/tasks"
            }
          },
          "state": "submitted",
          "type": "static_dataset",
          "config": {
            "source_id": "test-source-id",
            "target_id": "test-target-id",
            "type": "static_dataset"
          }
        }
        """
      And the following document exists in the "events" collection:
        """
        {
          "_id": "event-123e4567-e89b-12d3-a456-426614174000",
          "job_id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
          "created_at": "2025-11-19T13:30:00Z",
          "action": "submitted",
          "requested_by": {
            "id": "user-123",
            "email": "publisher@ons.gov.uk"
          },
          "links": {
            "self": {
              "href": "http://localhost:30100/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/events/event-123e4567-e89b-12d3-a456-426614174000"
            },
            "job": {
              "href": "http://localhost:30100/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
            }
          }
        }
        """
      When I GET "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/events"
      Then I should receive the following JSON response with status "200":
        """
        {
          "count": 1,
          "items": [
            {
              "id": "event-123e4567-e89b-12d3-a456-426614174000",
              "job_id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
              "created_at": "2025-11-19T13:30:00Z",
              "action": "submitted",
              "requested_by": {
                "id": "user-123",
                "email": "publisher@ons.gov.uk"
              },
              "links": {
                "self": {
                  "href": "http://localhost:30100/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/events/event-123e4567-e89b-12d3-a456-426614174000"
                },
                "job": {
                  "href": "http://localhost:30100/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
                }
              }
            }
          ],
          "limit": 10,
          "offset": 0,
          "total_count": 1
        }
        """

    Scenario: Get a list of events for non-existent job returns 404
      When I GET "/v1/migration-jobs/non-existent-job-id/events"
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

    Scenario: Get a list of events when no events are available for existing job
      Given the following document exists in the "jobs" collection:
        """
        {
          "_id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
          "last_updated": "2025-11-19T13:28:00Z",
          "links": {
            "self": {
              "href": "http://localhost:30100/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
            },
            "events": {
              "href": "http://localhost:30100/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/events"
            },
            "tasks": {
              "href": "http://localhost:30100/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/tasks"
            }
          },
          "state": "submitted",
          "type": "static_dataset",
          "config": {
            "source_id": "test-source-id",
            "target_id": "test-target-id",
            "type": "static_dataset"
          }
        }
        """
      When I GET "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/events"
      Then I should receive the following JSON response with status "200":
        """
        {
          "count": 0,
          "items": [],
          "limit": 10,
          "offset": 0,
          "total_count": 0
        }
        """

    @InvalidInput
    Scenario: Get a list of events with limit exceeding maximum allowed
      Given the following document exists in the "jobs" collection:
        """
        {
          "_id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
          "last_updated": "2025-11-19T13:28:00Z",
          "links": {
            "self": {
              "href": "http://localhost:30100/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
            },
            "events": {
              "href": "http://localhost:30100/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/events"
            },
            "tasks": {
              "href": "http://localhost:30100/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/tasks"
            }
          },
          "state": "submitted",
          "type": "static_dataset",
          "config": {
            "source_id": "test-source-id",
            "target_id": "test-target-id",
            "type": "static_dataset"
          }
        }
        """
      When I GET "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/events?limit=2000"
      Then I should receive the following JSON response with status "400":
        """
        {
          "errors": [
            {
              "code": 400,
              "description": "limit parameter exceeds maximum allowed"
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