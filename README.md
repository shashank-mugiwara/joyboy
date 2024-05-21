# joyboy
[NOTE] The project is currently in active development, but still can be used for simple use-cases.

Open-source Amazon ECS alternative to run on a single machine for personal hosting/serving.

### To run:
1. Clone the repository and checkout to **main** branch.
2. go build
3. ./joyboy

### Adding and running a task
joyboy exposes REST based endpoints to add/update the task configurations. New changes in the config are detected every 5 seconds.

Example to add a task:
```sh
curl '{server-url}:8070/api/v1/task/stop' \
--header 'Content-Type: application/json' \
--data '{
    "name": "nginx-apac-001",
    "image": "nginx:stable-alpine3.17-slim",
    "portMapping": {
        "80":"8211"
    },
    "resources": {
        "memory": 512,
        "cpus": 1.5
    }
}'
```

| **KEY**  |  **DESCRIPTION** |
|---|---|
| name  | name of the docker image, for internal use, could be any string.  |
| image  | offical docker image name with tag  |
|  portMapping | port mapping  |
|  resources.cpus | cpu resources for the tasks to run in cores (floating points also allowed)  |
|  resources.memory | memory for running the task  |


To fetch all the running tasks information you can use:
```sh
curl '{server-url}:8070/api/v1/task/tasks'
```