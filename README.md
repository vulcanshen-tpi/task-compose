# Task-Compose

`task-compose` is a convenient command-line utility built with Go and Cobra. It's designed to orchestrate and execute a series of commands based on a declarative YAML configuration file, similar to how container orchestrators manage services. This tool simplifies the management of complex multi-command setups, making it ideal for local development environments, testing suites, or task automation.

With `task-compose`, you define your tasks, their execution parameters, and their dependencies, allowing you to bring up and manage an entire set of related processes with a single command. It includes robust health check capabilities, ensuring that each task is healthy before dependent tasks are started.

### Features

- Declarative Configuration: Define all your commands and their settings in a human-readable YAML file.
- Process Orchestration: Run multiple commands concurrently or sequentially based on defined dependencies.
- Comprehensive Health Checks: Ensure your tasks are healthy before proceeding.
    - HTTP Health Checks: Verify service availability via HTTP GET requests.
      - JSON Path Validation: Extract and validate specific values from JSON HTTP responses.
    - Command Health Checks: Use custom shell commands to determine task health.
- Dependency Management: Specify task dependencies to control the startup order.
- Flexible Execution: Set base_dir, executable, args, and envs for each task.

### How To Use
To use task-compose, you'll typically place your configuration in a YAML file (e.g., task-compose.yaml).

### Basic Command

To execute tasks defined in your configuration file:

```bash
task-compose -f task-compose.yaml up
```

### Command Line Interface (CLI)

`task-compose` provides a straightforward command-line interface.

**Usage:**

```bash
task-compose [command]
```

**Available Commands:**

| Command    | Descriptions                                                |
|:-----------|:------------------------------------------------------------|
 | check      | Confirm the correctness of the YAML content format.         |
 | completion | Generate the autocompletion script for the specified shell. |
 | down       | Kill previous tasks processes.                              |
 | help       | Help about any command.                                     |
 | up         | Execute tasks according to the YAML configuration file.     |
 | version    | Show version number and build details of task-compose.      |


### Example Configuration (task-compose.yaml demo)

Here's an example of how you might configure `task-compose` to start Elasticsearch and Kibana, with proper health checks and dependencies:


```yaml
tasks:
  - name: elasticsearch
    base_dir: ../elk/elasticsearch
    executable: bin/elasticsearch
    args:
      - -E
      - xpack.security.enabled=false
      - -E
      - xpack.security.http.ssl.enabled=false
      - -E
      - xpack.security.transport.ssl.enabled=false
      - -E
      - xpack.monitoring.collection.enabled=true
    healthcheck:
      frequency:
        interval: 5s
        timeout: 10s
        retries: 5
        delay: 5s
      http:
        url: http://localhost:9200
  - name: kibana
    base_dir: ../elk
    executable: kibana/bin/kibana
    args:
      - -c
      - kibana.yml
    healthcheck:
      frequency:
        interval: 5s
        timeout: 10s
        retries: 5
        delay: 5s
      http:
        url: http://localhost:5601/api/status
        expect:
          json:
            value: available
            jsonpath: "$.status.overall.level"
    depends_on:
      - elasticsearch
  - name: curl1
    executable: curl
    args:
      - -v
      - http://localhost:9200
    depends_on:
      - elasticsearch
  - name: curl2
    executable: curl
    args:
      - -v
      - http://localhost:5601/api/status
    depends_on:
      - kibana
```

### Configuration File Reference

The `tasks` key at the root of your YAML file contains a list of individual task definitions. Each task can have the following properties:

- `name` (string, required): A unique identifier for the task.
- `base_dir` (string): The working directory for the command. cmd.Dir will be set to this path. If not specified, the current working directory of task-compose will be used.
- `executable` (string, required): The path to the executable command (e.g., node, java, ./my-app).
- `args` ([]string): A list of arguments to pass to the executable.
- `envs` ([]string): A list of environment variables to set for the command (e.g., KEY=VALUE). These are merged with the parent process's environment variables.
- `depends_on` ([]string): A list of task names that this task depends on. This task will only start after all its dependencies have successfully passed their health checks.
- `healthcheck` (object): Defines how task-compose determines if a task is healthy.
  - `healthcheck.http` (object): Configures an HTTP GET health check.
    - `healthcheck.http.url` (string, required): The URL to send the HTTP GET request to.
    - `healthcheck.http.expect` (object, optional): Defines expected responses.
    If not set, a 2xx HTTP status code indicates health.
      - `healthcheck.http.expect.json` (object): Expects a JSON response.
        - `healthcheck.http.expect.json.jsonpath` (string, required): A JSONPath expression to extract a value from the response.
        - `healthcheck.http.expect.json.value` (string, required): The expected value (as a string) to match against the extracted JSONPath value.
    - `healthcheck.command` (object): Configures a command-based health check.
    - `healthcheck.command.scripts` ([]string, required): A list where the first element is the command, and subsequent elements are its arguments. The command is considered healthy if it exits with a zero status code.
    - `healthcheck.frequency` (object): Controls the timing of health checks.
      - `healthcheck.frequency.interval` (duration string): The time between consecutive health check attempts (e.g., 5s, 1m).
      - `healthcheck.frequency.timeout` (duration string): The maximum time allowed for a single health check attempt (e.g., 10s).
      - `healthcheck.frequency.retries` (int): The maximum number of consecutive failed health checks before the task is considered unhealthy.
      - `healthcheck.frequency.delay` (duration string): The initial delay before the first health check attempt is made after a task starts (e.g., 5s).