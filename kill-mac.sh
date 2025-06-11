#!/bin/bash
# shellcheck disable=SC2046
kill -9 $(lsof -i :5601 | tail -n 1 |  awk '{print $2}' )
kill -9 $(lsof -i :18080 | tail -n 1 |  awk '{print $2}' )
kill -9 $(lsof -i :9200 | tail -n 1 |  awk '{print $2}' )
kill -9 $(lsof -i :8440 | tail -n 1 |  awk '{print $2}' )
