package main

// Register as loader.
import (
	_ "github.com/tukaelu/ikesu/internal/config/loader/file"
	_ "github.com/tukaelu/ikesu/internal/config/loader/s3"
)
