@Migrate @MigrateStaticDataset
Feature: Migrate a static dataset job

  Background:
    Given an admin user has the "migrations:read" permission
    And I am an admin user
    And the migration service is running

  Scenario: Migrate a static dataset job
    Given a get page data request to zebedee for "/test-source-id" returns with status 200 and payload:
      """
      {
        "type": "dataset_landing_page",
        "title": "Test Dataset Series",
        "datasets": [
          {
            "id": "/test-dataset-1",
            "title": "Test Dataset 1"
          },
          {
            "id": "/test-dataset-2",
            "title": "Test Dataset 2"
          }
        ]
      }
      """
    And the following document exists in the "jobs" collection:
      """
      {
        "_id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
        "last_updated": "2025-11-19T13:28:00Z",
        "links": {
          "self": {
            "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
          }
        },
        "state": "submitted",
        "config": {
          "source_id": "/test-source-id",
          "target_id": "test-target-id",
          "type": "static_dataset"
        }
      }
      """
    And I wait 3 seconds for the job processor to process jobs but not tasks
    When I GET "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
    Then I should receive the following JSON response with status "200":
      """
      {
        "id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
        "state": "migrating",
        "last_updated": "{{DYNAMIC_TIMESTAMP}}",
        "config": {
          "source_id": "/test-source-id",
          "target_id": "test-target-id",
          "type": "static_dataset"
        },
        "links": {
          "self": {
            "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
          }
        }
      }
      """
    When I GET "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/tasks"
    Then I should receive the following JSON response with status "200":
      """
      {
        "items": [
          {
            "id": "{{DYNAMIC_UUID}}",
            "job_id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
            "type": "dataset_series",
            "state": "submitted",
            "last_updated": "{{DYNAMIC_RECENT_TIMESTAMP}}",
            "links": {
              "self": {
                "href": "{{DYNAMIC_URI_PATH}}"
              },
              "job": {
                "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
              }
            },
            "source": {
              "id": "/test-dataset-1",
              "label": ""
            },
            "target": {
              "id": "test-target-id",
              "label": ""
            }
          }
        ],
        "count": 1,
        "total_count": 1,
        "offset": 0,
        "limit": 10
      }
      """

  Scenario: Failed to migrate a dataset series
    Given a get page data request to zebedee for "/test-source-id" returns a page of type "error" with status 404
    And the following document exists in the "jobs" collection:
      """
      {
        "_id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
        "last_updated": "2025-11-19T13:28:00Z",
        "links": {
          "self": {
            "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
          }
        },
        "state": "submitted",
        "config": {
          "source_id": "/test-source-id",
          "target_id": "test-target-id",
          "type": "static_dataset"
        }
      }
      """
    And I wait 4 seconds for the job processor to process tasks and jobs
    When I GET "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
    Then I should receive the following JSON response with status "200":
      """
      {
        "id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
        "state": "failed_migration",
        "last_updated": "{{DYNAMIC_TIMESTAMP}}",
        "config": {
          "source_id": "/test-source-id",
          "target_id": "test-target-id",
          "type": "static_dataset"
        },
        "links": {
          "self": {
            "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
          }
        }
      }
      """
    When I GET "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/tasks"
    Then I should receive the following JSON response with status "200":
      """
      {
        "items": [
          {
            "id": "{{DYNAMIC_UUID}}",
            "job_id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
            "type": "dataset_series",
            "state": "failed_migration",
            "last_updated": "{{DYNAMIC_RECENT_TIMESTAMP}}",
            "links": {
              "self": {
                "href": "{{DYNAMIC_URI_PATH}}"
              },
              "job": {
                "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
              }
            },
            "source": {
              "id": "/test-dataset-1",
              "label": ""
            },
            "target": {
              "id": "test-target-id",
              "label": ""
            }
          }
        ],
        "count": 1,
        "total_count": 1,
        "offset": 0,
        "limit": 10
      }
      """



