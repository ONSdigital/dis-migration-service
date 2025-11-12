@Job @CreateJob
Feature: Create a Job
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
  Scenario: Create a job with an invalid source type
    Given a get page data request to zebedee for "/test-source-id" returns a page of type "bulletin" with status 200
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
  Scenario: Create a job with a target already existing
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

  @StoreError
  Scenario: Create a job with valid input
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
    And a get page data request to zebedee for "/test-source-id" returns a page of type "dataset_landing_page" with status 200
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

