@Job
Feature: Create a Job

  Scenario: Create a job with valid input
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
       "id": "{{DYNAMIC_ID}}",
       "last_updated": "{{DYNAMIC_RECENT_TIMESTAMP}}",
       "state": "submitted",
       "config": {
           "source_id": "/test-source-id",
           "target_id": "test-target-id",
           "type": "static_dataset"
       }
   }
   """

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
                "description": "source ID is invalid"
            },
            {
                "code": 400,
                "description": "target ID is invalid"
            }
        ]
    }
    """

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
