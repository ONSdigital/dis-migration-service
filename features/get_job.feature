Feature: Get a Job

  Scenario: Get a job which exists
    When I GET "/v1/migration-jobs/1"
    Then I should receive the following JSON response with status "200":
    """
    {
        "id": "1",
        "last_updated": "test-time",
        "state": "submitted",
        "config": {
            "source_id": "test-source-id",
            "target_id": "test-target-id",
            "type": "test-type"
        }
    }
    """
