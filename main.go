package main

import (
	"main/TaskAnalysis"
)

func main() {

	// Констатные пути-названия для файлов ввода-вывода
	const inputDataFilePath string = "Data/data.csv"
	const outputCSVFilePath string = "Data/output-graph.csv"
	const outputJSONFilePath string = "Data/output.json"

	// До увязки с БД фактическую трудность и вероятность угадывания пишем сюда
	actualDifficulty := 2.0
	guessingProbability := 0.001388889

	// Инициализация объекта анализатора
	analyzer := TaskAnalysis.NewTaskAnalyzer(
		actualDifficulty,
		guessingProbability,
		inputDataFilePath,
	)

	// Рожаем данные в CSV (только точки графика) и JSON (вообще всё)
	analyzer.WriteGraphDataToCSV(outputCSVFilePath)
	analyzer.WriteAllMethodsDataToJSON(outputJSONFilePath)

}
