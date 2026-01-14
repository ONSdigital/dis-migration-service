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
          "job_number": 12,
          "label": "Labour Market statistics",
          "last_updated": "2025-11-19T13:28:00Z",
          "links": {
            "self": {
              "href": "/v1/migration-jobs/12"
            }
          },
          "state": "migrating",
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
              "job_number": 12,
              "label": "Labour Market statistics",
              "last_updated": "2025-11-19T13:28:00Z",
              "links": {
                "self": {
                  "href": "/v1/migration-jobs/12"
                }
              },
              "state": "migrating",
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
          "job_number": 3,
          "label": "Labour Market statistics",
          "last_updated": "2025-11-19T13:28:00Z",
          "links": {
            "self": {
              "href": "/v1/migration-jobs/3"
            }
          },
          "state": "migrating",
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
          "job_number": 33,
          "label": "Retail Sales Index",
          "last_updated": "2025-11-20T10:15:00Z",
          "links": {
            "self": {
              "href": "/v1/migration-jobs/33"
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
              "job_number": 33,
              "label": "Retail Sales Index",
              "last_updated": "2025-11-20T10:15:00Z",
              "links": {
                "self": {
                  "href": "/v1/migration-jobs/33"
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
              "job_number": 3,
              "label": "Labour Market statistics",
              "last_updated": "2025-11-19T13:28:00Z",
              "links": {
                "self": {
                  "href": "/v1/migration-jobs/3"
                }
              },
              "state": "migrating",
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
              "job_number": 3,
              "label": "Labour Market statistics",
              "last_updated": "2025-11-19T13:28:00Z",
              "links": {
                "self": {
                  "href": "/v1/migration-jobs/3"
                }
              },
              "state": "migrating",
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
              "job_number": 33,
              "label": "Retail Sales Index",
              "last_updated": "2025-11-20T10:15:00Z",
              "links": {
                "self": {
                  "href": "/v1/migration-jobs/33"
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

    Scenario: Get a list of jobs filtered by a single state
      Given the following document exists in the "jobs" collection:
        """
        {
          "_id": "job-submitted-1",
          "job_number": 23,
          "label": "job-submitted-1",
          "last_updated": "2025-11-19T13:28:00Z",
          "state": "migrating",
          "config": {
            "source_id": "s1",
            "target_id": "t1",
            "type": "type1"
          }
        }
        """
      And the following document exists in the "jobs" collection:
        """
        {
          "_id": "job-approved-1",
          "job_number": 24,
          "label": "job-approved-1",
          "last_updated": "2025-11-19T14:00:00Z",
          "state": "approved",
          "config": {
            "source_id": "s2",
            "target_id": "t2",
            "type": "type2"
          }
        }
        """
      When I GET "/v1/migration-jobs?state=migrating"
      Then I should receive the following JSON response with status "200":
        """
        {
          "count": 1,
          "items": [
            {
              "id": "job-submitted-1",
              "job_number": 23,
              "label": "job-submitted-1",
              "state": "migrating",
              "config": {
                "source_id": "s1",
                "target_id": "t1",
                "type": "type1"
              },
              "last_updated": "2025-11-19T13:28:00Z",
              "links": {}
            }
          ],
          "limit": 10,
          "offset": 0,
          "total_count": 1
        }
        """

    Scenario: Get a list of jobs filtered by multiple states using repeated query param
      Given the following document exists in the "jobs" collection:
        """
        {
          "_id": "job-submitted-1",
          "job_number": 20,
          "label": "job-submitted-1",
          "last_updated": "2025-11-19T13:28:00Z",
          "state": "migrating",
          "config": {
            "source_id": "s1",
            "target_id": "t1",
            "type": "type1"
          }
        }
        """
      And the following document exists in the "jobs" collection:
        """
        {
          "_id": "job-approved-1",
          "job_number": 21,
          "label": "job-approved-1",
          "last_updated": "2025-11-19T14:00:00Z",
          "state": "approved",
          "config": {
            "source_id": "s2",
            "target_id": "t2",
            "type": "type2"
          }
        }
        """
      When I GET "/v1/migration-jobs?state=migrating&state=approved"
      Then I should receive the following JSON response with status "200":
        """
        {
          "count": 2,
          "items": [
            {
              "id": "job-approved-1",
              "job_number": 21,
              "label": "job-approved-1",
              "last_updated": "2025-11-19T14:00:00Z",
              "links": {},
              "state": "approved",
              "config": {
                "source_id": "s2",
                "target_id": "t2",
                "type": "type2"
              }
            },
            {
              "id": "job-submitted-1",
              "job_number": 20,
              "label": "job-submitted-1",
              "last_updated": "2025-11-19T13:28:00Z",
              "links": {},
              "state": "migrating",
              "config": {
                "source_id": "s1",
                "target_id": "t1",
                "type": "type1"
              }
            }
          ],
          "limit": 10,
          "offset": 0,
          "total_count": 2
        }
        """

    Scenario: Get a list of jobs filtered by multiple states using comma-separated values
      Given the following document exists in the "jobs" collection:
        """
        {
          "_id": "job-submitted-1",
          "job_number": 1,
          "label": "job-submitted-1",
          "last_updated": "2025-11-19T13:28:00Z",
          "state": "migrating",
          "config": {
            "source_id": "s1",
            "target_id": "t1",
            "type": "type1"
          }
        }
        """
      And the following document exists in the "jobs" collection:
        """
        {
          "_id": "job-approved-1",
          "job_number": 2,
          "label": "job-approved-1",
          "last_updated": "2025-11-19T14:00:00Z",
          "state": "approved",
          "config": {
            "source_id": "s2",
            "target_id": "t2",
            "type": "type2"
          }
        }
        """
      When I GET "/v1/migration-jobs?state=migrating,approved"
      Then I should receive the following JSON response with status "200":
        """
        {
          "count": 2,
          "items": [
            {
              "id": "job-approved-1",
              "job_number": 2,
              "label": "job-approved-1",
              "last_updated": "2025-11-19T14:00:00Z",
              "links": {},
              "state": "approved",
              "config": {
                "source_id": "s2",
                "target_id": "t2",
                "type": "type2"
              }
            },
            {
              "id": "job-submitted-1",
              "job_number": 1,
              "label": "job-submitted-1",
              "last_updated": "2025-11-19T13:28:00Z",
              "links": {},
              "state": "migrating",
              "config": {
                "source_id": "s1",
                "target_id": "t1",
                "type": "type1"
              }
            }
          ],
          "limit": 10,
          "offset": 0,
          "total_count": 2
        }
        """
    Scenario: Get a list of 2 jobs using valid sort parameters
      Given the following document exists in the "jobs" collection:
        """
        {
          "_id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
          "job_number": 3,
          "label": "Labour Market statistics",
          "last_updated": "2025-11-19T13:28:00Z",
          "links": {
            "self": {
              "href": "/v1/migration-jobs/3"
            }
          },
          "state": "migrating",
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
          "job_number": 33,
          "label": "Retail Sales Index",
          "last_updated": "2025-11-20T10:15:00Z",
          "links": {
            "self": {
              "href": "/v1/migration-jobs/33"
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
      When I GET "/v1/migration-jobs?sort=job_number:desc"
      Then I should receive the following JSON response with status "200":
        """
        {
          "count": 2,
          "items": [
            {
              "id": "4874ee9e-1cec-44f8-9b6d-998cf2062791",
              "job_number": 33,
              "label": "Retail Sales Index",
              "last_updated": "2025-11-20T10:15:00Z",
              "links": {
                "self": {
                  "href": "/v1/migration-jobs/33"
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
              "job_number": 3,
              "label": "Labour Market statistics",
              "last_updated": "2025-11-19T13:28:00Z",
              "links": {
                "self": {
                  "href": "/v1/migration-jobs/3"
                }
              },
              "state": "migrating",
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
      When I GET "/v1/migration-jobs?sort=job_number:asc"
      Then I should receive the following JSON response with status "200":
        """
        {
          "count": 2,
          "items": [
            {
              "id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
              "job_number": 3,
              "label": "Labour Market statistics",
              "last_updated": "2025-11-19T13:28:00Z",
              "links": {
                "self": {
                  "href": "/v1/migration-jobs/3"
                }
              },
              "state": "migrating",
              "config": {
                "source_id": "test-source-id",
                "target_id": "test-target-id",
                "type": "test-type"
              }
            },
            {
              "id": "4874ee9e-1cec-44f8-9b6d-998cf2062791",
              "job_number": 33,
              "label": "Retail Sales Index",
              "last_updated": "2025-11-20T10:15:00Z",
              "links": {
                "self": {
                  "href": "/v1/migration-jobs/33"
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
          "limit": 10,
          "offset": 0,
          "total_count": 2
        }
        """
      When I GET "/v1/migration-jobs?sort=label:asc"
      Then I should receive the following JSON response with status "200":
        """
        {
          "count": 2,
          "items": [
            {
              "id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
              "job_number": 3,
              "label": "Labour Market statistics",
              "last_updated": "2025-11-19T13:28:00Z",
              "links": {
                "self": {
                  "href": "/v1/migration-jobs/3"
                }
              },
              "state": "migrating",
              "config": {
                "source_id": "test-source-id",
                "target_id": "test-target-id",
                "type": "test-type"
              }
            },
            {
              "id": "4874ee9e-1cec-44f8-9b6d-998cf2062791",
              "job_number": 33,
              "label": "Retail Sales Index",
              "last_updated": "2025-11-20T10:15:00Z",
              "links": {
                "self": {
                  "href": "/v1/migration-jobs/33"
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
          "limit": 10,
          "offset": 0,
          "total_count": 2
        }
        """
      When I GET "/v1/migration-jobs?sort=label:desc"
      Then I should receive the following JSON response with status "200":
        """
        {
          "count": 2,
          "items": [
            {
              "id": "4874ee9e-1cec-44f8-9b6d-998cf2062791",
              "job_number": 33,
              "label": "Retail Sales Index",
              "last_updated": "2025-11-20T10:15:00Z",
              "links": {
                "self": {
                  "href": "/v1/migration-jobs/33"
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
              "job_number": 3,
              "label": "Labour Market statistics",
              "last_updated": "2025-11-19T13:28:00Z",
              "links": {
                "self": {
                  "href": "/v1/migration-jobs/3"
                }
              },
              "state": "migrating",
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

    @InvalidInput
    Scenario: Get a list of jobs with an invalid sort field parameter
      When I GET "/v1/migration-jobs?sort=unknown:asc"
      Then I should receive the following JSON response with status "400":
        """
        {
          "errors": [
            {
              "code": 400,
              "description": "field is invalid in sort parameter"
            }
          ]
        }
        """

    @InvalidInput
    Scenario: Get a list of jobs with an invalid sort direction parameter
      When I GET "/v1/migration-jobs?sort=job_number:unknown"
      Then I should receive the following JSON response with status "400":
        """
        {
          "errors": [
            {
              "code": 400,
              "description": "direction is invalid in sort parameter"
            }
          ]
        }
        """

    @InvalidInput
    Scenario: Get a list of jobs with an invalid state
      When I GET "/v1/migration-jobs?state=unknown"
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
