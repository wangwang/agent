#!/bin/ruby
# rubocop:disable all

require_relative '../../e2e'

start_job <<-JSON
  {
    "id": "#{$JOB_ID}",

    "env_vars": [],

    "files": [],

    "commands": [
      { "directive": "echo Hello World", "alias": "Display Hello World" }
    ],

    "epilogue_always_commands": [],

    "callbacks": {
      "finished": "https://httpbin.org/status/200",
      "teardown_finished": "https://httpbin.org/status/200"
    }
  }
JSON

wait_for_job_to_finish

assert_job_log <<-LOG
  {"event":"job_started",  "timestamp":"*"}
  {"event":"cmd_started",  "timestamp":"*", "directive":"Exporting environment variables"}
  {"event":"cmd_finished", "timestamp":"*", "directive":"Exporting environment variables","exit_code":0,"finished_at":"*","started_at":"*"}
  {"event":"cmd_started",  "timestamp":"*", "directive":"Injecting Files"}
  {"event":"cmd_finished", "timestamp":"*", "directive":"Injecting Files","exit_code":0,"finished_at":"*","started_at":"*"}
  {"event":"cmd_started",  "timestamp":"*", "directive":"Display Hello World"}
  {"event":"cmd_output",   "timestamp":"*", "output":"Running: echo Hello World\\n"}
  {"event":"cmd_output",   "timestamp":"*", "output":"Hello World\\n"}
  {"event":"cmd_finished", "timestamp":"*", "directive":"Display Hello World","exit_code":0,"finished_at":"*","started_at":"*"}
  {"event":"cmd_started",  "timestamp":"*", "directive":"export SEMAPHORE_JOB_RESULT=passed"}
  {"event":"cmd_finished", "timestamp":"*", "directive":"export SEMAPHORE_JOB_RESULT=passed","exit_code":0,"finished_at":"*","started_at":"*"}
  {"event":"job_finished", "timestamp":"*", "result":"passed"}
LOG
