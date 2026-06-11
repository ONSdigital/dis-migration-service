@Publish @PublishStaticDataset
Feature: Publish a static dataset job

  Background:
     Given an admin user has the following permissions as JSON:
        """
        {
            "migrations:read": {
                "groups/role-admin": [{"id": "1"}]
            },
            "migrations:edit": {
                "groups/role-admin": [{"id": "1"}]
            }
        }
        """
     And I am an admin user
     And the migration service is running


  Scenario: Publish a static dataset job
    Given the following document exists in the "jobs" collection:
      """
      {
        "_id": "3985ff0f-2d46-51gg-022g-6b33f8173802",
        "job_number": 30,
        "last_updated": "2025-11-19T13:28:00Z",
        "links": {
          "self": {
            "href": "/v1/migration-jobs/30"
          }
        },
        "state": "in_review",
        "label": "Test Publish Dataset Series",
        "config": {
          "collection_id": "migration-job-test-collection",
          "source_id": "/test-static-dataset-job",
          "target_id": "test-target-id",
          "type": "static_dataset"
        }
      }
      """
    And the following document exists in the "tasks" collection:
      """
      {
        "_id": "task-223e4567-e89b-12d3-a456-426614174000",
        "job_number": 30,
        "last_updated": "2025-11-19T13:30:00Z",
        "state": "in_review",
        "type": "dataset_series",
        "source": {
          "id": "/test-static-dataset-job"
        },
        "target": {
          "id": "test-target-id"
        },
        "links": {
          "self": {
            "href": "/v1/migration-jobs/30/tasks/task-223e4567-e89b-12d3-a456-426614174000"
          },
          "job": {
            "href": "/v1/migration-jobs/30"
          }
        }
      }
      """
    When I PUT "/v1/migration-jobs/30/state"
      """
      {
        "state": "approved"
      }
      """
    Then the HTTP status code should be "204"
    # Post-publish runs automatically once job reaches published state
    And I wait 5 seconds for the job processor to process tasks and jobs
    When I GET "/v1/migration-jobs/30"
    Then I should receive the following JSON response with status "200":
      """
      {
        "id": "3985ff0f-2d46-51gg-022g-6b33f8173802",
        "job_number": 30,
        "state": "completed",
        "last_updated": "{{DYNAMIC_TIMESTAMP}}",
        "config": {
          "collection_id": "migration-job-test-collection",
          "source_id": "/test-static-dataset-job",
          "target_id": "test-target-id",
          "type": "static_dataset"
        },
        "label": "Test Publish Dataset Series",
        "links": {
          "self": {
            "href": "/v1/migration-jobs/30"
          }
        }
      }
      """
    When I GET "/v1/migration-jobs/30/tasks"
    Then I should receive the following JSON response with status "200":
      """
      {
        "items": [
          {
            "id": "task-223e4567-e89b-12d3-a456-426614174000",
            "job_number": 30,
            "type": "dataset_series",
            "state": "completed",
            "last_updated": "{{DYNAMIC_RECENT_TIMESTAMP}}",
            "links": {
              "self": {
                "href": "/v1/migration-jobs/30/tasks/task-223e4567-e89b-12d3-a456-426614174000"
              },
              "job": {
                "href": "/v1/migration-jobs/30"
              }
            },
            "source": {
              "id": "/test-static-dataset-job"
            },
            "target": {
              "id": "test-target-id"
            }
          }
        ],
        "count": 1,
        "total_count": 1,
        "offset": 0,
        "limit": 10
      }
      """
