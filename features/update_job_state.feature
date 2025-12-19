@Job @UpdateJobState
Feature: Update Job State

  Rule: User that is authorised and authenticated
    Background:
      Given an admin user has the "migrations:edit" permission
      And I am an admin user
      And the migration service is running

    Scenario: Update a job state successfully
      Given the following document exists in the "jobs" collection:
        """
        {
          "_id": "1",
          "state": "in_review"
        }
        """
      When I PUT "/v1/migration-jobs/1/state"
        """
        {
          "state": "approved"
        }
        """
      Then the HTTP status code should be "204"

    @InvalidInput
    Scenario: Update job state with an empty body
      When I PUT "/v1/migration-jobs/1/state"
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
    Scenario: Update job state with an invalid state
      When I PUT "/v1/migration-jobs/1/state"
        """
        {
          "state": "not_a_state"
        }
        """
      Then I should receive the following JSON response with status "400":
        """
        {
          "errors": [
            {
              "code": 400,
              "description": "job state parameter is invalid"
            }
          ]
        }
        """

    @InvalidInput
    Scenario: Update job state with a state not allowed for this endpoint
      When I PUT "/v1/migration-jobs/1/state"
        """
        {
          "state": "migrating"
        }
        """
      Then I should receive the following JSON response with status "400":
        """
        {
          "errors": [
            {
              "code": 400,
              "description": "state not allowed for this endpoint"
            }
          ]
        }
        """

    @InvalidTransition
    Scenario: Update job state with an invalid state transition
      Given the following document exists in the "jobs" collection:
        """
        {
          "_id": "1",
          "state": "completed"
        }
        """
      When I PUT "/v1/migration-jobs/1/state"
        """
        {
          "state": "approved"
        }
        """
      Then I should receive the following JSON response with status "409":
        """
        {
          "errors": [
            {
              "code": 409,
              "description": "state change is not allowed"
            }
          ]
        }
        """

    Scenario: Update a job that does not exist
      When I PUT "/v1/migration-jobs/4000/state"
        """
        {
          "state": "approved"
        }
        """
      Then I should receive the following JSON response with status "404":
        """
        {
          "errors": [
            {
              "code": 404,
              "description": "job not found"
            }
          ]
        }
        """

    @Auth
    Rule: Users that are not authorised or authenticated
    Background:
      Given an admin user has the "incorrect" permission
      And the migration service is running

    Scenario: User that is not authenticated
      Given I am not authorised
      When I PUT "/v1/migration-jobs/1/state"
        """
        {
          "state": "approved"
        }
        """
      Then the HTTP status code should be "401"

    Scenario: User that is not authorised
      Given I am an admin user
      When I PUT "/v1/migration-jobs/1/state"
        """
        {
          "state": "approved"
        }
        """
      Then the HTTP status code should be "403"