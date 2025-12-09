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
        "label": "Test Dataset Series",
        "config": {
          "source_id": "/test-static-dataset-job",
          "target_id": "test-target-id",
          "type": "static_dataset"
        }
      }
      """
    And I wait 3 seconds for the job processor to process tasks and jobs
    When I GET "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
    Then I should receive the following JSON response with status "200":
      """
      {
        "id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
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
            "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
          }
        }
      }
      """
