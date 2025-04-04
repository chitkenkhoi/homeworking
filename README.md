# Backend Service Requirement Document

## 1. Introduction

This document defines the requirements for developing a backend service in Golang for a task management system. The primary goal is to train a developer in Go by implementing key functionalities, following best practices, and understanding core backend development concepts.

## 2. Functional Requirements

### 2.1 User Management
- Users should be able to sign up and log in using JWT authentication.
- The system should support different roles: ADMIN, PROJECT_MANAGER, and TEAM_MEMBER.
- Users should be able to update their profile and password.
- Admins should be able to view all users.

**API Endpoints:**
- `POST /users` - Create a new user.
- `POST /login` - Authenticate a user and return a JWT token.
- `GET /users` - Retrieve a list of all users (admin-only access).
- `GET /users/{userId}` - Retrieve a specific user.
- `PUT /users/{userId}` - Update user details.
- `DELETE /users/{userId}` - Soft delete a user.
- `GET /me` - Retrieve details of the currently authenticated user.

### 2.2 Project Management
- Users can create, update, and manage projects.
- Each project should have a name, description, start and end dates, and status.
- A project should have a manager and multiple team members.

**API Endpoints:**
- `POST /projects` - Create a new project.
- `GET /projects` - Retrieve all projects with filtering options.
- `GET /projects/{projectId}` - Get details of a specific project.
- `PUT /projects/{projectId}` - Update project details.
- `DELETE /projects/{projectId}` - Soft delete a project.

### 2.3 Task Management
- Users can create and manage tasks within projects.
- Tasks should support priority levels and statuses.
- A task can be assigned to a user and belong to a project.
- Subtasks should be supported.

**API Endpoints:**
- `POST /tasks` - Create a new task.
- `GET /tasks` - Retrieve all tasks with filtering options.
- `GET /tasks/{taskId}` - Get task details, including subtasks.
- `PUT /tasks/{taskId}` - Update task details.
- `DELETE /tasks/{taskId}` - Soft delete a task.
- `GET /projects/{projectId}/tasks` - Retrieve all tasks for a specific project.
- `GET /users/{userId}/tasks` - Retrieve all tasks assigned to a user.

### 2.4 Sprint Management
- Sprints group tasks into time-boxed iterations.
- Each sprint belongs to a project and contains multiple tasks.

**API Endpoints:**
- `POST /sprints` - Create a new sprint.
- `GET /sprints` - Retrieve all sprints with filtering.
- `GET /sprints/{sprintId}` - Get details of a specific sprint.
- `PUT /sprints/{sprintId}` - Update sprint details.
- `DELETE /sprints/{sprintId}` - Soft delete a sprint.

## 3. Non-Functional Requirements
- The service should use **Go Fiber** for API routing.
- PostgreSQL will be the primary database.
- JWT should be used for authentication.
- Swagger/OpenAPI should be used for API documentation.
- Redis can be used for caching to improve performance.

## 4. Database Schema

### User Table
| Column    | Type   | Description                         |
|-----------|--------|-------------------------------------|
| id        | Int    | Primary key                         |
| username  | String | Unique username                     |
| email     | String | Unique email                        |
| password  | String | Hashed password                     |
| role      | Enum   | ADMIN, PROJECT_MANAGER, TEAM_MEMBER |
| firstName | String | First name                          |
| lastName  | String | Last name                           |

### Project Table
| Column      | Type   | Description                           |
|-------------|--------|---------------------------------------|
| id          | Int    | Primary key                           |
| name        | String | Project name                          |
| description | String | Project description                   |
| startDate   | Time   | Start date                            |
| endDate     | Time   | End date                              |
| status      | Enum   | ACTIVE, COMPLETED, ON_HOLD, CANCELLED |
| managerId   | Int    | Project manager                       |

### Task Table
| Column      | Type   | Description                               |
|-------------|--------|-------------------------------------------|
| id          | Int    | Primary key                               |
| title       | String | Task title                                |
| description | String | Task description                          |
| assigneeId  | Int    | Assigned user                             |
| projectId   | Int    | Associated project                        |
| status      | Enum   | TO_DO, IN_PROGRESS, REVIEW, DONE, BLOCKED |
| priority    | Enum   | HIGH, MEDIUM, LOW, CRITICAL               |
| dueDate     | Time   | Due date                                  |
### Sprint Table  
| Column     | Type   | Description                     |  
|------------|--------|---------------------------------|  
| id         | Int    | Primary key                     |  
| name       | String | Sprint name                     |  
| startDate  | Time   | Sprint start date               |  
| endDate    | Time   | Sprint end date                 |  
| projectId  | Int    | Associated project ID           |  
| goal       | String | Sprint goal                     |
## 5. Developer Understanding Questions
To ensure the developer understands the requirements, consider asking these questions:

1. What data validation strategies will you apply to user input?
2. How would you handle errors and logging in this service?
3. How would you design a database schema to support soft deletes efficiently?
4. What are some secure methods for storing and verifying user passwords ?
5. How would you implement authentication using JWT in Go?
6. What techniques can be used to sign and verify JWT securely?
7. How can you implement role-based access control (RBAC) in this system?
8. What is the benefit of using Redis for caching, and how would you integrate it?
9. How would you write unit tests for the API endpoints?
10. How would you write a Dockerfile with multi-stage builds for optimizing the container size?
11. How would you create a `docker-compose.yml` file to manage dependencies and local development?
12. What techniques would you use to ensure database migrations work smoothly?
13. What are the advantages of using Go Fiber over other Go frameworks like Gin?
14. How can you implement and enforce API rate limiting in this service?
15. What tools and approaches can be used to monitor the performance and health of the backend service?
16. Generic, OpenAPI, Temporal scheduler,
17. "net/http" ServeMux, viết unit test

