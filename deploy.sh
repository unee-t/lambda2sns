#!/bin/bash
#This script is created to deploy lambda2sqs 
#Variable is the environment like dev, demo or prod
#To run this script, run this command: ./deploy.sh [STAGE] where STAGE is dev, demo or prod
#
#Step 1: Setup the parameters
export AWS_PROFILE=ins-dev
source aws-env.$1

#Step 2: Run Makefile.
make $1
