@Job
Feature: Create a Job
  @Linden
  Scenario: Create a job with valid input
    Given a get page data request to zebedee for "/test-source-id" returns a page of type "dataset_landing_page" with status 200
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
        "id": "{{DYNAMIC_UUID}}",
        "last_updated": "{{DYNAMIC_RECENT_TIMESTAMP}}",
        "state": "submitted",
        "config": {
          "source_id": "/test-source-id",
          "target_id": "test-target-id",
          "type": "static_dataset"
        },
        "links": {
          "self": {
            "href": "{{DYNAMIC_URL}}"
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
            "description": "source ID is invalid"
          },
          {
            "code": 400,
            "description": "target ID is invalid"
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
  Scenario: Create a job without a source
    Given a get page data request to zebedee for "/test-source-id" returns a page of type "dataset_landing_page" with status 404
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
            "description": "source ID is invalid"
          }
        ]
      }
      """

  @InvalidUpstream
  Scenario: Create a job without a source
    Given a get page data request to zebedee for "/test-source-id" returns a page of type "dataset_landing_page" with status 200
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
