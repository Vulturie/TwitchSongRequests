#!/bin/bash

# The purpose of this script is to simplify zipping up the files needed for the lambda deployment.
npm ci && zip -r function.zip index.js node_modules
