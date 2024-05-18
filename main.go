package main

import (
	"main/TaskAnalysis"
)

func main() {
	const inputDataFilePath string = "Data/data.csv"
	const outputCSVFilePath string = "Data/output-graph.csv"
	const outputJSONFilePath string = "Data/output.json"

	actualDifficulty := 2.0
	guessingProbability := 0.001388889

	analyzer := TaskAnalysis.NewTaskAnalyzer(
		actualDifficulty,
		guessingProbability,
		inputDataFilePath,
	)

	analyzer.WriteGraphDataToCSV(outputCSVFilePath)
	analyzer.WriteAllMethodsDataToJSON(outputJSONFilePath)
}