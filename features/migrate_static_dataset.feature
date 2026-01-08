@Migrate @MigrateStaticDataset
Feature: Migrate a static dataset job

  Background:
    Given an admin user has the "migrations:read" permission
    And I am an admin user
    And the migration service is running

  Scenario: Migrate a static dataset job
    Given a get page data request to zebedee for "/test-static-dataset-job" returns with status 200 and payload:
      """
      {
        "type": "dataset_landing_page",
        "description": {
          "title": "Test Dataset Series"
        },
        "datasets": [
          {
            "id": "/test-edition-1",
            "title": "Edition 1"
          }
        ]
      }
      """
    And a get page data request to zebedee for "/test-edition-1" returns with status 200 and payload:
      """
      {
        "type": "dataset",
        "description": {
          "title": "Edition 1"
        }
      }
      """
    And the Dataset API responds successfully to create dataset requests
    And the following document exists in the "jobs" collection:
      """
      {
        "_id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
        "job_number": 20,
        "last_updated": "2025-11-19T13:28:00Z",
        "links": {
          "self": {
            "href": "/v1/migration-jobs/20"
          }
        },
        "state": "submitted",
        "label": "Test Dataset Series",
        "config": {
          "source_id": "/test-static-dataset-job",
          "target_id": "test-target-id",
          "type": "static_dataset"
        }
      }
      """
    And I wait 3 seconds for the job processor to process tasks and jobs
    When I GET "/v1/migration-jobs/20"
    Then I should receive the following JSON response with status "200":
      """
      {
        "id":"2874ee9e-1cec-44f8-9b6d-998cf2062791",
        "job_number": 20,
        "state": "migrating",
        "last_updated": "{{DYNAMIC_TIMESTAMP}}",
        "config": {
          "source_id": "/test-static-dataset-job",
          "target_id": "test-target-id",
          "type": "static_dataset"
        },
        "label": "Test Dataset Series",
        "links": {
          "self": {
            "href": "/v1/migration-jobs/20"
          }
        }
      }
      """

  Scenario: Failed to migrate a dataset series
    Given a get page data request to zebedee for "/test-static-dataset-job-failed" returns with status 404 and payload:
      """
      page not found
      """
    And the following document exists in the "jobs" collection:
      """
      {
        "_id": "13de71de-4201-494a-851b-3d65575235e6",
        "job_number": 21,
        "last_updated": "2025-11-19T13:28:00Z",
        "links": {
          "self": {
            "href": "/v1/migration-jobs/13de71de-4201-494a-851b-3d65575235e6"
          }
        },
        "state": "submitted",
        "label": "Test Failed Dataset Series",
        "config": {
          "source_id": "/test-static-dataset-job-failed",
          "target_id": "test-target-id",
          "type": "static_dataset"
        }
      }
      """
    And I wait 3 seconds for the job processor to process tasks and jobs
    When I GET "/v1/migration-jobs/21"
    Then I should receive the following JSON response with status "200":
      """
      {
        "id": "13de71de-4201-494a-851b-3d65575235e6",
        "job_number": 21,
        "state": "failed_migration",
        "last_updated": "{{DYNAMIC_TIMESTAMP}}",
        "config": {
          "source_id": "/test-static-dataset-job-failed",
          "target_id": "test-target-id",
          "type": "static_dataset"
        },
        "label": "Test Failed Dataset Series",
        "links": {
          "self": {
            "href": "/v1/migration-jobs/13de71de-4201-494a-851b-3d65575235e6"
          }
        }
      }
      """
    When I GET "/v1/migration-jobs/21/tasks"
    Then I should receive the following JSON response with status "200":
      """
      {
        "items": [
          {
            "id": "{{DYNAMIC_UUID}}",
            "job_number": 21,
            "type": "dataset_series",
            "state": "failed_migration",
            "last_updated": "{{DYNAMIC_RECENT_TIMESTAMP}}",
            "links": {
              "self": {
                "href": "{{DYNAMIC_URI_PATH}}"
              },
              "job": {
                "href": "/v1/migration-jobs/13de71de-4201-494a-851b-3d65575235e6"
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
