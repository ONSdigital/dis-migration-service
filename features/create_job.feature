#Feature: Create a Job
#
#  Scenario: Create a job with valid input
#    When I POST "/v1/migration-jobs"
#    """
#    {
#        "source_id": "test-source-id",
#        "target_id": "test-target-id",
#        "type": "dataset"
#    }
#    """
#    Then I should receive the following JSON response with status "202":
#    """
#    {
#        "id": "test-id",
#        "last_updated": "test-time",
#        "state": "submitted",
#        "config": {
#            "source_id": "test-source-id",
#            "target_id": "test-target-id",
#            "type": "dataset"
#        }
#    }
#    """
#
#  Scenario: Create a job with invalid input
#    When I POST "/v1/migration-jobs"
#    """
#    """
#    Then I should receive the following JSON response with status "400":
#    """
#    {
#        "errors": [
#            {
#                "code": 400,
#                "description": "unable to read submitted body"
#            }
#        ]
#    }
#    """
#
#    Scenario: Create a job with missing parameters
#    When I POST "/v1/migration-jobs"
#    """
#    {
#        "source_id": "test-source-id"
#    }
#    """
#    Then I should receive the following JSON response with status "400":
#    """
#    {
#        "errors": [
#            {
#                "code": 400,
#                "description": "target ID not provided"
#            },
#            {
#                "code": 400,
#                "description": "job type not provided"
#            }
#        ]
#    }
#    """
