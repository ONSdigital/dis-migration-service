@Job @GetJobs
Feature: Get list of jobs

  Rule: User that is authorised and authenticated
    Background:
      Given an admin user has the "migrations:read" permission
      And I am an admin user
      And the migration service is running

    Scenario: Get a list of 1 job with defaults
      Given the following document exists in the "jobs" collection:
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
            "source_id": "test-source-id",
            "target_id": "test-target-id",
            "type": "test-type"
          }
        }
        """
      When I GET "/v1/migration-jobs"
      Then I should receive the following JSON response with status "200":
        """
        {
          "count": 1,
          "items": [
            {
              "id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
              "last_updated": "2025-11-19T13:28:00Z",
              "links": {
                "self": {
                  "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
                }
              },
              "state": "submitted",
              "config": {
                "source_id": "test-source-id",
                "target_id": "test-target-id",
                "type": "test-type"
              }
            }
          ],
          "limit": 10,
          "offset": 0,
          "total_count": 1
        }
        """


    Scenario: Get a list of 2 jobs using limit and offset to paginate
      Given the following document exists in the "jobs" collection:
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
            "source_id": "test-source-id",
            "target_id": "test-target-id",
            "type": "test-type"
          }
        }
        """
      And the following document exists in the "jobs" collection:
        """
        {
          "_id": "4874ee9e-1cec-44f8-9b6d-998cf2062791",
          "last_updated": "2025-11-20T10:15:00Z",
          "links": {
            "self": {
              "href": "/v1/migration-jobs/4874ee9e-1cec-44f8-9b6d-998cf2062791"
            }
          },
          "state": "in_progress",
          "config": {
            "source_id": "another-source-id",
            "target_id": "another-target-id",
            "type": "another-type"
          }
        }
        """
      When I GET "/v1/migration-jobs"
      Then I should receive the following JSON response with status "200":
        """
        {
          "count": 2,
          "items": [
            {
              "id": "4874ee9e-1cec-44f8-9b6d-998cf2062791",
              "last_updated": "2025-11-20T10:15:00Z",
              "links": {
                "self": {
                  "href": "/v1/migration-jobs/4874ee9e-1cec-44f8-9b6d-998cf2062791"
                }
              },
              "state": "in_progress",
              "config": {
                "source_id": "another-source-id",
                "target_id": "another-target-id",
                "type": "another-type"
              }
            },
            {
              "id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
              "last_updated": "2025-11-19T13:28:00Z",
              "links": {
                "self": {
                  "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
                }
              },
              "state": "submitted",
              "config": {
                "source_id": "test-source-id",
                "target_id": "test-target-id",
                "type": "test-type"
              }
            }
          ],
          "limit": 10,
          "offset": 0,
          "total_count": 2
        }
        """
      When I GET "/v1/migration-jobs?limit=1&offset=1"
      Then I should receive the following JSON response with status "200":
        """
        {
          "count": 1,
          "items": [
            {
              "id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
              "last_updated": "2025-11-19T13:28:00Z",
              "links": {
                "self": {
                  "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
                }
              },
              "state": "submitted",
              "config": {
                "source_id": "test-source-id",
                "target_id": "test-target-id",
                "type": "test-type"
              }
            }
          ],
          "limit": 1,
          "offset": 1,
          "total_count": 2
        }
        """
      When I GET "/v1/migration-jobs?limit=1"
      Then I should receive the following JSON response with status "200":
        """
        {
          "count": 1,
          "items": [
            {
              "id": "4874ee9e-1cec-44f8-9b6d-998cf2062791",
              "last_updated": "2025-11-20T10:15:00Z",
              "links": {
                "self": {
                  "href": "/v1/migration-jobs/4874ee9e-1cec-44f8-9b6d-998cf2062791"
                }
              },
              "state": "in_progress",
              "config": {
                "source_id": "another-source-id",
                "target_id": "another-target-id",
                "type": "another-type"
              }
            }
          ],
          "limit": 1,
          "offset": 0,
          "total_count": 2
        }
        """

    @InvalidInput
    Scenario: Get a list of jobs with limit exceeding maximum allowed
      When I GET "/v1/migration-jobs?limit=2000"
      Then I should receive the following JSON response with status "400":
        """
        {
          "errors": [
            {
              "code": 400,
              "description": "limit parameter exceeds maximum allowed"
            }
          ]
        }
        """

    @InvalidInput
    Scenario: Get a list of jobs with offset and limit invalid
      When I GET "/v1/migration-jobs?limit=invalid&offset=-10"
      Then I should receive the following JSON response with status "400":
        """
        {
          "errors": [
            {
              "code": 400,
              "description": "limit parameter is invalid"
            },
            {
              "code": 400,
              "description": "offset parameter is invalid"
            }
          ]
        }
        """

    Scenario: Get a list of jobs when no jobs are available
      When I GET "/v1/migration-jobs"
      Then I should receive the following JSON response with status "200":
        """
        {
          "count": 0,
          "items": [],
          "limit": 10,
          "offset": 0,
          "total_count": 0
        }
        """

  @Auth
  Rule: Users that are not authorised or authenticated
    Background:
      Given an admin user has the "incorrect" permission
      And the migration service is running

    Scenario: User that is not authenticated
      Given I am not authorised
      And the migration service is running
      When I GET "/v1/migration-jobs"
      Then the HTTP status code should be "401"

    Scenario: User that is not authorised
      Given I am an admin user
      When I GET "/v1/migration-jobs"
      Then the HTTP status code should be "403"
