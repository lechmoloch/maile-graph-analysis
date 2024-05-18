package main

import (
	"main/TaskAnalysis"
)

func main() {
	const inputFilePath string = "Data/data.csv"
	const outputFilePath string = "Data/output.csv"

	actualDifficulty := 2.0
	guessingProbability := 0.001388889

	analyzer := TaskAnalysis.NewTaskAnalyzer(
		actualDifficulty,
		guessingProbability,
		inputFilePath,
	)

	analyzer.WriteGraphData(outputFilePath)
}