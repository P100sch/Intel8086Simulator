package tests

import "embed"

//go:embed data/*
var testFiles embed.FS
