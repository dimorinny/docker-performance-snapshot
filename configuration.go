package main

type Configuration struct {
	ContainerID     string `env:"CONTAINER,required"`
	ResultDirectory string `env:"RESULT_DIRECTORY,required"`
}
