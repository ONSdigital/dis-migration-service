@Job @GetJobTasks
Feature: Get list of job tasks

  Rule: User that is authorised and authenticated
    Background:
      Given an admin user has the "migrations:read" permission
      And I am an admin user
      And the migration service is running

    Scenario: Get a list of 1 task with defaults for existing job
      Given the following document exists in the "jobs" collection:
        """
        {
          "_id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
          "last_updated": "2025-11-19T13:28:00Z",
          "links": {
            "self": {
              "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
            },
            "events": {
              "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/events"
            },
            "tasks": {
              "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/tasks"
            }
          },
          "state": "submitted",
          "type": "static_dataset",
          "config": {
            "source_id": "test-source-id",
            "target_id": "test-target-id",
            "type": "static_dataset"
          }
        }
        """
      And the following document exists in the "tasks" collection:
        """
        {
          "_id": "task-123e4567-e89b-12d3-a456-426614174000",
          "job_id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
          "last_updated": "2025-11-19T13:30:00Z",
          "state": "migrating",
          "type": "dataset",
          "source": {
            "id": "source-dataset-1",
            "label": "Source Dataset 1",
            "uri": "/economy/inflationandpriceindices/datasets/consumerpriceinflation/data"
          },
          "target": {
            "id": "target-dataset-1",
            "label": "Target Dataset 1",
            "uri": "/data/target/dataset1"
          },
          "links": {
            "self": {
              "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/tasks/task-123e4567-e89b-12d3-a456-426614174000"
            },
            "job": {
              "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
            }
          }
        }
        """
      When I GET "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/tasks"
      Then I should receive the following JSON response with status "200":
        """
        {
          "count": 1,
          "items": [
            {
              "id": "task-123e4567-e89b-12d3-a456-426614174000",
              "job_id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
              "last_updated": "2025-11-19T13:30:00Z",
              "state": "migrating",
              "type": "dataset",
              "source": {
                "id": "source-dataset-1",
                "label": "Source Dataset 1",
                "uri": "/economy/inflationandpriceindices/datasets/consumerpriceinflation/data"
              },
              "target": {
                "id": "target-dataset-1",
                "label": "Target Dataset 1",
                "uri": "/data/target/dataset1"
              },
              "links": {
                "self": {
                  "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/tasks/task-123e4567-e89b-12d3-a456-426614174000"
                },
                "job": {
                  "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
                }
              }
            }
          ],
          "limit": 10,
          "offset": 0,
          "total_count": 1
        }
        """

    Scenario: Get a list of 2 tasks using limit and offset to paginate
      Given the following document exists in the "jobs" collection:
        """
        {
          "_id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
          "last_updated": "2025-11-19T13:28:00Z",
          "links": {
            "self": {
              "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
            },
            "events": {
              "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/events"
            },
            "tasks": {
              "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/tasks"
            }
          },
          "state": "submitted",
          "type": "static_dataset",
          "config": {
            "source_id": "test-source-id",
            "target_id": "test-target-id",
            "type": "static_dataset"
          }
        }
        """
      And the following document exists in the "tasks" collection:
        """
        {
          "_id": "task-123e4567-e89b-12d3-a456-426614174000",
          "job_id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
          "last_updated": "2025-11-19T13:30:00Z",
          "state": "submitted",
          "type": "dataset",
          "source": {
            "id": "source-dataset-1",
            "label": "Source Dataset 1",
            "uri": "/data/source1"
          },
          "target": {
            "id": "target-dataset-1",
            "label": "Target Dataset 1",
            "uri": "/data/target1"
          },
          "links": {
            "self": {
              "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/tasks/task-123e4567-e89b-12d3-a456-426614174000"
            },
            "job": {
              "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
            }
          }
        }
        """
      And the following document exists in the "tasks" collection:
        """
        {
          "_id": "task-456e7890-e89b-12d3-a456-426614174001",
          "job_id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
          "last_updated": "2025-11-19T13:35:00Z",
          "state": "publishing",
          "type": "dataset_edition",
          "source": {
            "id": "source-edition-2",
            "label": "Source Edition 2",
            "uri": "/data/source2"
          },
          "target": {
            "id": "target-edition-2",
            "label": "Target Edition 2",
            "uri": "/data/target2"
          },
          "links": {
            "self": {
              "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/tasks/task-456e7890-e89b-12d3-a456-426614174001"
            },
            "job": {
              "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
            }
          }
        }
        """
      When I GET "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/tasks"
      Then I should receive the following JSON response with status "200":
        """
        {
          "count": 2,
          "items": [
            {
              "id": "task-456e7890-e89b-12d3-a456-426614174001",
              "job_id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
              "last_updated": "2025-11-19T13:35:00Z",
              "state": "publishing",
              "type": "dataset_edition",
              "source": {
                "id": "source-edition-2",
                "label": "Source Edition 2",
                "uri": "/data/source2"
              },
              "target": {
                "id": "target-edition-2",
                "label": "Target Edition 2",
                "uri": "/data/target2"
              },
              "links": {
                "self": {
                  "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/tasks/task-456e7890-e89b-12d3-a456-426614174001"
                },
                "job": {
                  "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
                }
              }
            },
            {
              "id": "task-123e4567-e89b-12d3-a456-426614174000",
              "job_id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
              "last_updated": "2025-11-19T13:30:00Z",
              "state": "submitted",
              "type": "dataset",
              "source": {
                "id": "source-dataset-1",
                "label": "Source Dataset 1",
                "uri": "/data/source1"
              },
              "target": {
                "id": "target-dataset-1",
                "label": "Target Dataset 1",
                "uri": "/data/target1"
              },
              "links": {
                "self": {
                  "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/tasks/task-123e4567-e89b-12d3-a456-426614174000"
                },
                "job": {
                  "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
                }
              }
            }
          ],
          "limit": 10,
          "offset": 0,
          "total_count": 2
        }
        """
      When I GET "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/tasks?limit=1&offset=1"
      Then I should receive the following JSON response with status "200":
        """
        {
          "count": 1,
          "items": [
            {
              "id": "task-123e4567-e89b-12d3-a456-426614174000",
              "job_id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
              "last_updated": "2025-11-19T13:30:00Z",
              "state": "submitted",
              "type": "dataset",
              "source": {
                "id": "source-dataset-1",
                "label": "Source Dataset 1",
                "uri": "/data/source1"
              },
              "target": {
                "id": "target-dataset-1",
                "label": "Target Dataset 1",
                "uri": "/data/target1"
              },
              "links": {
                "self": {
                  "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/tasks/task-123e4567-e89b-12d3-a456-426614174000"
                },
                "job": {
                  "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
                }
              }
            }
          ],
          "limit": 1,
          "offset": 1,
          "total_count": 2
        }
        """

    Scenario: Get a list of tasks when no tasks are available for existing job
      Given the following document exists in the "jobs" collection:
        """
        {
          "_id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
          "last_updated": "2025-11-19T13:28:00Z",
          "links": {
            "self": {
              "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
            },
            "events": {
              "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/events"
            },
            "tasks": {
              "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/tasks"
            }
          },
          "state": "submitted",
          "type": "static_dataset",
          "config": {
            "source_id": "test-source-id",
            "target_id": "test-target-id",
            "type": "static_dataset"
          }
        }
        """
      When I GET "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/tasks"
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

    Scenario: Get a list of tasks for non-existent job returns 404
      When I GET "/v1/migration-jobs/non-existent-job-id/tasks"
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

    @InvalidInput
    Scenario: Get a list of tasks with limit exceeding maximum allowed
      Given the following document exists in the "jobs" collection:
        """
        {
          "_id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
          "last_updated": "2025-11-19T13:28:00Z",
          "links": {
            "self": {
              "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
            },
            "events": {
              "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/events"
            },
            "tasks": {
              "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/tasks"
            }
          },
          "state": "submitted",
          "type": "static_dataset",
          "config": {
            "source_id": "test-source-id",
            "target_id": "test-target-id",
            "type": "static_dataset"
          }
        }
        """
      When I GET "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/tasks?limit=2000"
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
    Scenario: Get a list of tasks with offset and limit invalid
      Given the following document exists in the "jobs" collection:
        """
        {
          "_id": "2874ee9e-1cec-44f8-9b6d-998cf2062791",
          "last_updated": "2025-11-19T13:28:00Z",
          "links": {
            "self": {
              "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791"
            },
            "events": {
              "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/events"
            },
            "tasks": {
              "href": "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/tasks"
            }
          },
          "state": "submitted",
          "type": "static_dataset",
          "config": {
            "source_id": "test-source-id",
            "target_id": "test-target-id",
            "type": "static_dataset"
          }
        }
        """
      When I GET "/v1/migration-jobs/2874ee9e-1cec-44f8-9b6d-998cf2062791/tasks?limit=invalid&offset=-10"
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

  @Auth
  Rule: Users that are not authorised or authenticated
    Background:
      Given an admin user has the "incorrect" permission
      And the migration service is running

    Scenario: User that is not authenticated
      Given I am not authorised
      And the migration service is running
      When I GET "/v1/migration-jobs/1"
      Then the HTTP status code should be "401"

    Scenario: User that is not authorised
      Given I am an admin user
      When I GET "/v1/migration-jobs/1"
      Then the HTTP status code should be "403"