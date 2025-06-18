package cmd

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "",
	Short: "UniAI is a CLI client for interacting with UniAI models.",
	Long: `UniAI is a command-line interface (CLI) client designed to interact with UniAI models, 
providing functionalities such as pdf to text generation, document QA, and make structured data.`,
}

func Execute() {
	err := godotenv.Load() // by default loads ".env"
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
