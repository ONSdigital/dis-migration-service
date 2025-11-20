@Job @GetJob
Feature: Get a Job

  Scenario: Get a job which exists
    Given the following document exists in the "jobs" collection:
    """
    {
        "_id": "1",
        "last_updated": "2025-11-19T13:28:00Z",
        "state": "submitted",
        "config": {
            "source_id": "test-source-id",
            "target_id": "test-target-id",
            "type": "test-type"
        }
    }
    """
    When I GET "/v1/migration-jobs/1"
    Then I should receive the following JSON response with status "200":
    """
    {
        "id": "1",
        "last_updated": "2025-11-19T13:28:00Z",
        "state": "submitted",
        "config": {
            "source_id": "test-source-id",
            "target_id": "test-target-id",
            "type": "test-type"
        },
        "links":{}
    }
    """
