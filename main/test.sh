#!/bin/bash

source ../douji.env #only needed when using LeanCloud db so ok to ignore errors when testing locally.
go build
exec ./main