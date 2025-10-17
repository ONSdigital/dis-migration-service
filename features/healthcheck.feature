Feature: Health endpoint

    Background: Service setup
        Given the migration service is running

    Scenario: Returning a OK (200) status when health endpoint called
        Given mongodb is healthy
        And all its expected collections exist
        And I have a healthcheck interval of 1 second
        And I wait 2 seconds for the healthcheck to be available
        When I GET "/health"
        Then the health checks should have completed within 6 seconds
        And I should receive the following health JSON response:
        """
            {
              "status": "OK",
              "version": {
                "git_commit": "6584b786caac36b6214ffe04bf62f058d4021538",
                "language": "go",
                "language_version": "go1.25.0",
                "version": "v1.2.3"
              },
              "checks": [
                {
                  "name": "Mongo DB",
                  "status": "OK",
                  "message": "mongodb is OK and all expected collections exist"
                }
              ]
            }
        """

Scenario: Returning a WARNING (429) status when health endpoint called
    Given mongodb is healthy
    And all its expected collections exist
    And mongodb stops running
    And I have a healthcheck interval of 1 second
    And I wait 2 seconds for the healthcheck to be available
    When I GET "/health"
    Then the HTTP status code should be "429"
    And the response header "Content-Type" should be "application/json; charset=utf-8"
    When the health checks should have completed within 4 seconds
    Then I should receive the following health JSON response:
    """
        {
            "status": "WARNING",
            "version": {
                "git_commit": "6584b786caac36b6214ffe04bf62f058d4021538",
                "language": "go",
                "language_version": "go1.17.8",
                "version": "v1.2.3"
            },
            "checks": [
                {
                    "name": "Mongo DB",
                    "status": "CRITICAL",
                    "message": "Failed to ping datastore: client is disconnected"
                }
            ]
        }
    """
#
#    Scenario: Returning a CRITICAL (500) status when health endpoint called
#        Given redis stops running
#        And I have a healthcheck interval of 1 second
#        And I wait 6 seconds to pass the critical timeout
#        And I GET "/health"
#        Then the HTTP status code should be "500"
#        And the response header "Content-Type" should be "application/json; charset=utf-8"
#        When the health checks should have completed within 6 seconds
#        Then I should receive the following health JSON response:
#        """
#            {
#                "status": "CRITICAL",
#                "version": {
#                    "git_commit": "6584b786caac36b6214ffe04bf62f058d4021538",
#                    "language": "go",
#                    "language_version": "go1.17.8",
#                    "version": "v1.2.3"
#                },
#                "checks": [
#                    {
#                        "name": "Redis",
#                        "status": "CRITICAL",
#                        "status_code": 500,
#                        "message": "couldn't connect to redis"
#                    }
#                ]
#            }
#        """
