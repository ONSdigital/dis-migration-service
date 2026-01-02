@Job @CreateJob
Feature: Create a Job

  Rule: User that is authorised and authenticated
    Background:
      Given an admin user has the "migrations:create" permission
      And I am an admin user
      And the migration service is running

    Scenario: Create a job with valid input
      Given a get page data request to zebedee for "/test-source-id" returns with status 200 and payload:
        """
        {
          "type": "dataset_landing_page",
          "description": {
            "title": "Test Dataset Series"
          }
        }
        """
      And a get dataset request to the dataset API for "test-target-id" returns with status 404
      When I POST "/v1/migration-jobs"
        """
        {
          "source_id": "/test-source-id",
          "target_id": "test-target-id",
          "type": "static_dataset"
        }
        """
      Then I should receive the following JSON response with status "202":
        """
        {
          "job_number":1,
          "last_updated": "{{DYNAMIC_RECENT_TIMESTAMP}}",
          "label": "Test Dataset Series",
          "state": "submitted",
          "config": {
            "source_id": "/test-source-id",
            "target_id": "test-target-id",
            "type": "static_dataset"
          },
          "links": {
            "self": {
              "href": "{{DYNAMIC_URI_PATH}}"
            },
            "tasks": {
              "href": "{{DYNAMIC_URI_PATH}}"
            },
            "events": {
              "href": "{{DYNAMIC_URI_PATH}}"
            }
          }
        }
        """

    @InvalidInput
    Scenario: Create a job with invalid input
      When I POST "/v1/migration-jobs"
        """
        """
      Then I should receive the following JSON response with status "400":
        """
        {
          "errors": [
            {
              "code": 400,
              "description": "unable to read submitted body"
            }
          ]
        }
        """

    @InvalidInput
    Scenario: Create a static_dataset job with missing or invalid parameters
      When I POST "/v1/migration-jobs"
        """
        {
          "source_id": "test-source-id",
          "target_id": "bad target id",
          "type": "static_dataset"
        }
        """
      Then I should receive the following JSON response with status "400":
        """
        {
          "errors": [
            {
              "code": 400,
              "description": "source ID URI path must start with '/', not end with '/', not contain query strings or hashbangs"
            },
            {
              "code": 400,
              "description": "target id must be lowercase alphanumeric with optional hyphen separators"
            }
          ]
        }
        """

    @InvalidInput
    Scenario: Create a job with invalid job type
      When I POST "/v1/migration-jobs"
        """
        {
          "source_id": "test-source-id",
          "target_id": "bad target id",
          "type": "invalid_type"
        }
        """
      Then I should receive the following JSON response with status "400":
        """
        {
          "errors": [
            {
              "code": 400,
              "description": "job type is invalid"
            }
          ]
        }
        """

    @InvalidUpstream
    Scenario: Create a job with a target already existing
      Given a get page data request to zebedee for "/test-source-id" returns with status 200 and payload:
        """
        {
          "type": "dataset_landing_page",
          "description": {
            "title": "Test Dataset Series"
          }
        }
        """
      And a get dataset request to the dataset API for "test-target-id" returns with status 200
      When I POST "/v1/migration-jobs"
        """
        {
          "source_id": "/test-source-id",
          "target_id": "test-target-id",
          "type": "static_dataset"
        }
        """
      Then I should receive the following JSON response with status "400":
        """
        {
          "errors": [
            {
              "code": 400,
              "description": "target ID is invalid"
            }
          ]
        }
        """

    @InvalidUpstream
    Scenario: Create a job with an invalid source type
      Given a get page data request to zebedee for "/test-incorrect-source" returns with status 200 and payload:
        """
        {
          "type": "dataset",
          "description": {
            "title": "Test Dataset Series"
          }
        }
        """
      When I POST "/v1/migration-jobs"
        """
        {
          "source_id": "/test-incorrect-source",
          "target_id": "test-target-id",
          "type": "static_dataset"
        }
        """
      Then I should receive the following JSON response with status "400":
        """
        {
          "errors": [
            {
              "code": 400,
              "description": "source ID is invalid"
            }
          ]
        }
        """

    @InvalidUpstream
    Scenario: Create a job with a target already existing
      Given a get page data request to zebedee for "/test-source-id" returns with status 200 and payload:
        """
        {
          "type": "dataset_landing_page",
          "description": {
            "title": "Test Dataset Series"
          }
        }
        """
      And a get dataset request to the dataset API for "test-target-id" returns with status 200
      When I POST "/v1/migration-jobs"
        """
        {
          "source_id": "/test-source-id",
          "target_id": "test-target-id",
          "type": "static_dataset"
        }
        """
      Then I should receive the following JSON response with status "400":
        """
        {
          "errors": [
            {
              "code": 400,
              "description": "target ID is invalid"
            }
          ]
        }
        """

    @StoreError
    Scenario: Create a job which is already running
      Given the following document exists in the "jobs" collection:
        """
        {
          "id": "1",
          "state": "submitted",
          "config": {
            "source_id": "/test-source-id",
            "target_id": "test-target-id",
            "type": "static_dataset"
          }
        }
        """
      And a get page data request to zebedee for "/test-source-id" returns with status 200 and payload:
        """
        {
          "type": "dataset_landing_page",
          "description": {
            "title": "Test Dataset Series"
          }
        }
        """
      And a get dataset request to the dataset API for "test-target-id" returns with status 404
      When I POST "/v1/migration-jobs"
        """
        {
          "source_id": "/test-source-id",
          "target_id": "test-target-id",
          "type": "static_dataset"
        }
        """
      Then I should receive the following JSON response with status "409":
        """
        {
          "errors": [
            {
              "code": 409,
              "description": "job already running"
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
      When I POST "/v1/migration-jobs"
        """
        {
          "source_id": "/test-source-id",
          "target_id": "test-target-id",
          "type": "static_dataset"
        }
        """
      Then the HTTP status code should be "401"

    Scenario: User that is not authorised
      Given I am an admin user
      When I POST "/v1/migration-jobs"
        """
        {
          "source_id": "/test-source-id",
          "target_id": "test-target-id",
          "type": "static_dataset"
        }
        """
      Then the HTTP status code should be "403"
